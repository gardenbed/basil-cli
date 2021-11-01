package compile

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	goast "go/ast"
	goparser "go/parser"
	gotoken "go/token"

	"github.com/gardenbed/basil-cli/internal/log"
)

// PackageInfo contains information about a parsed package.
type PackageInfo struct {
	ModuleName  string
	PackageName string
	ImportPath  string
	BaseDir     string
	RelativeDir string
}

// FileInfo contains information about a parsed file.
type FileInfo struct {
	PackageInfo
	FileName string
	FileSet  *gotoken.FileSet
}

// TypeInfo contains information about a parsed type.
type TypeInfo struct {
	FileInfo
	TypeName string
}

// IsExported determines whether or not a type is exported.
func (i *TypeInfo) IsExported() bool {
	return i.TypeName == strings.Title(i.TypeName)
}

// FuncInfo contains information about a parsed function.
type FuncInfo struct {
	FileInfo
	FuncName string
	RecvName string
	RecvType goast.Expr
}

// IsExported determines whether or not a function is exported.
func (i *FuncInfo) IsExported() bool {
	return i.FuncName == strings.Title(i.FuncName)
}

// IsMethod determines if a function is a method of a struct.
func (i *FuncInfo) IsMethod() bool {
	return i.RecvName != "" && i.RecvType != nil
}

// Consumer is used for processing AST nodes.
// This is meant to be provided by downstream packages.
type Consumer struct {
	Name      string
	Package   func(*PackageInfo, *goast.Package) bool
	FilePre   func(*FileInfo, *goast.File) bool
	Import    func(*FileInfo, *goast.ImportSpec)
	Struct    func(*TypeInfo, *goast.StructType)
	Interface func(*TypeInfo, *goast.InterfaceType)
	FuncType  func(*TypeInfo, *goast.FuncType)
	FuncDecl  func(*FuncInfo, *goast.FuncType, *goast.BlockStmt)
	FilePost  func(*FileInfo, *goast.File) error
}

// ParseOptions configure how Go source code files should be parsed.
type ParseOptions struct {
	MergePackageFiles bool
	SkipTestFiles     bool
	TypeNames         []string
	TypeRegexp        *regexp.Regexp
}

func (o ParseOptions) matchType(name *goast.Ident) bool {
	// If no filter specified, it is a match
	if len(o.TypeNames) == 0 && o.TypeRegexp == nil {
		return true
	}

	for _, t := range o.TypeNames {
		if name.Name == t {
			return true
		}
	}

	if o.TypeRegexp != nil && o.TypeRegexp.MatchString(name.Name) {
		return true
	}

	return false
}

// Parser is used for parsing Go source code files.
type parser struct {
	logger    log.Logger
	consumers []*Consumer
}

// Parse parses all Go source code files recursively from a given path.
func (p *parser) Parse(path string, opts ParseOptions) error {
	// Sanitize the path
	if _, err := os.Stat(path); err != nil {
		return err
	}

	p.logger.Infof("Parsing ...")

	module, err := getModuleName(path)
	if err != nil {
		return err
	}

	// Create a new file set for each package
	fset := gotoken.NewFileSet()

	return readPackages(path, func(baseDir, relDir string) error {
		pkgDir := filepath.Join(baseDir, relDir)
		importPath := filepath.Join(module, relDir)

		// Parse all Go packages and files in the currecnt directory
		p.logger.Debugf("  Parsing directory: %s", pkgDir)
		pkgs, err := goparser.ParseDir(fset, pkgDir, nil, goparser.AllErrors)
		if err != nil {
			return err
		}

		// Visit all parsed Go files in the current directory
		for pkgName, pkg := range pkgs {
			p.logger.Debugf("    Package: %s", pkg.Name)

			pkgInfo := PackageInfo{
				ModuleName:  module,
				PackageName: pkgName,
				ImportPath:  importPath,
				BaseDir:     baseDir,
				RelativeDir: relDir,
			}

			// Keeps track of interested consumers in the files in the current package
			fileConsumers := make([]*Consumer, 0)

			// PACKAGE
			for _, c := range p.consumers {
				if c.Package != nil {
					cont := c.Package(&pkgInfo, pkg)
					if cont {
						fileConsumers = append(fileConsumers, c)
					}
					p.logger.Debugf("      %s.Package: %t", c.Name, cont)
				}
			}

			// Proceed to the next package if no consumer
			if len(fileConsumers) == 0 {
				continue
			}

			// Merge all file ASTs in the package and process a single file
			if opts.MergePackageFiles {
				mergedFile := goast.MergePackageFiles(pkg, goast.FilterImportDuplicates|goast.FilterUnassociatedComments)
				if err := p.processFile(pkgInfo, fset, "merged.go", mergedFile, fileConsumers, opts); err != nil {
					return err
				}
			} else {
				for fileName, file := range pkg.Files {
					if opts.SkipTestFiles && strings.HasSuffix(fileName, "_test.go") {
						continue
					}

					if err := p.processFile(pkgInfo, fset, fileName, file, fileConsumers, opts); err != nil {
						return err
					}
				}
			}
		}

		return nil
	})
}

func (p *parser) processFile(pkgInfo PackageInfo, fset *gotoken.FileSet, fileName string, file *goast.File, fileConsumers []*Consumer, opts ParseOptions) error {
	p.logger.Debugf("      File: %s", fileName)

	fileInfo := FileInfo{
		PackageInfo: pkgInfo,
		FileName:    filepath.Base(fileName),
		FileSet:     fset,
	}

	// Keeps track of interested consumers in the declarations in the current file
	declConsumers := make([]*Consumer, 0)

	// FILE (pre)
	for _, c := range fileConsumers {
		if c.FilePre != nil {
			cont := c.FilePre(&fileInfo, file)
			if cont {
				declConsumers = append(declConsumers, c)
			}
			p.logger.Debugf("        %s.FilePre: %t", c.Name, cont)
		}
	}

	// Proceed to the next file if no consumer
	if len(declConsumers) == 0 {
		return nil
	}

	goast.Inspect(file, func(n goast.Node) bool {
		switch v := n.(type) {
		// IMPORT
		case *goast.ImportSpec:
			p.logger.Debugf("          ImportSpec: %s", v.Path.Value)
			for _, c := range declConsumers {
				if c.Import != nil {
					c.Import(&fileInfo, v)
					p.logger.Debugf("            %s.Import", c.Name)
				}
			}
			return false

		// Handle Types
		case *goast.TypeSpec:
			typeInfo := TypeInfo{
				FileInfo: fileInfo,
				TypeName: v.Name.Name,
			}

			switch w := v.Type.(type) {
			// STRUCT
			case *goast.StructType:
				p.logger.Debugf("          StructType: %s", v.Name.Name)
				for _, c := range declConsumers {
					if c.Struct != nil {
						if opts.matchType(v.Name) {
							c.Struct(&typeInfo, w)
							p.logger.Debugf("            %s.Struct", c.Name)
						}
					}
				}
				return false

			// INTERFACE
			case *goast.InterfaceType:
				p.logger.Debugf("          InterfaceType: %s", v.Name.Name)
				for _, c := range declConsumers {
					if c.Interface != nil {
						if opts.matchType(v.Name) {
							c.Interface(&typeInfo, w)
							p.logger.Debugf("            %s.Interface", c.Name)
						}
					}
				}
				return false

			// FUNCTION (type)
			case *goast.FuncType:
				p.logger.Debugf("          FuncType: %s", v.Name.Name)
				for _, c := range declConsumers {
					if c.FuncType != nil {
						if opts.matchType(v.Name) {
							c.FuncType(&typeInfo, w)
							p.logger.Debugf("            %s.FuncType", c.Name)
						}
					}
				}
				return false
			}

		// FUNCTION (declaration)
		case *goast.FuncDecl:
			p.logger.Debugf("          FuncDecl: %s", v.Name.Name)

			funcInfo := FuncInfo{
				FileInfo: fileInfo,
				FuncName: v.Name.Name,
			}

			if v.Recv != nil && len(v.Recv.List) == 1 {
				if len(v.Recv.List[0].Names) == 1 {
					funcInfo.RecvName = v.Recv.List[0].Names[0].Name
				}
				funcInfo.RecvType = v.Recv.List[0].Type
			}

			for _, c := range declConsumers {
				if c.FuncDecl != nil {
					c.FuncDecl(&funcInfo, v.Type, v.Body)
					p.logger.Debugf("            %s.FuncDecl", c.Name)
				}
			}

			return false
		}

		return true
	})

	// FILE (post)
	for _, c := range declConsumers {
		if c.FilePost != nil {
			err := c.FilePost(&fileInfo, file)
			if err != nil {
				return err
			}
			p.logger.Debugf("        %s.FilePost", c.Name)
		}
	}

	return nil
}
