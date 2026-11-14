// Package quick provides simple, no-configuration source code highlighting.
package quick

import (
	"io"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// Highlight some text.
//
// Lexer, formatter and style may be empty, in which case a best-effort is made.
func Highlight(w io.Writer, source, lexer, formatter, style string) error {
	// Determine lexer.
	selectedLexer := lexers.Get(lexer)
	if selectedLexer == nil {
		selectedLexer = lexers.Analyse(source)
	}
	if selectedLexer == nil {
		selectedLexer = lexers.Fallback
	}
	selectedLexer = chroma.Coalesce(selectedLexer)

	// Determine formatter.
	selectedFormatter := formatters.Get(formatter)
	if selectedFormatter == nil {
		selectedFormatter = formatters.Fallback
	}

	// Determine style.
	selectedStyle := styles.Get(style)
	if selectedStyle == nil {
		selectedStyle = styles.Fallback
	}

	tokens, err := selectedLexer.Tokenise(nil, source)
	if err != nil {
		return err
	}
	return selectedFormatter.Format(w, selectedStyle, tokens)
}
