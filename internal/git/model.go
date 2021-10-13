package git

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Signature determines who and when created a commit or tag.
type Signature struct {
	Name  string
	Email string
	Time  time.Time
}

// Before determines if a given signature is chronologically before another signature.
func (s Signature) Before(t Signature) bool {
	return s.Time.Before(t.Time)
}

// After determines if a given signature is chronologically after another signature.
func (s Signature) After(t Signature) bool {
	return s.Time.After(t.Time)
}

func (s Signature) String() string {
	return fmt.Sprintf("%s <%s> %s", s.Name, s.Email, s.Time.Format(time.RFC3339))
}

// Commit represents a Git commit.
type Commit struct {
	Hash      string
	Author    Signature
	Committer Signature
	Message   string
	Parents   []string
}

func toCommit(c *object.Commit) Commit {
	return Commit{
		Hash: c.Hash.String(),
		Author: Signature{
			Name:  c.Author.Name,
			Email: c.Author.Email,
			Time:  c.Author.When,
		},
		Committer: Signature{
			Name:  c.Committer.Name,
			Email: c.Committer.Email,
			Time:  c.Committer.When,
		},
		Message: c.Message,
	}
}

// Equal determines if two commits are the same.
// Two commits are the same if they both have the same hash.
func (c Commit) Equal(d Commit) bool {
	return c.Hash == d.Hash
}

// Before determines if a given commit is chronologically before another commit.
func (c Commit) Before(d Commit) bool {
	return c.Committer.Before(d.Committer)
}

// After determines if a given commit is chronologically after another commit.
func (c Commit) After(d Commit) bool {
	return c.Committer.After(d.Committer)
}

// ShortMessage returns a one-line truncated commit message.
func (c Commit) ShortMessage() string {
	message := strings.Split(c.Message, "\n")[0]
	if len(message) > 100 {
		message = message[:100] + " ..."
	}
	return message
}

func (c Commit) String() string {
	return fmt.Sprintf("%s %s", c.Hash[:7], c.ShortMessage())
}

// Text returns a multi-line commit string.
func (c Commit) Text() string {
	return fmt.Sprintf("%s\nAuthor:    %s\nCommitter: %s\n%s", c.Hash, c.Author, c.Committer, c.Message)
}

// Commits is a map of Git commits.
type Commits []Commit

// TagType determines type a Git tag.
type TagType int

const (
	// Lightweight is a lightweight Git tag.
	Lightweight TagType = iota
	// Annotated is an annotated Git tag.
	Annotated
)

func (t TagType) String() string {
	switch t {
	case Lightweight:
		return "Lightweight"
	case Annotated:
		return "Annotated"
	default:
		return "Invalid"
	}
}

// Tag represents a Git tag.
type Tag struct {
	Type    TagType
	Hash    string
	Name    string
	Tagger  *Signature
	Message *string
	Commit  Commit
}

func toLightweightTag(r *plumbing.Reference, c *object.Commit) Tag {
	// It is assumed that the given reference is a tag reference
	name := strings.TrimPrefix(string(r.Name()), "refs/tags/")

	return Tag{
		Type:   Lightweight,
		Hash:   r.Hash().String(),
		Name:   name,
		Commit: toCommit(c),
	}
}

func toAnnotatedTag(t *object.Tag, c *object.Commit) Tag {
	return Tag{
		Type: Annotated,
		Hash: t.Hash.String(),
		Name: t.Name,
		Tagger: &Signature{
			Name:  t.Tagger.Name,
			Email: t.Tagger.Email,
			Time:  t.Tagger.When,
		},
		Message: &t.Message,
		Commit:  toCommit(c),
	}
}

// IsZero determines if a tag is a zero tag instance.
func (t Tag) IsZero() bool {
	return reflect.ValueOf(t).IsZero()
}

// Equal determines if two tags are the same.
// Two tags are the same if they both have the same name.
func (t Tag) Equal(u Tag) bool {
	return t.Name == u.Name
}

// Before determines if a given tag is chronologically before another tag.
// Two tags are compared using the commits they refer to.
func (t Tag) Before(u Tag) bool {
	return t.Commit.Before(u.Commit)
}

// After determines if a given tag is chronologically after another tag.
// Two tags are compared using the commits they refer to.
func (t Tag) After(u Tag) bool {
	return t.Commit.After(u.Commit)
}

func (t Tag) String() string {
	if t.IsZero() {
		return ""
	}
	return fmt.Sprintf("%s %s %s Commit[%s %s]", t.Type, t.Hash, t.Name, t.Commit.Hash, t.Commit.ShortMessage())
}

// Tags is a list of Git tags.
type Tags []Tag

// First returns the first tag that satisifies the given predicate.
// If you pass a nil function, the first tag will be returned.
func (t Tags) First(f func(Tag) bool) (Tag, bool) {
	if f == nil {
		if len(t) > 0 {
			return t[0], true
		}
		return Tag{}, false
	}

	for _, tag := range t {
		if f(tag) {
			return tag, true
		}
	}

	return Tag{}, false
}

// Last returns the last tag that satisifies the given predicate.
// If you pass a nil function, the last tag will be returned.
func (t Tags) Last(f func(Tag) bool) (Tag, bool) {
	if f == nil {
		if l := len(t); l > 0 {
			return t[l-1], true
		}
		return Tag{}, false
	}

	for i := len(t) - 1; i >= 0; i-- {
		if f(t[i]) {
			return t[i], true
		}
	}

	return Tag{}, false
}

// Select partitions a list of tags by a given predicate.
// The first return value is the collection of selected tags (satisfying the predicate).
// The second return value is the collection of unselected tags (not satisfying the predicate).
func (t Tags) Select(f func(Tag) bool) (Tags, Tags) {
	selected := Tags{}
	unselected := Tags{}

	for _, tag := range t {
		if f(tag) {
			selected = append(selected, tag)
		} else {
			unselected = append(unselected, tag)
		}
	}

	return selected, unselected
}
