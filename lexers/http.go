package lexers

import (
	"strings"

	. "github.com/alecthomas/chroma/v2" // nolint
)

// HTTP lexer.
var HTTP = Register(httpBodyContentTypeLexer(MustNewLexer(
	&Config{
		Name:         "HTTP",
		Aliases:      []string{"http"},
		Filenames:    []string{},
		MimeTypes:    []string{},
		NotMultiline: true,
		DotAll:       true,
	},
	httpRules,
)))

func httpRules() Rules {
	return Rules{
		"root": {
			{`(GET|POST|PUT|DELETE|HEAD|OPTIONS|TRACE|PATCH|CONNECT)( +)([^ ]+)( +)(HTTP)(/)([123](?:\.[01])?)(\r?\n|\Z)`, ByGroups(NameFunction, Text, NameNamespace, Text, KeywordReserved, Operator, LiteralNumber, Text), Push("headers")},
			{`(HTTP)(/)([123](?:\.[01])?)( +)(\d{3})( *)([^\r\n]*)(\r?\n|\Z)`, ByGroups(KeywordReserved, Operator, LiteralNumber, Text, LiteralNumber, Text, NameException, Text), Push("headers")},
		},
		"headers": {
			{`([^\s:]+)( *)(:)( *)([^\r\n]+)(\r?\n|\Z)`, EmitterFunc(httpHeaderBlock), nil},
			{`([\t ]+)([^\r\n]+)(\r?\n|\Z)`, EmitterFunc(httpContinuousHeaderBlock), nil},
			{`\r?\n`, Text, Push("content")},
		},
		"content": {
			{`.+`, EmitterFunc(httpContentBlock), nil},
		},
	}
}

func httpContentBlock(groups []string, state *LexerState) Iterator {
	tokens := []Token{
		{Generic, groups[0]},
	}
	return Literator(tokens...)
}

func httpHeaderBlock(groups []string, state *LexerState) Iterator {
	tokens := []Token{
		{Name, groups[1]},
		{Text, groups[2]},
		{Operator, groups[3]},
		{Text, groups[4]},
		{Literal, groups[5]},
		{Text, groups[6]},
	}
	return Literator(tokens...)
}

func httpContinuousHeaderBlock(groups []string, state *LexerState) Iterator {
	tokens := []Token{
		{Text, groups[1]},
		{Literal, groups[2]},
		{Text, groups[3]},
	}
	return Literator(tokens...)
}

// parseContentType strips any parameters (everything from the first ";") off a
// Content-Type header value and trims the surrounding whitespace.
func parseContentType(value string) string {
	contentType := strings.TrimSpace(value)
	if pos := strings.Index(contentType, ";"); pos > 0 {
		contentType = strings.TrimSpace(contentType[:pos])
	}
	return contentType
}

// normalizeStructuredSyntax collapses a structured-syntax suffix such as
// "application/calendar+xml" down to its base type "application/xml", so the
// lookup can fall back to the base type's lexer.
func normalizeStructuredSyntax(contentType string) string {
	slashPos := strings.Index(contentType, "/")
	plusPos := strings.LastIndex(contentType, "+")
	return contentType[:slashPos+1] + contentType[plusPos+1:]
}

func httpBodyContentTypeLexer(lexer Lexer) Lexer { return &httpBodyContentTyper{lexer} }

type httpBodyContentTyper struct{ Lexer }

func (d *httpBodyContentTyper) Tokenise(options *TokeniseOptions, text string) (Iterator, error) { // nolint: gocognit
	var contentType string
	var isContentType bool
	var subIterator Iterator

	it, err := d.Lexer.Tokenise(options, text)
	if err != nil {
		return nil, err
	}

	return func() Token {
		token := it()

		if token == EOF {
			if subIterator != nil {
				return subIterator()
			}
			return EOF
		}

		switch {
		case token.Type == Name && strings.ToLower(token.Value) == "content-type":
			isContentType = true
		case token.Type == Literal && isContentType:
			isContentType = false
			contentType = parseContentType(token.Value)
		case token.Type == Generic && contentType != "":
			lexer := MatchMimeType(contentType)

			// application/calendar+xml can be treated as application/xml
			// if there's not a better match.
			if lexer == nil && strings.Contains(contentType, "+") {
				contentType = normalizeStructuredSyntax(contentType)
				lexer = MatchMimeType(contentType)
			}

			if lexer == nil {
				token.Type = Text
			} else {
				subIterator, err = lexer.Tokenise(nil, token.Value)
				if err != nil {
					panic(err)
				}
				return EOF
			}
		}
		return token
	}, nil
}
