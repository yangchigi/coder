package cliui

import (
	"os"

	"github.com/muesli/termenv"
	"golang.org/x/xerrors"

	"github.com/coder/pretty"
)

var Canceled = xerrors.New("canceled")

// DefaultStyles compose visual elements of the UI.
var DefaultStyles Styles

type Styles struct {
	Code,
	DateTimeStamp,
	Error,
	Field,
	Keyword,
	Placeholder,
	Prompt,
	FocusedPrompt,
	Fuchsia,
	Warn,
	Wrap pretty.Style
}

var color = termenv.NewOutput(os.Stdout).ColorProfile()

var (
	Green   = color.Color("#04B575")
	Red     = color.Color("#ED567A")
	Fuschia = color.Color("#EE6FF8")
	Yellow  = color.Color("#ECFD65")
)

func ifTerm(f pretty.Formatter) pretty.Formatter {
	if color != termenv.Ascii {
		return f
	}
	return pretty.Nop
}

// Bold returns a formatter that renders text in bold
// if the terminal supports it.
func Bold() pretty.Formatter {
	return ifTerm(Bold())
}

func init() {
	// We do not adapt the color based on whether the terminal is light or dark.
	// Doing so would require a round-trip between the program and the terminal
	// due to the OSC query and response.
	DefaultStyles = Styles{
		Code: pretty.Style{
			pretty.XPad(1, 1),
			pretty.FgColor(Red),
		},
		DateTimeStamp: pretty.Style{
			pretty.FgColor(color.Color("#7571F9")),
		},
		Error: pretty.Style{
			pretty.FgColor(Red),
		},
		Field: pretty.Style{
			pretty.XPad(1, 1),
			pretty.FgColor(color.Color("#FFFFFF")),
			pretty.BgColor(color.Color("#2b2a2a")),
		},
		Keyword: pretty.Style{
			pretty.FgColor(Green),
		},
		Placeholder: pretty.Style{
			pretty.FgColor(color.Color("#4d46b3")),
		},
		Prompt: pretty.Style{
			pretty.FgColor(color.Color("#5C5C5C")),
			pretty.Wrap(">", ""),
		},
		Warn: pretty.Style{
			pretty.FgColor(Yellow),
		},
		Wrap: pretty.Style{
			pretty.LineWrap(80),
		},
	}

	DefaultStyles.FocusedPrompt = append(
		DefaultStyles.Prompt,
		pretty.FgColor(Fuschia),
	)
}

// ValidateNotEmpty is a helper function to disallow empty inputs!
func ValidateNotEmpty(s string) error {
	if s == "" {
		return xerrors.New("Must be provided!")
	}
	return nil
}
