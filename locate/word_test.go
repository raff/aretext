package locate

import (
	"testing"

	"github.com/aretext/aretext/syntax"
	"github.com/aretext/aretext/text"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNextWordStart(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		syntaxLanguage syntax.Language
		pos            uint64
		expectedPos    uint64
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "next word from current word, same line",
			inputString: "abc   defg   hij",
			pos:         1,
			expectedPos: 6,
		},
		{
			name:        "next word from whitespace, same line",
			inputString: "abc   defg   hij",
			pos:         4,
			expectedPos: 6,
		},
		{
			name:        "next word from different line",
			inputString: "abc\n   123",
			pos:         1,
			expectedPos: 7,
		},
		{
			name:        "next word to empty line",
			inputString: "abc\n\n   123",
			pos:         1,
			expectedPos: 4,
		},
		{
			name:        "empty line to next word",
			inputString: "abc\n\n   123",
			pos:         4,
			expectedPos: 8,
		},
		{
			name:        "multiple empty lines",
			inputString: "\n\n\n\n",
			pos:         1,
			expectedPos: 2,
		},
		{
			name:           "next syntax token",
			inputString:    "123+456",
			syntaxLanguage: syntax.LanguageGo,
			pos:            1,
			expectedPos:    3,
		},
		{
			name:           "next syntax token skip empty",
			inputString:    "123    +      456",
			syntaxLanguage: syntax.LanguageGo,
			pos:            1,
			expectedPos:    7,
		},
		{
			name:           "syntax token starts with whitespace",
			inputString:    "//    foobar",
			syntaxLanguage: syntax.LanguageGo,
			pos:            0,
			expectedPos:    6,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			tokenTree, err := syntax.TokenizeString(tc.syntaxLanguage, tc.inputString)
			require.NoError(t, err)
			actualPos := NextWordStart(textTree, tokenTree, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestNextWordEnd(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		syntaxLanguage syntax.Language
		pos            uint64
		expectedPos    uint64
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "end of word from start of current word",
			inputString: "abc   defg   hij",
			pos:         6,
			expectedPos: 9,
		},
		{
			name:        "end of word from middle of current word",
			inputString: "abc   defg   hij",
			pos:         7,
			expectedPos: 9,
		},
		{
			name:        "next word from end of current word",
			inputString: "abc   defg   hij",
			pos:         2,
			expectedPos: 9,
		},
		{
			name:        "next word from whitespace",
			inputString: "abc   defg   hij",
			pos:         4,
			expectedPos: 9,
		},
		{
			name:        "next word past empty line",
			inputString: "abc\n\n   123   xyz",
			pos:         2,
			expectedPos: 10,
		},
		{
			name:        "empty line to next word",
			inputString: "abc\n\n   123  xyz",
			pos:         4,
			expectedPos: 10,
		},
		{
			name:           "next syntax token",
			inputString:    "123+456",
			syntaxLanguage: syntax.LanguageGo,
			pos:            2,
			expectedPos:    3,
		},
		{
			name:           "next syntax token skip empty",
			inputString:    "123    +      456",
			syntaxLanguage: syntax.LanguageGo,
			pos:            2,
			expectedPos:    7,
		},
		{
			name:           "next syntax token ends with whitespace",
			inputString:    `"    abcd    "`,
			syntaxLanguage: syntax.LanguageGo,
			pos:            8,
			expectedPos:    13,
		},
		{
			name:           "end of current syntax token",
			inputString:    "123+456",
			syntaxLanguage: syntax.LanguageGo,
			pos:            0,
			expectedPos:    2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			tokenTree, err := syntax.TokenizeString(tc.syntaxLanguage, tc.inputString)
			require.NoError(t, err)
			actualPos := NextWordEnd(textTree, tokenTree, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestPrevWordStart(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		syntaxLanguage syntax.Language
		pos            uint64
		expectedPos    uint64
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "prev word from current word, same line",
			inputString: "abc   defg   hij",
			pos:         6,
			expectedPos: 0,
		},
		{
			name:        "prev word from whitespace, same line",
			inputString: "abc   defg   hij",
			pos:         12,
			expectedPos: 6,
		},
		{
			name:        "prev word from different line",
			inputString: "abc\n   123",
			pos:         7,
			expectedPos: 0,
		},
		{
			name:        "prev word to empty line",
			inputString: "abc\n\n   123",
			pos:         8,
			expectedPos: 4,
		},
		{
			name:        "empty line to prev word",
			inputString: "abc\n\n   123",
			pos:         4,
			expectedPos: 0,
		},
		{
			name:        "multiple empty lines",
			inputString: "\n\n\n\n",
			pos:         2,
			expectedPos: 1,
		},
		{
			name:           "prev syntax token",
			inputString:    "123+456",
			syntaxLanguage: syntax.LanguageGo,
			pos:            4,
			expectedPos:    3,
		},
		{
			name:           "prev syntax token skip empty",
			inputString:    "123    +      456",
			syntaxLanguage: syntax.LanguageGo,
			pos:            14,
			expectedPos:    7,
		},
		{
			name:           "prev syntax token starts with whitespace",
			inputString:    "// abcd",
			syntaxLanguage: syntax.LanguageGo,
			pos:            3,
			expectedPos:    0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			tokenTree, err := syntax.TokenizeString(tc.syntaxLanguage, tc.inputString)
			require.NoError(t, err)
			actualPos := PrevWordStart(textTree, tokenTree, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestCurrentWordStart(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		syntaxLanguage syntax.Language
		pos            uint64
		expectedPos    uint64
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "start of document",
			inputString: "abc   defg   hij",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "start of word in middle of document",
			inputString: "abc   defg   hij",
			pos:         6,
			expectedPos: 6,
		},
		{
			name:        "middle of word to start of word",
			inputString: "abc   defg   hij",
			pos:         8,
			expectedPos: 6,
		},
		{
			name:        "end of word to start of word",
			inputString: "abc   defg   hij",
			pos:         9,
			expectedPos: 6,
		},
		{
			name:        "start of whitespace",
			inputString: "abc   defg   hij",
			pos:         3,
			expectedPos: 3,
		},
		{
			name:        "middle of whitespace",
			inputString: "abc   defg   hij",
			pos:         4,
			expectedPos: 3,
		},
		{
			name:        "end of whitespace",
			inputString: "abc   defg   hij",
			pos:         5,
			expectedPos: 3,
		},
		{
			name:        "word at start of line",
			inputString: "abc\nxyz",
			pos:         5,
			expectedPos: 4,
		},
		{
			name:        "whitespace at start of line",
			inputString: "abc\n    xyz",
			pos:         6,
			expectedPos: 4,
		},
		{
			name:        "empty line",
			inputString: "abc\n\n   123",
			pos:         4,
			expectedPos: 4,
		},
		{
			name:           "adjacent syntax tokens",
			inputString:    "123+456",
			syntaxLanguage: syntax.LanguageGo,
			pos:            5,
			expectedPos:    4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			tokenTree, err := syntax.TokenizeString(tc.syntaxLanguage, tc.inputString)
			require.NoError(t, err)
			actualPos := CurrentWordStart(textTree, tokenTree, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}

func TestCurrentWordEnd(t *testing.T) {
	testCases := []struct {
		name           string
		inputString    string
		syntaxLanguage syntax.Language
		pos            uint64
		expectedPos    uint64
	}{
		{
			name:        "empty",
			inputString: "",
			pos:         0,
			expectedPos: 0,
		},
		{
			name:        "end of document",
			inputString: "abc   defg   hijk",
			pos:         14,
			expectedPos: 17,
		},
		{
			name:        "start of word in middle of document",
			inputString: "abc   defg   hij",
			pos:         6,
			expectedPos: 10,
		},
		{
			name:        "middle of word to end of word",
			inputString: "abc   defg   hij",
			pos:         7,
			expectedPos: 10,
		},
		{
			name:        "end of word",
			inputString: "abc   defg   hij",
			pos:         9,
			expectedPos: 10,
		},
		{
			name:        "start of whitespace",
			inputString: "abc   defg   hij",
			pos:         3,
			expectedPos: 6,
		},
		{
			name:        "middle of whitespace",
			inputString: "abc   defg   hij",
			pos:         4,
			expectedPos: 6,
		},
		{
			name:        "end of whitespace",
			inputString: "abc   defg   hij",
			pos:         5,
			expectedPos: 6,
		},
		{
			name:        "word before end of line",
			inputString: "abc\nxyz",
			pos:         1,
			expectedPos: 3,
		},
		{
			name:        "whitespace at end of line",
			inputString: "abc     \nxyz",
			pos:         4,
			expectedPos: 8,
		},
		{
			name:        "empty line",
			inputString: "abc\n\n   123",
			pos:         4,
			expectedPos: 4,
		},
		{
			name:           "adjacent syntax tokens",
			inputString:    "123+456",
			syntaxLanguage: syntax.LanguageGo,
			pos:            1,
			expectedPos:    3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			textTree, err := text.NewTreeFromString(tc.inputString)
			require.NoError(t, err)
			tokenTree, err := syntax.TokenizeString(tc.syntaxLanguage, tc.inputString)
			require.NoError(t, err)
			actualPos := CurrentWordEnd(textTree, tokenTree, tc.pos)
			assert.Equal(t, tc.expectedPos, actualPos)
		})
	}
}