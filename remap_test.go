package chroma

import (
	"testing"

	assert "github.com/alecthomas/assert/v2"
)

func TestRemappingLexer(t *testing.T) {
	rules := Rules{
		"root": {
			{`\s+`, Whitespace, nil},
			{`\w+`, Name, nil},
		},
	}

	base := mustNewLexer(t, nil, rules) // nolint: forbidigo
	lexer := TypeRemappingLexer(base, TypeMapping{
		{Name, Keyword, []string{"if", "else"}},
	})

	it, err := lexer.Tokenise(nil, "if true then print else end")
	assert.NoError(t, err)

	words := []string{"if", "true", "then", "print", "else", "end"}
	expected := make([]Token, 0, len(words)*2-1)

	for i, word := range words {
		if i > 0 {
			expected = append(expected, Token{TextWhitespace, " "})
		}

		tokenType := Name
		if word == "if" || word == "else" {
			tokenType = Keyword
		}
		expected = append(expected, Token{tokenType, word})
	}

	assert.Equal(t, expected, it.Tokens())
}