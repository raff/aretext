package locate

import (
	"github.com/aretext/aretext/syntax/parser"
	"github.com/aretext/aretext/text"
)

// MatchingCodeBlockDelimiter locates the matching paren or brace at a position, if it exists.
func MatchingCodeBlockDelimiter(textTree *text.Tree, syntaxParser *parser.P, pos uint64) (uint64, bool) {
	reader := textTree.ReaderAtPosition(pos)
	r, _, err := reader.ReadRune()
	if err != nil || !(r == '(' || r == ')' || r == '{' || r == '}' || r == '[' || r == ']') {
		return 0, false
	}

	switch r {
	case '(':
		return searchForwardMatch(textTree, syntaxParser, pos, '(', ')')
	case ')':
		return searchBackwardMatch(textTree, syntaxParser, pos, '(', ')')
	case '[':
		return searchForwardMatch(textTree, syntaxParser, pos, '[', ']')
	case ']':
		return searchBackwardMatch(textTree, syntaxParser, pos, '[', ']')
	case '{':
		return searchForwardMatch(textTree, syntaxParser, pos, '{', '}')
	case '}':
		return searchBackwardMatch(textTree, syntaxParser, pos, '{', '}')
	default:
		return 0, false
	}
}

func searchForwardMatch(textTree *text.Tree, syntaxParser *parser.P, pos uint64, openRune rune, closeRune rune) (uint64, bool) {
	startToken := stringOrCommentTokenAtPos(syntaxParser, pos)
	pos++
	depth := 1
	reader := textTree.ReaderAtPosition(pos)
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return 0, false
		}

		if r == openRune || r == closeRune {
			if startToken == stringOrCommentTokenAtPos(syntaxParser, pos) {
				if r == openRune {
					depth++
				} else {
					depth--
				}
			}
		}

		if depth == 0 {
			return pos, true
		}

		pos++
	}
}

func searchBackwardMatch(textTree *text.Tree, syntaxParser *parser.P, pos uint64, openRune rune, closeRune rune) (uint64, bool) {
	startToken := stringOrCommentTokenAtPos(syntaxParser, pos)
	depth := 1
	reader := textTree.ReverseReaderAtPosition(pos)
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return 0, false
		}

		pos--

		if r == openRune || r == closeRune {
			if startToken == stringOrCommentTokenAtPos(syntaxParser, pos) {
				if r == openRune {
					depth--
				} else if r == closeRune {
					depth++
				}
			}
		}

		if depth == 0 {
			return pos, true
		}
	}
}

func stringOrCommentTokenAtPos(syntaxParser *parser.P, pos uint64) parser.Token {
	if syntaxParser == nil {
		return parser.Token{}
	}
	token := syntaxParser.TokenAtPosition(pos)
	if token.Role != parser.TokenRoleComment && token.Role != parser.TokenRoleString {
		return parser.Token{}
	}
	return token
}