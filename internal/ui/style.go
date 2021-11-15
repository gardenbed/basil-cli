package ui

import (
	"fmt"
	"strconv"
	"strings"
)

type ANSICode int

const (
	Reset ANSICode = iota
	Bold
	Faint
	Italic
	Underline
	BlinkSlow
	BlinkRapid
	ReverseVideo
	Concealed
	CrossedOut
)

const (
	FgBlack ANSICode = iota + 30
	FgRed
	FgGreen
	FgYellow
	FgBlue
	FgMagenta
	FgCyan
	FgWhite
)

const (
	BgBlack ANSICode = iota + 40
	BgRed
	BgGreen
	BgYellow
	BgBlue
	BgMagenta
	BgCyan
	BgWhite
)

type Style []ANSICode

var (
	Black   = Style{FgBlack}
	Red     = Style{FgRed}
	Green   = Style{FgGreen}
	Yellow  = Style{FgYellow}
	Blue    = Style{FgBlue}
	Magenta = Style{FgMagenta}
	Cyan    = Style{FgCyan}
	White   = Style{FgWhite}
)

func (s Style) sprintf(format string, a ...interface{}) string {
	const escape = "\x1b"

	codes := make([]string, len(s))
	for i, v := range s {
		codes[i] = strconv.Itoa(int(v))
	}
	sequence := strings.Join(codes, ";")

	ansiFormat := fmt.Sprintf("%s[%sm", escape, sequence)
	ansiReset := fmt.Sprintf("%s[%dm", escape, Reset)

	return ansiFormat + fmt.Sprintf(format, a...) + ansiReset
}
