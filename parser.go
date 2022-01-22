package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
	"unicode/utf8"
)

const eof = -1

type lexer struct {
	input      string
	start, pos int
	size       int // size of the last rune
	tokens     []token
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.size = 0 // didn't read anything: necessary for subsequent backup
		return eof
	}
	var r rune
	r, l.size = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.size
	return r
}

// backup can be called once per call of next
func (l *lexer) backup() {
	l.pos -= l.size
}

func (l *lexer) emit(typ tokenType) {
	l.tokens = append(l.tokens, token{typ: typ, text: l.input[l.start:l.pos]})
	l.start = l.pos
}

type tokenType int

const (
	tokenLLink tokenType = iota
	tokenRLink
	tokenHeading
	tokenNewLine
	tokenText

	tokenEOF
)

func (t tokenType) String() string {
	return [...]string{"LL", "RL", "HEAD", "NL", "TEXT", "EOF"}[t]
}

type token struct {
	typ  tokenType
	text string
}

func (t token) String() string {
	return fmt.Sprintf("{%v %q}", t.typ, t.text)
}

type stateFn func(l *lexer) stateFn

func lex(l *lexer) stateFn {
	r := l.next()
	switch r {
	case '[':
		l.emit(tokenLLink)
		return lex
	case ']':
		l.emit(tokenRLink)
		return lex
	case '#':
		l.emit(tokenHeading)
		return lex
	case '\n':
		l.emit(tokenNewLine)
		return lex
	case eof:
		l.emit(tokenEOF)
		return nil
	}
	return lexText
}

func lexText(l *lexer) stateFn {
	for {
		r := l.next()
		if r == '[' || r == ']' || r == '#' || r == '\n' || r == eof {
			break
		}
	}
	l.backup()
	l.emit(tokenText)
	return lex
}

func main() {
	buf, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	l := lexer{
		input: string(buf),
	}
	state := lex
	for state != nil {
		fmt.Println(l.tokens)
		time.Sleep(1 * time.Second)
		state = state(&l)
	}
	fmt.Println(l.tokens)
}
