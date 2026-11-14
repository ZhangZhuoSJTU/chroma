package formatters

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/alecthomas/chroma/v2"
)

// JSON formatter outputs the raw token structures as JSON.
var JSON = Register("json", chroma.FormatterFunc(func(w io.Writer, s *chroma.Style, it chroma.Iterator) error {
	if _, err := fmt.Fprintln(w, "["); err != nil {
		return err
	}
	first := true
	for {
		token := it()
		if token == chroma.EOF {
			break
		}

		if !first {
			if _, err := fmt.Fprintln(w, ","); err != nil {
				return err
			}
		}
		first = false
		encoded, err := json.Marshal(token)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprint(w, "  "+string(encoded)); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "]"); err != nil {
		return err
	}
	return nil
}))
