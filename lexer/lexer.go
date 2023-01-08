package lexer

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strconv"
	"unicode"
)

type TokenType int

type LexicalToken struct {
	TokenType TokenType
	Value     string
	Line      int
	Position  int
	File      string
}

func (token LexicalToken) ToString() string {
	return token.Value + " at line " + strconv.Itoa(token.Line) + " position " + strconv.Itoa(token.Position) + " at file " + token.File
}

func (token LexicalToken) GetType() TokenType {
	return token.TokenType
}

func (token LexicalToken) GetValue() string {
	return token.Value
}

func (token LexicalToken) GetLine() int {
	return token.Line
}

func (token LexicalToken) GetPosition() int {
	return token.Position
}

const (
	Keyword TokenType = iota
	Operator
	Identifier
	LeftParenthesis
	RightParenthesis
	NewLine
	Coma
	SemiColon
	Text
	Integer
)

const (
	Function = "function"
	Return   = "return"
	End      = "end"
	If       = "if"
	Else     = "else"
	True     = "true"
	False    = "false"
	While    = "while"
)

const (
	Gets           = "="
	Plus           = "+"
	Minus          = "-"
	Multiply       = "*"
	Divide         = "/"
	Equal          = "=="
	NotEqual       = "!="
	Greater        = ">"
	GreaterOrEqual = ">="
	Less           = "<"
	LessOrEqual    = "<="
	Not            = "!"
	Increase       = "++"
	Decrease       = "--"
)

func isKeyword(word string) bool {
	keywords := [...]string{Function, Return, End, If, Else, True, False, While}
	for _, r := range keywords {
		if r == word {
			return true
		}
	}
	return false
}

func isOperator(op string) bool {
	operators := [...]string{Gets, Plus, Minus, Multiply, Divide, Equal, NotEqual, Greater, GreaterOrEqual, Less, LessOrEqual, Not, Increase, Decrease}
	for _, r := range operators {
		if r == op {
			return true
		}
	}
	return false
}

type state struct {
	line          int
	oldPosition   int
	position      int
	tokenPosition int
	tokenLine     int
	tokens        []LexicalToken
	reader        *bufio.Reader
	buffer        bytes.Buffer
	err           error
	r             rune
	file          string
}

func Lex(br *bufio.Reader, fileName string) ([]LexicalToken, error) {
	s := &state{file: fileName}
	s.line = 1
	s.reader = br
	for s.err == nil {
		s.next()
		if unicode.IsLetter(s.r) {
			s.buffer.WriteRune(s.r)
			s.lexIdentifierOrKeyword()
		} else if unicode.IsDigit(s.r) {
			s.buffer.WriteRune(s.r)
			s.lexInteger()
		} else if s.r == '"' {
			s.lexString()
		} else if s.r == '(' {
			t := s.tokenTemplate()
			t.TokenType = LeftParenthesis
			s.tokens = append(s.tokens, t)
		} else if s.r == ')' {
			t := s.tokenTemplate()
			t.TokenType = RightParenthesis
			s.tokens = append(s.tokens, t)
		} else if s.r == ',' {
			t := s.tokenTemplate()
			t.TokenType = Coma
			s.tokens = append(s.tokens, t)
		} else if s.r == ';' {
			t := s.tokenTemplate()
			t.TokenType = SemiColon
			s.tokens = append(s.tokens, t)
		} else if s.r == '\n' {
			t := s.tokenTemplate()
			t.TokenType = NewLine
			s.tokens = append(s.tokens, t)
		} else if isOperator(string(s.r)) {
			s.buffer.WriteRune(s.r)
			s.lexOperator()
		}
	}
	if s.err == io.EOF {
		return s.tokens, nil
	}
	return nil, s.err
}

func (s *state) advance() {
	s.r, _, s.err = s.reader.ReadRune()
	s.oldPosition = s.position
	if s.r == '\n' {
		s.line++
		s.position = 0
	} else {
		s.position++
	}
}

func (s *state) next() {
	s.advance()
	s.tokenLine = s.line
	s.tokenPosition = s.position
	s.buffer.Reset()
}

func (s *state) fallBack() {
	s.reader.UnreadRune()
	if s.r == '\n' {
		s.line--
	}
	s.position = s.oldPosition
}

func (s *state) tokenTemplate() LexicalToken {
	return LexicalToken{Line: s.tokenLine, Position: s.tokenPosition, Value: s.buffer.String(), File: s.file}
}

func (s *state) lexIdentifierOrKeyword() {
	for {
		s.advance()
		if s.err != nil {
			s.lexIdentifierOrKeywordEnd()
			return
		}
		if unicode.IsLetter(s.r) || unicode.IsDigit(s.r) || s.r == '_' {
			s.buffer.WriteRune(s.r)
		} else {
			s.lexIdentifierOrKeywordEnd()
			s.fallBack()
			return
		}
	}
}

func (s *state) lexIdentifierOrKeywordEnd() {
	t := s.tokenTemplate()
	if isKeyword(t.Value) {
		t.TokenType = Keyword
	} else {
		t.TokenType = Identifier
	}
	s.tokens = append(s.tokens, t)
}

func (s *state) lexInteger() {
	for {
		s.advance()
		if s.err != nil {
			s.lexIntegerEnd()
			return
		}
		if unicode.IsDigit(s.r) {
			s.buffer.WriteRune(s.r)
		} else {
			s.lexIntegerEnd()
			s.fallBack()
			return
		}
	}
}

func (s *state) lexIntegerEnd() {
	t := s.tokenTemplate()
	t.TokenType = Integer
	s.tokens = append(s.tokens, t)
}

func (s *state) lexString() {
	escape := false
	for {
		s.advance()
		if s.err != nil {
			if s.err == io.EOF {
				s.err = errors.New("expected \" at Line " + strconv.Itoa(s.line) + " Position " + strconv.Itoa(s.position) + " at File " + s.file)
			}
			return
		}
		if escape {
			if s.r == '"' {
				s.buffer.WriteRune(s.r)
			} else if s.r == 'n' {
				s.buffer.WriteRune('\n')
			} else if s.r == '\\' {
				s.buffer.WriteRune(s.r)
			} else {
				s.buffer.WriteRune('\\')
				s.buffer.WriteRune(s.r)
			}
			escape = false
		} else if s.r == '\\' {
			escape = true
		} else if s.r == '"' {
			s.lexStringEnd()
			return
		} else {
			s.buffer.WriteRune(s.r)
		}
	}
}

func (s *state) lexStringEnd() {
	t := s.tokenTemplate()
	t.TokenType = Text
	s.tokens = append(s.tokens, t)
}

func (s *state) lexOperator() {
	for {
		s.advance()
		if s.err != nil {
			s.lexOperatorEnd()
			return
		}
		if isOperator(string(s.r)) && isOperator(s.buffer.String()+string(s.r)) {
			s.buffer.WriteRune(s.r)
		} else {
			s.lexOperatorEnd()
			s.fallBack()
			return
		}
	}
}

func (s *state) lexOperatorEnd() {
	t := s.tokenTemplate()
	t.TokenType = Operator
	s.tokens = append(s.tokens, t)
}
