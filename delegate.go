package chroma

import (
	"bytes"
)

type delegatingLexer struct {
	root     Lexer
	language Lexer
}

// DelegatingLexer combines two lexers to handle the common case of a language embedded inside another, such as PHP
// inside HTML or PHP inside plain text.
//
// It takes two lexer as arguments: a root lexer and a language lexer.  First everything is scanned using the language
// lexer, which must return "Other" for unrecognised tokens. Then all "Other" tokens are lexed using the root lexer.
// Finally, these two sets of tokens are merged.
//
// The lexers from the template lexer package use this base lexer.
func DelegatingLexer(root Lexer, language Lexer) Lexer {
	return &delegatingLexer{
		root:     root,
		language: language,
	}
}

func (d *delegatingLexer) SetTracing(enable bool) {
	if l, ok := d.language.(TracingLexer); ok {
		l.SetTracing(enable)
	}
	if l, ok := d.root.(TracingLexer); ok {
		l.SetTracing(enable)
	}
}

func (d *delegatingLexer) AnalyseText(text string) float32 {
	return d.root.AnalyseText(text)
}

func (d *delegatingLexer) SetAnalyser(analyser func(text string) float32) Lexer {
	d.root.SetAnalyser(analyser)
	return d
}

func (d *delegatingLexer) SetRegistry(r *LexerRegistry) Lexer {
	d.root.SetRegistry(r)
	d.language.SetRegistry(r)
	return d
}

func (d *delegatingLexer) Config() *Config {
	return d.language.Config()
}

// An insertion is the character range where language tokens should be inserted.
type insertion struct {
	start, end int
	tokens     []Token
}

func (d *delegatingLexer) Tokenise(options *TokeniseOptions, text string) (Iterator, error) { // nolint: gocognit
	languageTokens, err := Tokenise(Coalesce(d.language), options, text)
	if err != nil {
		return nil, err
	}
	// Compute insertions and gather "Other" languageTokens.
	otherText := &bytes.Buffer{}
	insertions := []*insertion{}
	var currentInsertion *insertion
	position := 0
	var previous Token
	for _, token := range languageTokens {
		switch token.Type {
		case Other:
			if previous != EOF && currentInsertion != nil && previous.Type != Other {
				currentInsertion.end = position
			}
			otherText.WriteString(token.Value)

		default:
			if previous == EOF || previous.Type == Other {
				currentInsertion = &insertion{start: position}
				insertions = append(insertions, currentInsertion)
			}
			currentInsertion.tokens = append(currentInsertion.tokens, token)
		}
		previous = token
		position += len(token.Value)
	}

	if len(insertions) == 0 {
		return d.root.Tokenise(options, text)
	}

	// Lex the other languageTokens.
	rootTokens, err := Tokenise(Coalesce(d.root), options, otherText.String())
	if err != nil {
		return nil, err
	}

	// Interleave the two sets of languageTokens.
	var output []Token
	position = 0 // Offset into text.
	tokenIndex := 0
	nextToken := func() Token {
		if tokenIndex >= len(rootTokens) {
			return EOF
		}
		token := rootTokens[tokenIndex]
		tokenIndex++
		return token
	}
	insertionIndex := 0
	nextInsertion := func() *insertion {
		if insertionIndex >= len(insertions) {
			return nil
		}
		entry := insertions[insertionIndex]
		insertionIndex++
		return entry
	}
	token := nextToken()
	entry := nextInsertion()
	for token != EOF || entry != nil {
		if token == EOF || (entry != nil && entry.start < position+len(token.Value)) {
			var leading Token
			leading, token = splitToken(token, entry.start-position)
			if leading != EOF {
				output = append(output, leading)
				position += len(leading.Value)
			}
			output = append(output, entry.tokens...)
			position += entry.end - entry.start
			if token == EOF {
				token = nextToken()
			}
			entry = nextInsertion()
		} else {
			output = append(output, token)
			position += len(token.Value)
			token = nextToken()
		}
	}
	return Literator(output...), nil
}

func splitToken(t Token, offset int) (l Token, r Token) {
	if t == EOF {
		return EOF, EOF
	}
	if offset == 0 {
		return EOF, t
	}
	if offset == len(t.Value) {
		return t, EOF
	}
	l = t.Clone()
	r = t.Clone()
	l.Value = l.Value[:offset]
	r.Value = r.Value[offset:]
	return
}
