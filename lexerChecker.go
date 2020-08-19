package abc

import (
	"bufio"
	"bytes"
	"regexp"
)

//lexerChecker is a support go file that checks whether a byte is something.
//As such, all functions in this file start with "is" and the rest is based on the ebnf of the abc files.
//it supports the act of reading the music body.
type byteToken struct {
	token []byte
}

func peekLexToken(r *bufio.Reader) (byteToken, error) {
	var t byteToken
	var err error
	t.token = make([]byte, 2, 2)
	t.token, err = r.Peek(2)
	if err != nil {
		return t, err
	}
	return t, nil
}

func (t *byteToken) isComment() bool {
	return t.token[0] == '%'
}

func (t *byteToken) isBrokenRhythm() bool {
	return t.token[0] == '<' || t.token[0] == '>'
}

func (t *byteToken) isNewline() bool {
	return t.token[0] == '\n'
}

func (t *byteToken) isElement() bool {
	return t.isNote() || t.isAnnotation() || t.isBarline() ||
		t.isSpace() || t.isInline() || t.isRepeat() || t.isChord() || t.isBrokenRhythm()
}

func (t *byteToken) isNote() bool {
	return t.isRest() || t.isPitch()
}

func (t *byteToken) isSpace() bool {
	return t.token[0] == ' '
}

func (t *byteToken) isAnnotation() bool {
	return t.token[0] == '"'
}

func (t *byteToken) isChord() bool {
	re := regexp.MustCompile(`[a-gA-G]`)
	return re.Match(t.token) && t.token[0] == '['
}

func (t *byteToken) isBarline() bool {
	var ret bool = false
	ret = ret || t.token[0] == '|'
	ret = ret || bytes.Compare(t.token, []byte("||")) == 0
	ret = ret || bytes.Compare(t.token, []byte("[|")) == 0
	ret = ret || bytes.Compare(t.token, []byte("|]")) == 0
	ret = ret || bytes.Compare(t.token, []byte(":|")) == 0
	ret = ret || bytes.Compare(t.token, []byte("|:")) == 0

	return ret

}

func (t *byteToken) isInline() bool {
	return t.token[0] == '[' && t.token[1] != '|'
}

func (t *byteToken) isRepeat() bool {
	if t.token[0] == '[' {
		return t.token[1] == '1' || t.token[1] == '2'
	}
	return false
}

func (t *byteToken) isRest() bool {
	return t.token[0] == 'z'
}

func (t *byteToken) isPitch() bool {
	re := regexp.MustCompile(`[',a-gA-G]`)
	return re.Match(t.token[:1])
}

func (t *byteToken) isDigit() bool {
	re := regexp.MustCompile(`[0-9]`)
	return re.Match([]byte(t.token[:1]))
}

func (t *byteToken) isTuneBodyInfoField() bool {
	re := regexp.MustCompile(`[A-Zw]:`)
	return re.Match(t.token)
}
