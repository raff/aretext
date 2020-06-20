package text

import (
	"bufio"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func repeat(c rune, n int) string {
	runes := make([]rune, n)
	for i := 0; i < n; i++ {
		runes[i] = c
	}
	return string(runes)
}

func lines(numLines int, charsPerLine int) string {
	lines := make([]string, 0, numLines)
	currentChar := byte(65)

	for i := 0; i < numLines; i++ {
		l := repeat(rune(currentChar), charsPerLine)
		lines = append(lines, l)
		currentChar++
		if currentChar > 90 { // letter Z
			currentChar = 65 // letter A
		}
	}

	return strings.Join(lines, "\n")
}

func allTextFromTree(t *testing.T, tree *Tree) string {
	cursor := tree.CursorAtPosition(0)
	retrievedBytes, err := ioutil.ReadAll(cursor)
	require.NoError(t, err)
	return string(retrievedBytes)
}

func TestEmptyTree(t *testing.T) {
	tree := NewTree()
	text := allTextFromTree(t, tree)
	assert.Equal(t, "", text)
}

func TestTreeBulkLoadAndReadAll(t *testing.T) {
	testCases := []struct {
		name string
		text string
	}{
		{"empty", ""},
		{"single ASCII char", "a"},
		{"multiple ASCII chars", "abcdefg"},
		{"very long ASCII chars", repeat('a', 300000)},
		{"single 2-byte char", "£"},
		{"multiple 2-byte chars", "£ôƊ"},
		{"very long 2-byte chars", repeat('£', 300000)},
		{"single 3-byte char", "፴"},
		{"multiple 3-byte chars:", "፴ऴஅ"},
		{"very long 3-byte char", repeat('፴', 3000000)},
		{"single 4-byte char", "\U0010AAAA"},
		{"multiple 4-byte chars", "\U0010AAAA\U0010BBBB\U0010CCCC"},
		{"very long 4-byte chars", repeat('\U0010AAAA', 300000)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := NewTreeFromString(tc.text)
			require.NoError(t, err)
			text := allTextFromTree(t, tree)
			assert.Equal(t, tc.text, text, "original str had len %d, output string had len %d", len(tc.text), len(text))
		})
	}
}

func TestCursorStartLocation(t *testing.T) {
	testCases := []struct {
		name  string
		runes []rune
	}{
		{
			name:  "short, ASCII",
			runes: []rune{'a', 'b', 'c', 'd'},
		},
		{
			name:  "short, mixed width characters",
			runes: []rune{'a', '£', 'b', '፴', 'c', 'd', '\U0010AAAA', 'e', 'ऴ'},
		},
		{
			name:  "medium, ASCII",
			runes: []rune(repeat('a', 4096)),
		},
		{
			name:  "short, 2-byte chars",
			runes: []rune(repeat('£', 10)),
		},
		{
			name:  "medium, 2-byte chars",
			runes: []rune(repeat('£', 4096)),
		},
		{
			name:  "short, 3-byte chars",
			runes: []rune(repeat('፴', 5)),
		},
		{
			name:  "medium, 3-byte chars",
			runes: []rune(repeat('፴', 4096)),
		},
		{
			name:  "short, 4-byte chars",
			runes: []rune(repeat('\U0010AAAA', 5)),
		},
		{
			name:  "medium, 4-byte chars",
			runes: []rune(repeat('\U0010AAAA', 4096)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := NewTreeFromString(string(tc.runes))
			require.NoError(t, err)

			// Check a cursor starting from each character position to the end
			for i := 0; i < len(tc.runes); i++ {
				cursor := tree.CursorAtPosition(uint64(i))
				retrieved, err := ioutil.ReadAll(cursor)
				require.NoError(t, err)
				require.Equal(t, string(tc.runes[i:]), string(retrieved), "invalid substring starting from character at position %d (expected len = %d, actual len = %d)", i, len(string(tc.runes[i:])), len(string(retrieved)))
			}
		})
	}
}

func TestCursorPastLastCharacter(t *testing.T) {
	testCases := []struct {
		name string
		text string
		pos  uint64
	}{
		{
			name: "empty, position zero",
			text: "",
			pos:  0,
		},
		{
			name: "empty, position one",
			text: "",
			pos:  1,
		},
		{
			name: "single char, position one",
			text: "a",
			pos:  1,
		},
		{
			name: "single char, position two",
			text: "a",
			pos:  2,
		},
		{
			name: "full leaf, position at end of leaf",
			text: repeat('a', maxBytesPerLeaf),
			pos:  maxBytesPerLeaf,
		},
		{
			name: "full leaf, position one after end of leaf",
			text: repeat('b', maxBytesPerLeaf),
			pos:  maxBytesPerLeaf + 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := NewTreeFromString(tc.text)
			require.NoError(t, err)
			cursor := tree.CursorAtPosition(tc.pos)
			retrieved, err := ioutil.ReadAll(cursor)
			require.NoError(t, err)
			assert.Equal(t, "", string(retrieved))
		})
	}
}

func TestCursorAtLine(t *testing.T) {
	testCases := []struct {
		name string
		text string
	}{
		{
			name: "empty",
			text: "",
		},
		{
			name: "single newline",
			text: "\n",
		},
		{
			name: "two newlines",
			text: "\n\n",
		},
		{
			name: "single line, same leaf",
			text: lines(1, 12),
		},
		{
			name: "single line, multiple leaves",
			text: lines(1, 4096),
		},
		{
			name: "two lines, same leaf",
			text: lines(2, 4),
		},
		{
			name: "two lines, multiple leaves",
			text: lines(2, 4096),
		},
		{
			name: "many lines, single character per line",
			text: lines(4096, 1),
		},
		{
			name: "many lines, many characters per line",
			text: lines(4096, 1024),
		},
		{
			name: "many lines, newline on previous leaf",
			text: lines(1024, maxBytesPerLeaf-1),
		},
		{
			name: "many lines, newline on next leaf",
			text: lines(1024, maxBytesPerLeaf),
		},
	}

	linesFromTree := func(tree *Tree, numLines int) []string {
		lines := make([]string, 0, numLines)
		for i := 0; i < numLines; i++ {
			cursor := tree.CursorAtLine(uint64(i))
			scanner := bufio.NewScanner(cursor)
			scanner.Split(bufio.ScanLines)

			for scanner.Scan() {
				lines = append(lines, scanner.Text())
				break
			}
		}
		return lines
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lines := strings.Split(tc.text, "\n")
			if len(lines) > 0 && lines[len(lines)-1] == "" {
				// match bufio.ScanLines behavior, which ignores last empty line
				lines = lines[:len(lines)-1]
			}

			tree, err := NewTreeFromString(tc.text)
			require.NoError(t, err)
			actualLines := linesFromTree(tree, len(lines))
			assert.Equal(t, lines, actualLines, "expected lines = %v, actual lines = %v", len(lines), len(actualLines))
		})
	}
}

func TestCursorPastLastLine(t *testing.T) {
	testCases := []struct {
		name    string
		text    string
		lineNum uint64
	}{
		{
			name:    "empty, line zero",
			text:    "",
			lineNum: 0,
		},
		{
			name:    "empty, line one",
			text:    "",
			lineNum: 1,
		},
		{
			name:    "single line, line one",
			text:    "abcdefgh",
			lineNum: 1,
		},
		{
			name:    "single line, line two",
			text:    "abcdefgh",
			lineNum: 2,
		},
		{
			name:    "multiple lines, one past last line",
			text:    "abc\ndefg\nhijk",
			lineNum: 3,
		},
		{
			name:    "multiple lines, two past last line",
			text:    "abc\ndefg\nhijk",
			lineNum: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := NewTreeFromString(tc.text)
			require.NoError(t, err)
			cursor := tree.CursorAtLine(tc.lineNum)
			retrieved, err := ioutil.ReadAll(cursor)
			require.NoError(t, err)
			assert.Equal(t, "", string(retrieved))
		})
	}
}

func TestDeleteAtPosition(t *testing.T) {
	testCases := []struct {
		name         string
		inputText    string
		deletePos    uint64
		expectedText string
	}{
		{
			name:         "empty",
			inputText:    "",
			deletePos:    0,
			expectedText: "",
		},
		{
			name:         "single character",
			inputText:    "A",
			deletePos:    0,
			expectedText: "",
		},
		{
			name:         "single character, delete past end",
			inputText:    "A",
			deletePos:    1,
			expectedText: "A",
		},
		{
			name:         "two characters, delete first",
			inputText:    "AB",
			deletePos:    0,
			expectedText: "B",
		},
		{
			name:         "two characters, delete second",
			inputText:    "AB",
			deletePos:    1,
			expectedText: "A",
		},
		{
			name:         "multi-byte character, delete before",
			inputText:    "a£b",
			deletePos:    0,
			expectedText: "£b",
		},
		{
			name:         "multi-byte character, delete on",
			inputText:    "a£b",
			deletePos:    1,
			expectedText: "ab",
		},
		{
			name:         "multi-byte character, delete after",
			inputText:    "a£b",
			deletePos:    2,
			expectedText: "a£",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := NewTreeFromString(tc.inputText)
			require.NoError(t, err)
			tree.DeleteAtPosition(tc.deletePos)
			text := allTextFromTree(t, tree)
			assert.Equal(t, tc.expectedText, text)
		})
	}
}

func TestDeleteAllCharsInLongStringFromBeginning(t *testing.T) {
	testCases := []struct {
		name string
		text string
	}{
		{
			name: "ASCII",
			text: repeat('a', 4096),
		},
		{
			name: "2-byte chars",
			text: repeat('£', 4096),
		},
		{
			name: "3-byte chars",
			text: repeat('፴', 4096),
		},
		{
			name: "4-byte chars",
			text: repeat('\U0010AAAA', 4096),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := NewTreeFromString(tc.text)
			require.NoError(t, err)
			for i := 0; i < len(tc.text); i++ {
				tree.DeleteAtPosition(0)
			}
			text := allTextFromTree(t, tree)
			assert.Equal(t, "", text)
		})
	}
}

func TestDeleteAllCharsInLongStringFromEnd(t *testing.T) {
	testCases := []struct {
		name string
		text string
	}{
		{
			name: "ASCII",
			text: repeat('a', 4096),
		},
		{
			name: "2-byte chars",
			text: repeat('£', 4096),
		},
		{
			name: "3-byte chars",
			text: repeat('፴', 4096),
		},
		{
			name: "4-byte chars",
			text: repeat('\U0010AAAA', 4096),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := NewTreeFromString(tc.text)
			require.NoError(t, err)
			for i := len(tc.text) - 1; i >= 0; i-- {
				tree.DeleteAtPosition(0)
			}
			text := allTextFromTree(t, tree)
			assert.Equal(t, "", text)
		})
	}
}

func TestDeleteNewline(t *testing.T) {
	tree, err := NewTreeFromString(lines(4096, 100))
	require.NoError(t, err)

	cursor := tree.CursorAtLine(4094) // read last two lines
	text, err := ioutil.ReadAll(cursor)
	require.NoError(t, err)
	assert.Equal(t, 201, len(text))

	tree.DeleteAtPosition(100)       // delete first newline
	cursor = tree.CursorAtLine(4094) // read last line
	text, err = ioutil.ReadAll(cursor)
	require.NoError(t, err)
	assert.Equal(t, 100, len(text))
}

func benchmarkLoad(b *testing.B, numBytes int) {
	text := repeat('a', numBytes)
	for n := 0; n < b.N; n++ {
		_, err := NewTreeFromString(text)
		if err != nil {
			b.Fatalf("err = %v", err)
		}
	}
}

func benchmarkRead(b *testing.B, numBytes int) {
	text := repeat('a', numBytes)
	tree, err := NewTreeFromString(text)
	if err != nil {
		b.Fatalf("err = %v", err)
	}

	for n := 0; n < b.N; n++ {
		cursor := tree.CursorAtPosition(0)
		_, err := ioutil.ReadAll(cursor)
		if err != nil {
			b.Fatalf("err = %v", err)
		}
	}
}

func BenchmarkLoad4096Bytes(b *testing.B)    { benchmarkLoad(b, 4096) }
func BenchmarkLoad1048576Bytes(b *testing.B) { benchmarkLoad(b, 1048576) }
func BenchmarkRead4096Bytes(b *testing.B)    { benchmarkRead(b, 4096) }
func BenchmarkRead1048576Bytes(b *testing.B) { benchmarkRead(b, 1048576) }