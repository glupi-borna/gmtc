package parser

import (
	"errors"
	"fmt"
	. "gmtc/utils"
	"slices"
	"strings"
)

type Location struct {
	Line, Char, Index int
}

type TOKEN_TYPE int

//go:generate stringer -type=TOKEN_TYPE -trimprefix=T_
const (
	T_UNKNOWN TOKEN_TYPE = iota

	T_IDENT
	T_NUMBER
	T_STRING

	T_DOT
	T_SEMICOLON
	T_COMMA
	T_QUESTION
	T_COLON
	T_EXCLAM

	T_ACC_LIST
	T_ACC_MAP
	T_ACC_GRID
	T_ACC_ARRAY
	T_ACC_STRUCT

	T_LPAREN
	T_RPAREN
	T_LSQUARE
	T_RSQUARE
	T_LCURLY
	T_RCURLY

	T_LEQ
	T_GEQ
	T_EQ
	T_NEQ
	T_LESS
	T_MORE

	T_PLUS
	T_MINUS
	T_DIV
	T_INTDIV
	T_MUL
	T_MOD

	T_OR
	T_AND
	T_NULLISH

	T_BITOR
	T_BITAND
	T_BITNOT
	T_BITXOR
	T_LSHIFT
	T_RSHIFT

	T_ASSIGN
	T_ASSIGN_ADD
	T_ASSIGN_SUB
	T_ASSIGN_MUL
	T_ASSIGN_DIV
	T_ASSIGN_OR
	T_ASSIGN_AND
	T_ASSIGN_NULLISH

	T_DECREMENT
	T_INCREMENT

	T_HASH
	T_BACKSLASH
	T_NEWLINE

	T_EOF
)

func (t TOKEN_TYPE) IsAny(types []TOKEN_TYPE) bool {
	for _, tt := range types {
		if t == tt { return true }
	}
	return false
}

type LToken struct {
	Value string
	Type  TOKEN_TYPE
}

var literal_tokens = []LToken{
	{"\r\n", T_NEWLINE},

	{"??=", T_ASSIGN_NULLISH},
	{"[|", T_ACC_LIST},
	{"[?", T_ACC_MAP},
	{"[#", T_ACC_GRID},
	{"[@", T_ACC_ARRAY},
	{"[$", T_ACC_STRUCT},
	{"<=", T_LEQ},
	{">=", T_GEQ},
	{"==", T_EQ},
	{"!=", T_NEQ},
	{"||", T_OR},
	{"&&", T_AND},
	{"+=", T_ASSIGN_ADD},
	{"-=", T_ASSIGN_SUB},
	{"*=", T_ASSIGN_MUL},
	{"/=", T_ASSIGN_DIV},
	{"|=", T_ASSIGN_OR},
	{"&=", T_ASSIGN_AND},
	{"??", T_NULLISH},

	{"--", T_DECREMENT},
	{"++", T_INCREMENT},

	{"|", T_BITOR},
	{"&", T_BITAND},
	{"~", T_BITNOT},
	{"^", T_BITXOR},
	{"<<", T_LSHIFT},
	{">>", T_RSHIFT},

	{".", T_DOT},
	{"?", T_QUESTION},
	{":", T_COLON},
	{"!", T_EXCLAM},
	{"(", T_LPAREN},
	{")", T_RPAREN},
	{"[", T_LSQUARE},
	{"]", T_RSQUARE},
	{"{", T_LCURLY},
	{"}", T_RCURLY},
	{"<", T_LESS},
	{">", T_MORE},
	{"+", T_PLUS},
	{"-", T_MINUS},
	{"*", T_MUL},
	{"%", T_MOD},
	{"/", T_DIV},
	{"=", T_ASSIGN},
	{";", T_SEMICOLON},
	{",", T_COMMA},

	{"#", T_HASH},
	{"\\", T_BACKSLASH},
	{"\n", T_NEWLINE},
}

type TOKEN_FLAG int

const (
	TF_DOT TOKEN_FLAG = 1 << iota
	TF_HEX
	TF_HEX_DOLLAR
	TF_HEX_HASH
)

type Token struct {
	Type  TOKEN_TYPE
	Value string
	Loc   Location
	Flags TOKEN_FLAG
}

func (t Token) IsAny(strs ...string) bool {
	if t.Type != T_IDENT { return false }
	for _, s := range strs {
		if t.Value == s { return true }
	}
	return false
}

type Tokens []Token

func (t Token) String() string {
	switch t.Type {
	case T_NUMBER:
		return fmt.Sprintf("NUM<%v>", t.Value)
	case T_STRING:
		return fmt.Sprintf("STR<%v>", t.Value)
	case T_IDENT:
		return fmt.Sprintf("IDENT<%v>", t.Value)
	case T_NEWLINE:
		return "NL<\\n>"
	default:
		return fmt.Sprintf("%v<%v>", t.Type, t.Value)
	}
}

type scanner struct {
	Text  string
	Saved Location
	Loc   Location
}

const EOF = byte(0)

func (s *scanner) save() {
	s.Saved = s.Loc
}

func (s *scanner) restore() {
	s.Loc = s.Saved
}

func (s *scanner) startsWith(substr string) bool {
	return strings.HasPrefix(s.Text[s.Loc.Index:], substr)
}

func (s *scanner) token(type_ TOKEN_TYPE) Token {
	t := Token{
		Type:  type_,
		Value: s.Text[s.Saved.Index:s.Loc.Index],
		Loc:   s.Saved,
	}
	s.save()
	return t
}

func (s *scanner) char() byte {
	if s.Loc.Index >= len(s.Text) {
		return EOF
	}
	// Multibyte compatibility
	ch := s.Text[s.Loc.Index]
	if ch > 127 {
		return 'a'
	}
	return ch
}

func (s *scanner) move() {
	s.Loc.Index++
	if s.char() == '\n' {
		s.Loc.Char = 0
		s.Loc.Line++
	} else {
		s.Loc.Char++
	}
}

func (s *scanner) eatBlockComments() bool {
	moved := false
	for {
		if s.startsWith("/*") {
			for s.char() != EOF && !s.startsWith("*/") {
				s.move()
				moved = true
			}
			if s.startsWith("*/") {
				s.move()
				s.move()
				break
			}
		}
		if !moved {
			break
		}
	}

	return moved
}

func (s *scanner) eatLineComments() bool {
	moved := false
	for {
		if !s.startsWith("//") {
			break
		}
		for s.char() != EOF && s.char() != '\n' {
			s.move()
			moved = true
		}
	}
	return moved
}

func (s *scanner) eatRegions() bool {
	moved := false
	for {
		if !s.startsWith("#region") && !s.startsWith("#endregion") {
			break
		}
		for s.char() != EOF && s.char() != '\n' {
			s.move()
			moved = true
		}
	}
	return moved
}

func (s *scanner) eatWhitespace() bool {
	has_moved := false
	for IsWhitespaceNoNL(s.char()) {
		s.move()
		has_moved = true
	}
	return has_moved
}

func (s *scanner) skip() {
	for {
		if !s.eatWhitespace() &&
			!s.eatLineComments() &&
			!s.eatBlockComments() &&
			!s.eatRegions() {
			return
		}
	}
}

func (s *scanner) parseIdent() Token {
	s.save()
	i := 0
	for IsIdentChar(s.char(), i) {
		s.move()
		i++
	}
	return s.token(T_IDENT)
}

func (s *scanner) parseNumber() (Token, error) {
	s.save()
	dot := false
	first_char := s.char()
	dollar_hex := first_char == '$'
	hash_hex := first_char == '#'
	ox_hex := s.startsWith("0x")
	hex := ox_hex || dollar_hex || hash_hex

	if hex {
		s.move()
	}

	if ox_hex {
		s.move()
	}

	if hex {
		for {
			char := s.char()
			if IsHexNumber(char) {
				s.move()
				continue
			}
			break
		}
	} else {
		for {
			char := s.char()
			if IsNumber(char) {
				s.move()
				continue
			}

			if !dot && char == '.' {
				dot = true
				s.move()
				continue
			}

			break
		}
	}

	loc := s.Saved
	t := s.token(T_NUMBER)

	if t.Value == "$" || t.Value == "0x" || t.Value == "." || t.Value == "#" {
		s.Saved = loc
		s.restore()
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
		if hash_hex {
			t.Flags |= TF_HEX_HASH
		}
	}

	return t, nil
}

func (s *scanner) parseLiteralToken() (Token, error) {
	s.save()

	for _, lt := range literal_tokens {
		if s.startsWith(lt.Value) {
			for i := 0; i < len(lt.Value); i++ {
				s.move()
			}
			return s.token(lt.Type), nil
		}
	}

	return Token{}, errors.New("No literal token found")
}

func (s *scanner) parseString() (Token, error) {
	first := s.char()
	if first != '"' && first != '\'' {
		return Token{}, errors.New("Invalid string starter!")
	}

	s.move()
	s.save()

	for s.char() != first {
		if s.char() == EOF {
			return Token{}, fmt.Errorf("Unterminated string literal starting at %v:%v", s.Saved.Line+1, s.Saved.Char+1)
		}
		if s.char() == '\\' {
			s.move()
		}
		s.move()
	}

	t := s.token(T_STRING)
	s.move()
	return t, nil
}

func (s *scanner) tokenize() (Tokens, error) {
	tokens := make(Tokens, 0, max(len(s.Text) / 10, 128))

	// last_loc is just a safeguard while the tokenizer is being developed.
	// It should be possible to remove it.
	last_loc := -1
	for {
		if last_loc == s.Loc.Index {
			return nil, fmt.Errorf("%v:%v: Stuck on '%v'", s.Loc.Line+1, s.Loc.Char+1, string(s.char()))
		}
		last_loc = s.Loc.Index
		s.skip()
		next_char := s.char()
		if next_char == EOF {
			break
		}

		if IsIdentChar(next_char, 0) {
			tok := s.parseIdent()
			tokens = append(tokens, tok)
			continue
		}

		if IsNumber(next_char) || next_char == '.' || next_char == '$' || next_char == '#' {
			tok, err := s.parseNumber()
			if err == nil {
				tokens = append(tokens, tok)
				continue
			}
		}

		if next_char == '"' || next_char == '\'' {
			tok, err := s.parseString()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, tok)
			continue
		}

		tok, err := s.parseLiteralToken()
		if err == nil {
			tokens = append(tokens, tok)
			continue
		}

		if s.startsWith("@'") || s.startsWith("@\"") {
			s.move()
			tok, err := s.parseString()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, tok)
			continue
		}

		return nil, fmt.Errorf("Unexpected character at %v:%v: %v (%v)", s.Loc.Line+1, s.Loc.Char+1, string(next_char), next_char)
	}

	tokens = append(tokens, Token{Type: T_EOF})

	return tokens, nil
}

func Pretokenize(text string) (Tokens, error) {
	s := scanner{Text: text}
	return s.tokenize()
}

func (ts Tokens) MatchTypeAt(index int, tt TOKEN_TYPE) bool {
	if index < 0 || index >= len(ts) {
		return false
	}
	return ts[index].Type == tt
}

func (ts Tokens) MatchValueAt(index int, value string) bool {
	if index < 0 || index >= len(ts) {
		return false
	}
	return ts[index].Value == value
}

type Macro struct {
	Name, Config    string
	Value           Tokens
	RawTokensLength int
}

func (ts Tokens) ExtractMacros() map[string]Macro {
	out := make(map[string]Macro)

	for i := range ts {
		macro_start := i
		if !ts.MatchTypeAt(i, T_HASH) {
			continue
		}
		if !ts.MatchValueAt(i+1, "macro") {
			continue
		}
		if !ts.MatchTypeAt(i+2, T_IDENT) {
			continue
		}

		macro_name := ""
		macro_config := ""
		if ts.MatchTypeAt(i+3, T_COLON) && ts.MatchTypeAt(i+4, T_IDENT) {
			// #macro platform:name abc
			macro_name = ts[i+4].Value
			macro_config = ts[i+2].Value
			i += 5
		} else {
			// #macro name abc
			macro_name = ts[i+2].Value
			i += 3
		}

		macro_value := Tokens{}
		macro_newlines := 0
		for {
			if ts[i].Type == T_NEWLINE || ts[i].Type == T_EOF {
				break
			}
			for {
				if ts[i].Type == T_BACKSLASH {
					i++
					continue
				}
				if ts[i].Type == T_NEWLINE {
					macro_newlines++
					i++
					continue
				}
				break
			}
			macro_value = append(macro_value, ts[i])
			i++
		}

		out[macro_name] = Macro{
			Name:            macro_name,
			Config:          macro_config,
			Value:           macro_value,
			RawTokensLength: i - macro_start,
		}
	}

	return out
}

func (ts Tokens) InsertMacros(macros map[string]Macro) Tokens {
	for i := 0; i < len(ts); i++ {
		if ts[i].Type != T_IDENT {
			continue
		}
		macro, ok := macros[ts[i].Value]
		if !ok {
			continue
		}
		if ts.MatchTypeAt(i-2, T_HASH) || ts.MatchTypeAt(i-4, T_HASH) {
			i += macro.RawTokensLength
			continue
		}
		ts = slices.Replace(ts, i, i+1, macro.Value...)
	}
	return ts
}

func (ts Tokens) Clean(macros map[string]Macro) Tokens {
	for i := 0; i < len(ts); i++ {
		if ts[i].Type == T_NEWLINE {
			// fmt.Println(ts[i].Loc.Line+1, "Remove newline", ts[i])
			ts = slices.Delete(ts, i, i+1)
			i--

		} else if ts[i].Type == T_HASH && ts.MatchValueAt(i+1, "macro") {
			macro_token := ts[i+2]
			if ts.MatchTypeAt(i+3, T_COLON) {
				macro_token = ts[i+4]
			}
			macro, ok := macros[macro_token.Value]
			if !ok {
				continue
			}
			// fmt.Println(ts[i].Loc.Line+1, "Remove macro", ts[i])
			ts = slices.Delete(ts, i, i+macro.RawTokensLength)
			i--
		} else {
			// fmt.Println(ts[i].Loc.Line+1, "Skip", ts[i])

		}
	}

	for _, t := range ts {
		if t.Type == T_BACKSLASH { panic(fmt.Sprint(t, t.Loc.Line+1)) }
		if t.Type == T_NEWLINE { panic(fmt.Sprint(t, t.Loc.Line+1)) }
		if t.Type == T_HASH { panic(fmt.Sprint(t, t.Loc.Line+1)) }
	}

	return ts
}

func TokenizeString(text string) (Tokens, error) {
	ts, err := Pretokenize(text)
	if err != nil {
		return nil, err
	}
	macros := ts.ExtractMacros()
	ts = ts.InsertMacros(macros).Clean(macros)
	return ts, nil
}
