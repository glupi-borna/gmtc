package ast

import (
	"errors"
	p "gmtc/parser"
	u "gmtc/utils"
)

type Scanner struct {
	Tokens p.Tokens
	Index  int
	Saved  []int
}

func (ts *Scanner) At(offset int) *p.Token {
	if ts.Index+offset < 0 {
		return nil
	}
	if ts.Index+offset >= len(ts.Tokens) {
		return nil
	}
	return &ts.Tokens[ts.Index+offset]
}

func (ts *Scanner) Move(amount int) {
	ts.Index += amount
}

func (ts *Scanner) Save() {
	ts.Saved = append(ts.Saved, ts.Index)
}

func (ts *Scanner) Restore() {
	last := len(ts.Tokens) - 1
	if last < 0 {
		panic("Empty restore stack!")
	}
	ts.Index = ts.Saved[last]
	ts.Saved = ts.Saved[:last]
}

func (ts *Scanner) Commit() {
	last := len(ts.Tokens) - 1
	if last < 0 {
		panic("Empty restore stack!")
	}
	ts.Saved = ts.Saved[:last]
}

type AST_TYPE int

//go:generate stringer -type=AST_TYPE -trimprefix=AST_
const (
	AST_UNKNOWN AST_TYPE = iota

	AST_SCRIPT

	AST_LITERAL_STRING
	AST_LITERAL_NUMBER
	AST_LITERAL_BOOL
	AST_LITERAL_ARRAY
	AST_LITERAL_STRUCT

	AST_STRUCT_FIELD

	AST_BINOP
	AST_UNOP
	AST_CALL
	AST_ACCESS
	AST_ATTR

	AST_OPERATOR

	AST_BLOCK

	AST_FUNC_DECL
	AST_ARG
)

type Node interface {
	Start() p.Location
	End() p.Location
}

type Base struct {
	Type AST_TYPE
}

type ScriptNode struct {
	Base
	Children []Statement
}

type Statement interface {
	Node
}

func ParseAST(tokens p.Tokens) ScriptNode {
	if len(tokens) == 0 {
		return ScriptNode{
			Base: Base{AST_SCRIPT},
		}
	}

	ts := Scanner{Tokens: tokens}

	stmts := ts.ParseStatements()

	sn := ScriptNode{
		Base:     Base{AST_SCRIPT},
		Children: stmts,
	}

	return sn
}

func (ts *Scanner) ParseExact(offset int, value string) *p.Token {
	t := ts.At(offset)
	if t == nil {
		return nil
	}
	if t.Value == value {
		return t
	}
	return nil
}

func (ts *Scanner) ParseType(offset int, tt p.TOKEN_TYPE) *p.Token {
	t := ts.At(offset)
	if t == nil {
		return nil
	}
	if t.Type == tt {
		return t
	}
	return nil
}

func (ts *Scanner) ParseStatements() []Statement {
	stmts := make([]Statement, 0)
	for {
		stmt := ts.ParseStatement()
		if stmt == nil {
			break
		}
		stmts = append(stmts, stmt)
	}
	return stmts
}

func (ts *Scanner) ParseStatement() Statement {
	if stmt, err := ts.ParseFuncDecl(); err == nil {
		return &stmt
	}
	return nil
}

type Block struct {
	Base
	openCurly, closeCurly *p.Token
	Statements            []Statement
}

func (b *Block) Start() p.Location { return b.openCurly.Loc }
func (b *Block) End() p.Location   { return b.closeCurly.Loc }

func (ts *Scanner) ParseBlock() (Block, error) {
	ts.Save()

	opening := ts.ParseType(0, p.T_LCURLY)
	if opening == nil {
		ts.Restore()
		return Block{}, errors.New("Missing opening curly brace")
	}
	ts.Move(1)

	stmts := ts.ParseStatements()
	closing := ts.ParseType(0, p.T_RCURLY)

	if closing == nil {
		ts.Restore()
		return Block{}, errors.New("Missing closing curly brace")
	}
	ts.Move(1)

	ts.Commit()
	return Block{
		Base:       Base{AST_BLOCK},
		openCurly:  opening,
		closeCurly: closing,
		Statements: stmts,
	}, nil
}

type FuncDecl struct {
	Base
	keyword *p.Token
	Name    *p.Token
	Args    []Arg
	Body    Block
}

func (fn *FuncDecl) Start() p.Location {
	return fn.keyword.Loc
}

func (fn *FuncDecl) End() p.Location {
	return fn.Body.End()
}

func (ts *Scanner) ParseFuncDecl() (FuncDecl, error) {
	ts.Save()
	kwd := ts.ParseExact(0, "function")
	name := ts.ParseType(1, p.T_IDENT)
	if u.AnyNil(kwd, name) {
		return FuncDecl{}, errors.New("Not a function node")
	}
	ts.Move(2)
	args, err := ts.ParseArgs()
	if err != nil {
		ts.Restore()
		return FuncDecl{}, err
	}

	body, err := ts.ParseBlock()
	if err != nil {
		ts.Restore()
		return FuncDecl{}, err
	}

	ts.Commit()
	return FuncDecl{
		Base:    Base{AST_FUNC_DECL},
		keyword: kwd,
		Name:    name,
		Args:    args,
		Body:    body,
	}, nil
}

func (ts *Scanner) ParseArgs() ([]Arg, error) {
	ts.Save()

	opening := ts.ParseType(0, p.T_LPAREN)
	if opening == nil {
		ts.Restore()
		return nil, errors.New("Missing opening parenthesis")
	}
	ts.Move(1)

	args := make([]Arg, 0)
	for {
		arg, err := ts.ParseArg()
		if err != nil {
			break
		}
		args = append(args, arg)

		comma := ts.ParseType(0, p.T_COMMA)
		if comma == nil {
			break
		}
		ts.Move(1)
	}

	closing := ts.ParseType(0, p.T_RPAREN)
	if closing == nil {
		ts.Restore()
		return nil, errors.New("Missing closing parenthesis")
	}
	ts.Move(1)

	ts.Commit()
	return args, nil
}

type Arg struct {
	Base
	Name    *p.Token
	Default Node
}

func (a *Arg) Start() p.Location { return a.Name.Loc }
func (a *Arg) End() p.Location {
	if a.Default != nil {
		return a.Default.End()
	}
	return a.Name.Loc
}

func (ts *Scanner) ParseArg() (Arg, error) {
	ts.Save()

	name := ts.ParseType(0, p.T_IDENT)
	if name == nil {
		return Arg{}, errors.New("Missing argument name")
	}
	ts.Move(1)

	var def Node = nil
	var err error = nil

	eq := ts.ParseType(0, p.T_ASSIGN)
	if eq != nil {
		ts.Move(1)
		def, err = ts.ParseExpr(false)
		if err != nil {
			ts.Restore()
			return Arg{}, err
		}
	}

	ts.Commit()
	return Arg{
		Base:    Base{AST_ARG},
		Name:    name,
		Default: def,
	}, nil
}

func (ts *Scanner) ParseExpr(no_prefix bool) (Node, error) {
	if !no_prefix {
		attr, err := ts.ParseAttr()
		if err == nil {
			return attr, nil
		}

		binop, err := ts.ParseBinop()
		if err == nil {
			return binop, nil
		}

		call, err := ts.ParseCall()
		if err == nil {
			return call, nil
		}

		acc, err := ts.ParseAccessor()
		if err == nil {
			return acc, nil
		}
	}

	unop, err := ts.ParseUnop()
	if err == nil {
		return unop, nil
	}

	lit, err := ts.ParseLiteral()
	if err == nil {
		return lit, nil
	}

	return nil, errors.New("Failed to parse expression")
}

type Simple struct {
	Base
	Value *p.Token
}

func (s *Simple) Start() p.Location { return s.Value.Loc }
func (s *Simple) End() p.Location   { return s.Value.Loc }

func (ts *Scanner) ParseLiteral() (Node, error) {
	num := ts.ParseType(0, p.T_NUMBER)
	if num != nil {
		ts.Move(1)
		ts.Commit()
		return &Simple{Base{AST_LITERAL_NUMBER}, num}, nil
	}

	str := ts.ParseType(0, p.T_STRING)
	if num != nil {
		ts.Move(1)
		ts.Commit()
		return &Simple{Base{AST_LITERAL_STRING}, str}, nil
	}

	ident := ts.ParseType(0, p.T_IDENT)
	if ident != nil {
		btrue := ts.ParseExact(0, "true")
		if btrue != nil {
			return &Simple{Base{AST_LITERAL_BOOL}, btrue}, nil
		}

		bfalse := ts.ParseExact(0, "false")
		if bfalse != nil {
			return &Simple{Base{AST_LITERAL_BOOL}, bfalse}, nil
		}
	}

	arr, err := ts.ParseArray()
	if err != nil {
		return &arr, nil
	}

	strct, err := ts.ParseStruct()
	if err != nil {
		return &strct, nil
	}

	return nil, errors.New("Failed to parse literal")
}

type Array struct {
	Base
	openSquare, closeSquare *p.Token
	Items                   []Node
}

func (b *Array) Start() p.Location { return b.openSquare.Loc }
func (b *Array) End() p.Location   { return b.closeSquare.Loc }

func (ts *Scanner) ParseArray() (Array, error) {
	ts.Save()

	opening := ts.ParseType(0, p.T_LSQUARE)
	if opening == nil {
		ts.Restore()
		return Array{}, errors.New("Missing opening square brace")
	}
	ts.Move(1)

	items := make([]Node, 0)
	for {
		item, err := ts.ParseExpr(false)
		if err != nil {
			break
		}
		items = append(items, item)

		comma := ts.ParseType(0, p.T_COMMA)
		if comma == nil {
			break
		}
		ts.Move(1)
	}

	closing := ts.ParseType(0, p.T_RSQUARE)
	if closing == nil {
		ts.Restore()
		return Array{}, errors.New("Missing closing square brace")
	}
	ts.Move(1)

	ts.Commit()
	return Array{
		Base:        Base{AST_LITERAL_ARRAY},
		openSquare:  opening,
		closeSquare: closing,
		Items:       items,
	}, nil
}

type Struct struct {
	Base
	openCurly, closeCurly *p.Token
	Fields                []Field
}

func (b *Struct) Start() p.Location { return b.openCurly.Loc }
func (b *Struct) End() p.Location   { return b.closeCurly.Loc }

func (ts *Scanner) ParseStruct() (Struct, error) {
	ts.Save()

	opening := ts.ParseType(0, p.T_LCURLY)
	if opening == nil {
		ts.Restore()
		return Struct{}, errors.New("Missing opening curly brace")
	}
	ts.Move(1)

	fields := make([]Field, 0)
	for {
		field, err := ts.ParseField()
		if err != nil {
			break
		}
		fields = append(fields, field)

		comma := ts.ParseType(0, p.T_COMMA)
		if comma == nil {
			break
		}
		ts.Move(1)
	}

	closing := ts.ParseType(0, p.T_RCURLY)
	if closing == nil {
		ts.Restore()
		return Struct{}, errors.New("Missing closing curly brace")
	}
	ts.Move(1)

	ts.Commit()
	return Struct{
		Base:       Base{AST_LITERAL_STRING},
		openCurly:  opening,
		closeCurly: closing,
		Fields:     fields,
	}, nil
}

type Field struct {
	Base
	Name  *p.Token
	Value Node
}

func (f *Field) Start() p.Location { return f.Name.Loc }
func (f *Field) End() p.Location {
	if f.Value != nil {
		return f.Value.End()
	}
	return f.Name.Loc
}

func (ts *Scanner) ParseField() (Field, error) {
	ts.Save()

	name := ts.ParseType(0, p.T_IDENT)
	if name == nil {
		return Field{}, errors.New("Missing field name")
	}
	ts.Move(1)

	var val Node = nil
	var err error = nil

	colon := ts.ParseType(0, p.T_COLON)
	if colon != nil {
		ts.Move(1)
		val, err = ts.ParseExpr(false)
		if err != nil {
			ts.Restore()
			return Field{}, err
		}
	}

	ts.Commit()
	return Field{
		Base:  Base{AST_STRUCT_FIELD},
		Name:  name,
		Value: val,
	}, nil
}

type Call struct {
	Base
	openParen  *p.Token
	closeParen *p.Token
	Function   Node
	Params     []Node
}

func (c *Call) Start() p.Location { return c.Function.Start() }
func (c *Call) End() p.Location   { return c.closeParen.Loc }

func (ts *Scanner) ParseCall() (Call, error) {
	ts.Save()

	fn, err := ts.ParseExpr(true)
	if err != nil { return nil }

	opening := ts.ParseType(0, p.T_LPAREN)
	if opening == nil {
		ts.Restore()
		return Call{}, errors.New("Missing opening parenthesis")
	}
	ts.Move(1)

	params := make([]Node, 0)
	for {
		param, err := ts.ParseExpr(false)
		if err != nil {
			break
		}
		params = append(params, param)

		comma := ts.ParseType(0, p.T_COMMA)
		if comma == nil {
			break
		}
		ts.Move(1)
	}


	closing := ts.ParseType(0, p.T_RPAREN)
	if closing == nil {
		ts.Restore()
		return Call{}, errors.New("Missing closing parenthesis")
	}

	ts.Move(1)

	ts.Commit()
	return Call{
		Base: Base{AST_CALL},
		openParen: opening,
		closeParen: closing,
		Function: fn,
		Params: params,
	}, nil
}
