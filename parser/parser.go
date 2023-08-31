package parser

import (
	"errors"
	"fmt"
	"strings"
)

type Location struct {
	Line, Char, Index int
}

type TOKEN_TYPE int

//go:generate stringer -type=TOKEN_TYPE -trimprefix=T_
const (
	T_UNKNOWN TOKEN_TYPE = 0

	T_IDENT TOKEN_TYPE = iota
	T_NUMBER
	T_STRING

	T_DOT
	T_SEMICOLON
	T_COMMA

	T_LPAREN
	T_RPAREN
	T_LSQUARE
	T_RSQUARE
	T_LCURLY
	T_RCURLY

	T_LEQ
	T_GEQ
	T_EQ
	T_LESS
	T_MORE

	T_PLUS
	T_MINUS
	T_DIV
	T_MUL

	T_OR
	T_AND

	T_ASSIGN
)

type LToken struct {
	Value string
	Type  TOKEN_TYPE
}

var literal_tokens = []LToken{
	{".", T_DOT},
	{"(", T_LPAREN},
	{")", T_RPAREN},
	{"[", T_LSQUARE},
	{"]", T_RSQUARE},
	{"{", T_LCURLY},
	{"}", T_RCURLY},
	{"<=", T_LEQ},
	{">=", T_GEQ},
	{"==", T_EQ},
	{"<", T_LESS},
	{">", T_MORE},
	{"+", T_PLUS},
	{"-", T_MINUS},
	{"*", T_MUL},
	{"/", T_DIV},
	{"||", T_OR},
	{"&&", T_AND},
	{"=", T_ASSIGN},
	{";", T_SEMICOLON},
	{",", T_COMMA},
}

type TOKEN_FLAG int

const (
	TF_DOT TOKEN_FLAG = 1 << iota
	TF_HEX
	TF_HEX_DOLLAR
)

type Token struct {
	Type  TOKEN_TYPE
	Value string
	Loc   Location
	Flags TOKEN_FLAG
}

func (t Token) String() string {
	switch t.Type {
	case T_NUMBER:
		return fmt.Sprintf("NUM<%v>", t.Value)
	case T_STRING:
		return fmt.Sprintf("STR<%v>", t.Value)
	case T_IDENT:
		return fmt.Sprintf("IDENT<%v>", t.Value)
	default:
		return fmt.Sprintf("TOK<%v, %v>", t.Type, t.Value)
	}
}

type Scanner struct {
	Text  string
	Saved Location
	Loc   Location
}

const EOF = byte(0)

func (s *Scanner) Save() {
	s.Saved = s.Loc
}

func (s *Scanner) Restore() {
	s.Loc = s.Saved
}

func (s *Scanner) StartsWith(substr string) bool {
	return strings.HasPrefix(s.Text[s.Loc.Index:], substr)
}

func (s *Scanner) Token(type_ TOKEN_TYPE) Token {
	t := Token{
		Type:  type_,
		Value: s.Text[s.Saved.Index:s.Loc.Index],
		Loc:   s.Saved,
	}
	s.Save()
	return t
}

func (s *Scanner) Char() byte {
	if s.Loc.Index >= len(s.Text) {
		return EOF
	}
	return s.Text[s.Loc.Index]
}

func (s *Scanner) Move() {
	s.Loc.Index += 1
	if s.Char() == '\n' {
		s.Loc.Char = 0
		s.Loc.Line++
	} else {
		s.Loc.Char++
	}
}

func IsBetween(char byte, start byte, end byte) bool {
	return char >= start && char <= end
}

func IsWhitespace(char byte) bool {
	return char == ' ' || char == '\n' || char == '\t' || char == '\r'
}

func IsHexNumber(char byte) bool {
	return IsNumber(char) || IsBetween(char, 'a', 'f') || IsBetween(char, 'A', 'f')
}

func IsNumber(char byte) bool {
	return IsBetween(char, '0', '9')
}

func IsLetter(char byte) bool {
	return IsBetween(char, 'A', 'Z') || IsBetween(char, 'a', 'z')
}

func IsIdentChar(char byte, i int) bool {
	if IsLetter(char) || char == '_' {
		return true
	}
	return i != 0 && IsNumber(char)
}

func (s *Scanner) EatComments() bool {
	has_moved := false
	for {
		moved := false
		if s.StartsWith("//") {
			for s.Char() != '\n' && s.Char() != EOF {
				s.Move()
				moved = true
				has_moved = true
			}
		}

		if s.StartsWith("/*") {
			for s.Char() != EOF && !s.StartsWith("*/") {
				s.Move()
				moved = true
				has_moved = true
			}
			if s.StartsWith("*/") {
				s.Move()
				s.Move()
			}
		}

		if !moved {
			break
		}
	}
	return has_moved
}

func (s *Scanner) EatWhitespace() bool {
	has_moved := false
	for IsWhitespace(s.Char()) {
		s.Move()
		has_moved = true
	}
	return has_moved
}

func (s *Scanner) SkipCommentsAndWS() {
	for {
		ws := s.EatWhitespace()
		comm := s.EatComments()
		if !ws && !comm {
			return
		}
	}
}

func (s *Scanner) ParseIdent() Token {
	s.Save()
	i := 0
	for IsIdentChar(s.Char(), i) {
		s.Move()
		i++
	}
	return s.Token(T_IDENT)
}

func (s *Scanner) ParseNumber() (Token, error) {
	s.Save()
	dot := false
	dollar_hex := s.StartsWith("$")
	ox_hex := !dollar_hex && s.StartsWith("0x")
	hex := ox_hex && dollar_hex

	if hex {
		s.Move()
	}

	if ox_hex {
		s.Move()
	}

	if hex {
		for {
			char := s.Char()
			if IsHexNumber(char) {
				s.Move()
				continue
			}
			break
		}
	} else {
		for {
			char := s.Char()
			if IsNumber(char) {
				s.Move()
				continue
			}

			if !dot && char == '.' {
				dot = true
				s.Move()
			}

			break
		}
	}

	loc := s.Saved
	t := s.Token(T_NUMBER)

	if t.Value == "$" || t.Value == "0x" || t.Value == "." {
		s.Saved = loc
		s.Restore()
		return Token{}, errors.New("Invalid number literal")
	}

	if dot {
		t.Flags |= TF_DOT
	}
	if hex {
		t.Flags |= TF_HEX
		if dollar_hex {
			t.Flags |= TF_HEX_DOLLAR
		}
	}

	return t, nil
}

func (s *Scanner) ParseLiteralToken() (Token, error) {
	s.Save()

	for _, lt := range literal_tokens {
		if s.StartsWith(lt.Value) {
			for i := 0; i < len(lt.Value); i++ {
				s.Move()
			}
			return s.Token(lt.Type), nil
		}
	}

	return Token{}, errors.New("No literal token found")
}

func (s *Scanner) ParseString() (Token, error) {
	first := s.Char()
	if first != '"' && first != '\'' {
		return Token{}, errors.New("Invalid string starter!")
	}

	s.Move()
	s.Save()

	for s.Char() != first {
		if s.Char() == EOF {
			return Token{}, fmt.Errorf("Unterminated string literal starting at %v:%v", s.Saved.Line+1, s.Saved.Char+1)
		}
		if s.Char() == '\\' {
			s.Move()
		}
		s.Move()
	}

	t := s.Token(T_STRING)
	s.Move()
	return t, nil
}

func Tokenize(s *Scanner) ([]Token, error) {
	tokens := []Token{}

	for {
		s.SkipCommentsAndWS()
		next_char := s.Char()
		if next_char == EOF {
			break
		}

		if IsIdentChar(next_char, 0) {
			tok := s.ParseIdent()
			tokens = append(tokens, tok)
			continue
		}

		if IsNumber(next_char) || next_char == '.' || next_char == '$' {
			tok, err := s.ParseNumber()
			if err == nil {
				tokens = append(tokens, tok)
				continue
			}
		}

		if next_char == '"' || next_char == '\'' {
			tok, err := s.ParseString()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, tok)
			continue
		}

		tok, err := s.ParseLiteralToken()
		if err == nil {
			tokens = append(tokens, tok)
			continue
		}

		return nil, fmt.Errorf("Unexpected character at %v:%v: %v", s.Loc.Line+1, s.Loc.Char+1, string(next_char))
	}

	return tokens, nil
}
