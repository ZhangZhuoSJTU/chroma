package formatters

import (
	"fmt"
	"io"

	"github.com/alecthomas/chroma/v2"
)

// Tokens formatter outputs the raw token structures.
var Tokens = Register("tokens", chroma.FormatterFunc(func(w io.Writer, s *chroma.Style, it chroma.Iterator) error {
	for {
		// Termination belongs to the value just acquired, rather than the
		// acquisition that follows the last value whose representation was emitted.
		current := it()
		if current == chroma.EOF {
			return nil
		}
		// Keep the representation boundary and the line boundary in one write call.
		if _, err := fmt.Fprintln(w, current.GoString()); err != nil {
			return err
		}
	}
}))
