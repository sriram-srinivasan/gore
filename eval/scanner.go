package eval

import (
	"strings"
)

type Scanner struct {
	Reader *strings.Reader
	Input  string
}

func NewScanner(text string) *Scanner {
	reader := strings.NewReader(text)
	return &Scanner{Reader: reader, Input: text}
}

func (scanner *Scanner) Mark() int {
	return scanner.Reader.Len()
}

func (scanner *Scanner) Reset(mark int) {
	r := scanner.Reader
	offset := r.Len() - mark
	_, err := r.Seek(int64(offset), 1) // relative move
	chk(err)
}

func (scanner *Scanner) ReadRune() (ch rune, err error) {
	ch, _, err = scanner.Reader.ReadRune()
	return ch, err
}

func (scanner *Scanner) UnreadRune() {
	err := scanner.Reader.UnreadRune()
	chk(err)
}

func (scanner *Scanner) Slice(mark int) (s string) {
	begin := len(scanner.Input) - mark
	end := scanner.Pos()
	return scanner.Input[begin:end]
}

func (scanner *Scanner) Pos() int {
	return len(scanner.Input) - scanner.Reader.Len()
}

// Panic if unexpected error
func chk(err error) {
	if err != nil {
		panic(err)
	}
}
