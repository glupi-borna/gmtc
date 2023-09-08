package ast

import (
	"errors"
	"fmt"
	p "gmtc/parser"
	"reflect"
	"strings"
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
	last := len(ts.Saved) - 1
	if last < 0 {
		panic("Empty restore stack!")
	}
	ts.Index = ts.Saved[last]
	ts.Saved = ts.Saved[:last]
}

func (ts *Scanner) Commit() {
	last := len(ts.Saved) - 1
	if last < 0 {
		panic("Empty restore stack!")
	}
	ts.Saved = ts.Saved[:last]
}

func (ts *Scanner) GuardStart() int {
	return len(ts.Saved)
}

func (ts *Scanner) GuardEnd(num int) {
	if len(ts.Saved) != num {
		panic("Unbalanced restore stack")
	}
}

type AST_TYPE int

//go:generate stringer -type=AST_TYPE -trimprefix=AST_
const (
	AST_UNKNOWN AST_TYPE = iota

	AST_SCRIPT

	AST_IDENT

	AST_LITERAL_STRING
	AST_LITERAL_NUMBER
	AST_LITERAL_BOOL
	AST_LITERAL_ARRAY
	AST_LITERAL_STRUCT

	AST_STRUCT_FIELD

	AST_BINOP
	AST_UNOP_PREFIX
	AST_UNOP_POSTFIX
	AST_CALL
	AST_ACCESS
	AST_ATTR

	AST_BLOCK

	AST_FUNC_DECL
	AST_ARG

	AST_VAR_DECL
	AST_ASSIGN

	AST_IF
	AST_ELIF
	AST_ELSE

	AST_FOR
	AST_WHILE
	AST_DOUNTIL
	AST_REPEAT

	AST_RETURN
	AST_CONTINUE
	AST_BREAK
)

type NodeBuilder struct {
	b      strings.Builder
	indent int
}

type NodeField struct {
	Name  string
	Value any
}

func (nb *NodeBuilder) Indent() { nb.indent++ }
func (nb *NodeBuilder) Dedent() { nb.indent-- }

func (nb *NodeBuilder) Write(text string) {
	nb.b.WriteString(text)
}

func (nb *NodeBuilder) WriteIndent() {
	nb.b.WriteString(strings.Repeat(".   ", nb.indent))
}

func (nb *NodeBuilder) Newline() { nb.b.WriteByte('\n') }

func (nb *NodeBuilder) WriteLine(text string) {
	nb.WriteIndent()
	nb.Write(text)
	nb.b.WriteByte('\n')
}

func (nb *NodeBuilder) WriteFields(fields ...NodeField) {
	for _, field := range fields {
		nb.WriteIndent()
		nb.Write(".")
		nb.Write(field.Name)

		rv := reflect.ValueOf(field.Value)
		if rv.IsZero() {
			nb.Write(" -")
			nb.Newline()
			continue
		}

		switch v := field.Value.(type) {
		case Node:
			nb.Newline()
			nb.Indent()
			v.Render(nb)
			nb.Dedent()
			break

		case Block:
			nb.Newline()
			nb.Indent()
			v.Render(nb)
			nb.Dedent()
			break

		case []Node:
			nb.Indent()
			for _, n := range v {
				nb.Newline()
				n.Render(nb)
			}
			nb.Dedent()
			break

		case []Statement:
			nb.Indent()
			for _, n := range v {
				nb.Newline()
				n.Render(nb)
			}
			nb.Dedent()
			break

		case []Arg:
			nb.Write("(")
			nb.Write(fmt.Sprint(len(v)))
			nb.Write(")")
			nb.Indent()
			for _, n := range v {
				nb.Newline()
				n.Render(nb)
			}
			nb.Dedent()
			break

		default:
			nb.Newline()
			nb.Indent()
			nb.WriteIndent()
			nb.Write(fmt.Sprintf("%v", v))
			nb.Dedent()
			break
		}

		nb.Newline()
	}
}

func (nb *NodeBuilder) RenderNode(b *Base, fields ...NodeField) {
	nb.WriteIndent()
	nb.Write(b.Type.String())
	nb.Newline()
	nb.WriteFields(fields...)
}

func (nb *NodeBuilder) String() string {
	return nb.b.String()
}

type Node interface {
	Start() p.Location
	End() p.Location
	Render(*NodeBuilder)
}

type Base struct {
	Type AST_TYPE
}

type ScriptNode struct {
	Base
	Children []Statement
}

func (sn *ScriptNode) Start() p.Location {
	if len(sn.Children) > 0 {
		return sn.Children[0].Start()
	}
	return p.Location{}
}

func (sn *ScriptNode) End() p.Location {
	if len(sn.Children) > 0 {
		return sn.Children[len(sn.Children)-1].Start()
	}
	return p.Location{}
}

func (sn *ScriptNode) Render(nb *NodeBuilder) {
	nb.RenderNode(
		&sn.Base,
		NodeField{"Children", sn.Children},
	)
}

type Statement interface {
	Node
}

func ParseAST(tokens p.Tokens) (ScriptNode, error) {
	if len(tokens) == 0 {
		return ScriptNode{
			Base: Base{AST_SCRIPT},
		}, nil
	}

	ts := Scanner{Tokens: tokens}

	stmts := ts.ParseStatements()

	t := ts.At(0)
	if t != nil && t.Type != p.T_EOF {
		return ScriptNode{}, fmt.Errorf("Failed at token %v at %v:%v\n", t, t.Loc.Line+1, t.Loc.Char+1)
	}

	return ScriptNode{
		Base:     Base{AST_SCRIPT},
		Children: stmts,
	}, nil
}

func (ts *Scanner) ParseExact(offset int, value string) *p.Token {
	t := ts.At(offset)
	if t == nil {
		return nil
	}
	if t.Type == p.T_IDENT && t.Value == value {
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

func (ts *Scanner) ParseAnyType(offset int, tts ...p.TOKEN_TYPE) *p.Token {
	t := ts.At(offset)
	if t == nil {
		return nil
	}
	for _, tt := range tts {
		if t.Type == tt {
			return t
		}
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

func (ts *Scanner) EatSemicolon() {
	if ts.ParseType(0, p.T_SEMICOLON) != nil {
		ts.Move(1)
	}
}

func (ts *Scanner) ParseStatement() Statement {
	if stmt, err := ts.ParseAssign(); err == nil {
		ts.EatSemicolon()
		return &stmt
	}

	if stmt, err := ts.ParseVarDecl(); err == nil {
		ts.EatSemicolon()
		return &stmt
	}

	if stmt, err := ts.ParseIfStmt(); err == nil {
		ts.EatSemicolon()
		return &stmt
	}

	if stmt, err := ts.ParseForLoop(); err == nil {
		ts.EatSemicolon()
		return &stmt
	}

	if stmt, err := ts.ParseWhileLoop(); err == nil {
		ts.EatSemicolon()
		return &stmt
	}

	if stmt, err := ts.ParseKwdStmt("return", AST_RETURN, VT_OPTIONAL); err == nil {
		ts.EatSemicolon()
		return &stmt
	}

	if stmt, err := ts.ParseKwdStmt("continue", AST_CONTINUE, VT_NONE); err == nil {
		ts.EatSemicolon()
		return &stmt
	}

	if stmt, err := ts.ParseKwdStmt("break", AST_BREAK, VT_NONE); err == nil {
		ts.EatSemicolon()
		return &stmt
	}

	if stmt, err := ts.ParseFuncDecl(); err == nil {
		ts.EatSemicolon()
		return &stmt
	}

	if expr_stmt, err := ts.ParseExpr(nil); err == nil {
		ts.EatSemicolon()
		return expr_stmt
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

func (b *Block) Render(nb *NodeBuilder) {
	for _, stmt := range b.Statements {
		stmt.Render(nb)
	}
}

func (ts *Scanner) ParseBlock() (Block, error) {
	g := ts.GuardStart()
	defer ts.GuardEnd(g)

	opening := ts.ParseType(0, p.T_LCURLY)
	if opening == nil {
		return Block{}, errors.New("Missing opening curly brace")
	}

	ts.Save()
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

type VarDecl struct {
	Base
	keyword *p.Token
	Name    *p.Token
	Value   Node
}

func (v *VarDecl) Start() p.Location { return v.keyword.Loc }
func (v *VarDecl) End() p.Location {
	if v.Value != nil {
		return v.Value.End()
	}
	return v.Name.Loc
}

func (v *VarDecl) Render(nb *NodeBuilder) {
	nb.RenderNode(
		&v.Base,
		NodeField{"Name", v.Name.Value},
		NodeField{"Value", v.Value},
	)
}

func (ts *Scanner) ParseVarDecl() (VarDecl, error) {
	g := ts.GuardStart()
	defer ts.GuardEnd(g)

	kwd := ts.ParseExact(0, "var")
	name := ts.ParseType(1, p.T_IDENT)

	if kwd == nil || name == nil {
		return VarDecl{}, errors.New("Not a var declaration")
	}

	ts.Save()
	ts.Move(2)

	var value Node = nil
	var err error = nil

	eq := ts.ParseType(0, p.T_ASSIGN)
	if eq != nil {
		ts.Move(1)
		value, err = ts.ParseExpr(nil)

		if err != nil {
			ts.Restore()
			return VarDecl{}, err
		}
	}

	ts.Commit()
	return VarDecl{
		Base:    Base{AST_VAR_DECL},
		Name:    name,
		Value:   value,
		keyword: kwd,
	}, nil
}

type Assign struct{ Binop }

func (ts *Scanner) ParseAssign() (Assign, error) {
	g := ts.GuardStart()
	defer ts.GuardEnd(g)

	ts.Save()

	left, err := ts.ParseExpr(nil)
	if err != nil {
		ts.Restore()
		return Assign{}, err
	}

	eq := ts.ParseAnyType(0,
		p.T_ASSIGN,
		p.T_ASSIGN_ADD,
		p.T_ASSIGN_SUB,
		p.T_ASSIGN_DIV,
		p.T_ASSIGN_MUL,
	)
	if eq == nil {
		ts.Restore()
		return Assign{}, errors.New("Missing assignment operator")
	}

	ts.Move(1)

	right, err := ts.ParseExpr(nil)
	if err != nil {
		ts.Restore()
		return Assign{}, err
	}

	ts.Commit()
	return Assign{
		Binop{
			Base:  Base{AST_ASSIGN},
			Op:    eq,
			Left:  left,
			Right: right,
		},
	}, nil
}

type FuncDecl struct {
	Base
	keyword *p.Token
	Name    *p.Token
	Args    []Arg
	Body    Block
}

func (fn *FuncDecl) Start() p.Location { return fn.keyword.Loc }
func (fn *FuncDecl) End() p.Location   { return fn.Body.End() }

func (fn *FuncDecl) Render(nb *NodeBuilder) {
	nb.RenderNode(
		&fn.Base,
		NodeField{"Name", fn.Name.Value},
		NodeField{"Args", fn.Args},
		NodeField{"Body", fn.Body},
	)
}

func (ts *Scanner) ParseFuncDecl() (FuncDecl, error) {
	g := ts.GuardStart()
	defer ts.GuardEnd(g)

	kwd := ts.ParseExact(0, "function")
	name := ts.ParseType(1, p.T_IDENT)
	if kwd == nil || name == nil {
		return FuncDecl{}, errors.New("Not a function node")
	}

	ts.Save()
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
	g := ts.GuardStart()
	defer ts.GuardEnd(g)

	opening := ts.ParseType(0, p.T_LPAREN)
	if opening == nil {
		return nil, errors.New("Missing opening parenthesis")
	}

	ts.Save()
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
func (a *Arg) Render(nb *NodeBuilder) {
	nb.RenderNode(
		&a.Base,
		NodeField{"Name", a.Name.Value},
		NodeField{"Default", a.Default},
	)
}

func (ts *Scanner) ParseArg() (Arg, error) {
	g := ts.GuardStart()
	defer ts.GuardEnd(g)

	name := ts.ParseType(0, p.T_IDENT)
	if name == nil {
		return Arg{}, errors.New("Missing argument name")
	}

	ts.Save()
	ts.Move(1)

	var def Node = nil
	var err error = nil

	eq := ts.ParseType(0, p.T_ASSIGN)
	if eq != nil {
		ts.Move(1)
		def, err = ts.ParseExpr(nil)
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

func (ts *Scanner) ParseExpr(expr_or_nil Node) (Node, error) {
	var err error
	if expr_or_nil == nil {
		expr_or_nil, err = ts.ParseExprPart()
		if err != nil {
			return nil, err
		}
	}

	expr := expr_or_nil
	if expr == nil {
		panic("Nil expression")
	}

	binop, err := ts.ParseBinop(expr)
	if err == nil {
		return ts.ParseExpr(&binop)
	}

	unop, err := ts.ParseUnopPostfix(expr)
	if err == nil {
		return ts.ParseExpr(&unop)
	}

	attr, err := ts.ParseAttr(expr)
	if err == nil {
		return ts.ParseExpr(attr)
	}

	call, err := ts.ParseCall(expr)
	if err == nil {
		return ts.ParseExpr(call)
	}

	acc, err := ts.ParseAccess(expr)
	if err == nil {
		return ts.ParseExpr(acc)
	}

	return expr, nil
}

func (ts *Scanner) ParseExprPart() (Node, error) {
	g := ts.GuardStart()
	defer ts.GuardEnd(g)

	unop, err := ts.ParseUnop()
	if err == nil {
		return &unop, nil
	}

	lit, err := ts.ParseLiteral()
	if err == nil {
		return lit, nil
	}

	ident, err := ts.ParseIdent()
	if err == nil {
		return &ident, nil
	}

	ts.Save()
	opening := ts.ParseType(0, p.T_LPAREN)
	if opening != nil {
		ts.Move(1)
		val, err := ts.ParseExpr(nil)
		if err == nil {
			closing := ts.ParseType(0, p.T_RPAREN)
			if closing != nil {
				ts.Move(1)
				ts.Commit()
				return val, nil
			} else {
				ts.Restore()
			}
		} else {
			ts.Restore()
		}
	} else {
		ts.Restore()
	}

	return nil, errors.New("Failed to parse expression")
}

type Simple struct {
	Base
	Value *p.Token
}

func (s *Simple) Start() p.Location { return s.Value.Loc }
func (s *Simple) End() p.Location   { return s.Value.Loc }

func (s *Simple) Render(nb *NodeBuilder) {
	nb.RenderNode(
		&s.Base,
		NodeField{"Value", s.Value.Value},
	)
}

func (ts *Scanner) ParseIdent() (Simple, error) {
	ident := ts.ParseType(0, p.T_IDENT)
	if ident == nil {
		return Simple{}, errors.New("Failed to parse ident")
	}
	ts.Move(1)
	return Simple{Base{AST_IDENT}, ident}, nil
}

func (ts *Scanner) ParseLiteral() (Node, error) {
	num := ts.ParseType(0, p.T_NUMBER)
	if num != nil {
		ts.Move(1)
		return &Simple{Base{AST_LITERAL_NUMBER}, num}, nil
	}

	str := ts.ParseType(0, p.T_STRING)
	if str != nil {
		ts.Move(1)
		return &Simple{Base{AST_LITERAL_STRING}, str}, nil
	}

	btrue := ts.ParseExact(0, "true")
	if btrue != nil {
		ts.Move(1)
		return &Simple{Base{AST_LITERAL_BOOL}, btrue}, nil
	}

	bfalse := ts.ParseExact(0, "false")
	if bfalse != nil {
		ts.Move(1)
		return &Simple{Base{AST_LITERAL_BOOL}, bfalse}, nil
	}

	arr, err := ts.ParseArray()
	if err == nil {
		return &arr, nil
	}

	strct, err := ts.ParseStruct()
	if err == nil {
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

func (a *Array) Render(nb *NodeBuilder) {
	nb.RenderNode(
		&a.Base,
		NodeField{"Items", a.Items},
	)
}

func (ts *Scanner) ParseArray() (Array, error) {
	g := ts.GuardStart()
	defer ts.GuardEnd(g)

	opening := ts.ParseType(0, p.T_LSQUARE)
	if opening == nil {
		return Array{}, errors.New("Missing opening square brace")
	}

	ts.Save()
	ts.Move(1)

	items := make([]Node, 0)
	for {
		item, err := ts.ParseExpr(nil)
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

func (s *Struct) Start() p.Location { return s.openCurly.Loc }
func (s *Struct) End() p.Location   { return s.closeCurly.Loc }

func (s *Struct) Render(nb *NodeBuilder) {
	nb.RenderNode(
		&s.Base,
		NodeField{"Fields", s.Fields},
	)
}

func (ts *Scanner) ParseStruct() (Struct, error) {
	g := ts.GuardStart()
	defer ts.GuardEnd(g)

	opening := ts.ParseType(0, p.T_LCURLY)
	if opening == nil {
		return Struct{}, errors.New("Missing opening curly brace")
	}

	ts.Save()
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

func (f *Field) Render(nb *NodeBuilder) {
	nb.RenderNode(
		&f.Base,
		NodeField{"Name", f.Name},
		NodeField{"Value", f.Value},
	)
}

func (ts *Scanner) ParseField() (Field, error) {
	g := ts.GuardStart()
	defer ts.GuardEnd(g)

	name := ts.ParseType(0, p.T_IDENT)
	colon := ts.ParseType(1, p.T_COLON)
	if name == nil {
		return Field{}, errors.New("Missing field name")
	}

	ts.Save()

	var val Node = nil
	var err error = nil

	if colon != nil {
		ts.Move(2)
		val, err = ts.ParseExpr(nil)
		if err != nil {
			ts.Restore()
			return Field{}, err
		}
	} else {
		ts.Move(1)
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

func (c *Call) Render(nb *NodeBuilder) {
	nb.RenderNode(
		&c.Base,
		NodeField{"Function", c.Function},
		NodeField{"Params", c.Params},
	)
}

func (ts *Scanner) ParseCall(fn Node) (Node, error) {
	g := ts.GuardStart()
	defer ts.GuardEnd(g)

	opening := ts.ParseType(0, p.T_LPAREN)
	if opening == nil {
		return nil, errors.New("Missing opening parenthesis")
	}

	ts.Save()
	ts.Move(1)

	params := make([]Node, 0)
	for {
		param, err := ts.ParseExpr(nil)
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
		return nil, errors.New("Missing closing parenthesis")
	}

	ts.Move(1)
	ts.Commit()

	return ts.ParseExpr(&Call{
		Base:       Base{AST_CALL},
		openParen:  opening,
		closeParen: closing,
		Function:   fn,
		Params:     params,
	})
}

type Attr struct {
	Base
	Value Node
	Name  *p.Token
}

func (a *Attr) Start() p.Location { return a.Value.Start() }
func (a *Attr) End() p.Location   { return a.Name.Loc }

func (a *Attr) Render(nb *NodeBuilder) {
	nb.RenderNode(
		&a.Base,
		NodeField{"Value", a.Value},
		NodeField{"Name", a.Name},
	)
}

func (ts *Scanner) ParseAttr(val Node) (Node, error) {
	dot := ts.ParseType(0, p.T_DOT)
	ident := ts.ParseType(1, p.T_IDENT)
	if dot == nil || ident == nil {
		return nil, errors.New("Not an attr")
	}

	ts.Move(2)
	return ts.ParseExpr(&Attr{
		Base:  Base{AST_ATTR},
		Value: val,
		Name:  ident,
	})
}

type Access struct {
	Base
	closingSquare *p.Token
	Type          *p.Token
	Value         Node
	Access        Node
}

func (a *Access) Render(nb *NodeBuilder) {
	nb.RenderNode(
		&a.Base,
		NodeField{"Type", a.Type.Type.String()},
		NodeField{"Value", a.Value},
		NodeField{"Access", a.Access},
	)
}

func (a *Access) Start() p.Location { return a.Value.Start() }
func (a *Access) End() p.Location   { return a.closingSquare.Loc }

func (ts *Scanner) ParseAccess(val Node) (Node, error) {
	g := ts.GuardStart()
	defer ts.GuardEnd(g)

	opening := ts.ParseAnyType(0,
		p.T_ACC_ARRAY,
		p.T_ACC_GRID,
		p.T_ACC_LIST,
		p.T_ACC_MAP,
		p.T_ACC_STRUCT,
		p.T_LSQUARE,
	)
	if opening == nil {
		return nil, errors.New("Missing accessor opening brace")
	}

	ts.Save()
	ts.Move(1)

	access, err := ts.ParseExpr(nil)
	if err != nil {
		ts.Restore()
		return nil, errors.New("Missing access")
	}

	closing := ts.ParseType(0, p.T_LSQUARE)
	if closing == nil {
		ts.Restore()
		return nil, errors.New("Missing closing square brace")
	}

	ts.Move(1)
	ts.Commit()
	return ts.ParseExpr(&Access{
		Base:          Base{AST_ACCESS},
		closingSquare: closing,
		Type:          opening,
		Value:         val,
		Access:        access,
	})
}

type Unop struct {
	Base
	Op    *p.Token
	Value Node
}

func (u *Unop) Render(nb *NodeBuilder) {
	nb.RenderNode(
		&u.Base,
		NodeField{"Operator", u.Op.Type.String()},
		NodeField{"Value", u.Value},
	)
}

func (u *Unop) Start() p.Location { return u.Op.Loc }
func (u *Unop) End() p.Location   { return u.Value.End() }

func (ts *Scanner) ParseUnop() (Unop, error) {
	op := ts.ParseAnyType(0,
		p.T_MINUS,
		p.T_EXCLAM,
		p.T_BITNOT,
		p.T_DECREMENT,
		p.T_INCREMENT,
	)
	if op == nil {
		return Unop{}, errors.New("Missing operator")
	}

	ts.Save()
	ts.Move(1)

	val, err := ts.ParseExpr(nil)
	if err != nil {
		ts.Restore()
		return Unop{}, errors.New("Missing value")
	}

	ts.Commit()
	return Unop{
		Base:  Base{AST_UNOP_PREFIX},
		Op:    op,
		Value: val,
	}, nil
}

func (ts *Scanner) ParseUnopPostfix(val Node) (Unop, error) {
	op := ts.ParseAnyType(0,
		p.T_MINUS,
		p.T_EXCLAM,
		p.T_BITNOT,
		p.T_DECREMENT,
		p.T_INCREMENT,
	)
	if op == nil {
		return Unop{}, errors.New("Missing operator")
	}

	ts.Move(1)
	return Unop{
		Base:  Base{AST_UNOP_POSTFIX},
		Op:    op,
		Value: val,
	}, nil
}

type Binop struct {
	Base
	Op          *p.Token
	Left, Right Node
}

func (b *Binop) Start() p.Location { return b.Left.Start() }
func (b *Binop) End() p.Location   { return b.Right.End() }

func (b *Binop) Render(nb *NodeBuilder) {
	nb.RenderNode(
		&b.Base,
		NodeField{"Operator", b.Op.Type.String()},
		NodeField{"Left", b.Left},
		NodeField{"Right", b.Right},
	)
}

func (ts *Scanner) ParseBinop(left Node) (Binop, error) {
	op := ts.ParseAnyType(0,
		p.T_PLUS,
		p.T_MINUS,
		p.T_DIV,
		p.T_MUL,
		p.T_MOD,
		p.T_AND,
		p.T_OR,
		p.T_BITAND,
		p.T_BITOR,
		p.T_BITXOR,
		p.T_LEQ,
		p.T_GEQ,
		p.T_EQ,
		p.T_LESS,
		p.T_MORE,
		p.T_LSHIFT,
		p.T_RSHIFT,
	)
	if op == nil {
		return Binop{}, errors.New("Missing operator")
	}

	ts.Save()
	ts.Move(1)

	right, err := ts.ParseExpr(nil)
	if err != nil {
		ts.Restore()
		return Binop{}, errors.New("Missing value")
	}

	ts.Commit()
	return Binop{
		Base:  Base{AST_BINOP},
		Op:    op,
		Left:  left,
		Right: right,
	}, nil
}

type KwdStmt struct {
	Base
	Kwd   *p.Token
	Value Node
}

func (k *KwdStmt) Start() p.Location { return k.Kwd.Loc }
func (k *KwdStmt) End() p.Location {
	if k.Value != nil {
		return k.Value.End()
	}
	return k.Kwd.Loc
}

func (k *KwdStmt) Render(nb *NodeBuilder) {
	nb.RenderNode(
		&k.Base,
		NodeField{"Value", k.Value},
	)
}

type VALUE_TYPE int

const (
	VT_NONE VALUE_TYPE = iota
	VT_OPTIONAL
	VT_REQUIRED
)

func (ts *Scanner) ParseKwdStmt(
	kwd_str string,
	t AST_TYPE,
	value_type VALUE_TYPE,
) (KwdStmt, error) {
	g := ts.GuardStart()
	defer ts.GuardEnd(g)

	kwd := ts.ParseExact(0, kwd_str)
	if kwd == nil {
		return KwdStmt{}, errors.New("Missing " + kwd_str + " keyword")
	}

	ts.Save()
	ts.Move(1)

	var value Node = nil
	if value_type != VT_NONE {
		var err error = nil
		value, err = ts.ParseExpr(nil)
		if err != nil {
			if value_type != VT_OPTIONAL {
				return KwdStmt{}, err
			}
			value = nil
		}
	}

	ts.Commit()
	return KwdStmt{
		Base:  Base{t},
		Kwd:   kwd,
		Value: value,
	}, nil
}

// BlockStmt represents a statement of the general form
//
//	kwd (Condition) {Body}
//
// However, the parentheses around "Condition" and braces
// around "Body" are fully optional, so more general form
// would be
//
//	kwd Condition Body
type BlockStmt struct {
	Base
	kwd       *p.Token
	Condition Node
	Body      Node
}

func (b *BlockStmt) Start() p.Location { return b.kwd.Loc }
func (b *BlockStmt) End() p.Location   { return b.Body.End() }

func (b *BlockStmt) Render(nb *NodeBuilder) {
	nb.RenderNode(
		&b.Base,
		NodeField{"Condition", b.Condition},
		NodeField{"Body", b.Body},
	)
}

func (ts *Scanner) ParseBlockStmt(kwd_str string, t AST_TYPE) (BlockStmt, error) {
	g := ts.GuardStart()
	defer ts.GuardEnd(g)

	kwd := ts.ParseExact(0, kwd_str)
	if kwd == nil {
		return BlockStmt{}, errors.New("Failed to parse block")
	}

	ts.Save()
	ts.Move(1)

	cond, err := ts.ParseExpr(nil)
	if err != nil {
		ts.Restore()
		return BlockStmt{}, err
	}

	var body Node
	b, err := ts.ParseBlock()
	if err != nil {
		body, err = ts.ParseExpr(nil)
		if err != nil {
			ts.Restore()
			return BlockStmt{}, err
		}
	} else {
		body = &b
	}

	ts.Commit()
	return BlockStmt{
		Base:      Base{t},
		kwd:       kwd,
		Condition: cond,
		Body:      body,
	}, nil
}

type IfStmt struct {
	BlockStmt
	Elseifs []BlockStmt
	Else    *BlockStmt
}

func (i *IfStmt) End() p.Location {
	if i.Else != nil {
		return i.Else.End()
	}
	if i.Elseifs != nil && len(i.Elseifs) > 0 {
		return i.Elseifs[len(i.Elseifs)-1].End()
	}
	return i.Body.End()
}

func (i *IfStmt) Render(nb *NodeBuilder) {
	nb.RenderNode(
		&i.Base,
		NodeField{"Condition", i.Condition},
		NodeField{"Body", i.Body},
		NodeField{"Elseifs", i.Elseifs},
		NodeField{"Else", i.Else},
	)
}

func (ts *Scanner) ParseIfStmt() (IfStmt, error) {
	g := ts.GuardStart()
	defer ts.GuardEnd(g)

	ts.Save()

	ifstmt, err := ts.ParseBlockStmt("if", AST_IF)
	if err != nil {
		ts.Restore()
		return IfStmt{}, errors.New("Not an if statement")
	}

	elifs := make([]BlockStmt, 0)
	for {
		kwd1 := ts.ParseExact(0, "else")
		kwd2 := ts.ParseExact(1, "if")
		if kwd1 == nil || kwd2 == nil {
			break
		}
		ts.Move(1)

		elif, err := ts.ParseBlockStmt("if", AST_ELIF)
		if err != nil {
			ts.Restore()
			return IfStmt{}, err
		}

		elifs = append(elifs, elif)
	}

	var else_block *BlockStmt
	if ts.ParseExact(0, "else") != nil {
		eb, err := ts.ParseBlockStmt("else", AST_ELSE)
		if err != nil {
			ts.Restore()
			return IfStmt{}, err
		}
		else_block = &eb
	}

	ts.Commit()
	return IfStmt{
		BlockStmt: ifstmt,
		Elseifs:   elifs,
		Else:      else_block,
	}, nil
}

type ForLoop struct {
	Base
	kwd    *p.Token
	Assign VarDecl
	Cond   Node
	Oper   Node
	Body   Node
}

func (f *ForLoop) Start() p.Location { return f.kwd.Loc }
func (f *ForLoop) End() p.Location   { return f.Body.End() }

func (f *ForLoop) Render(nb *NodeBuilder) {
	nb.RenderNode(
		&f.Base,
		NodeField{"Assign", f.Assign},
		NodeField{"Cond", f.Cond},
		NodeField{"Oper", f.Oper},
		NodeField{"Body", f.Body},
	)
}

func (ts *Scanner) ParseForLoop() (ForLoop, error) {
	g := ts.GuardStart()
	defer ts.GuardEnd(g)

	ts.Save()

	kwd := ts.ParseExact(0, "for")
	op := ts.ParseType(1, p.T_LPAREN)

	if kwd == nil || op == nil {
		ts.Restore()
		return ForLoop{}, errors.New("Not a for loop")
	}

	ts.Move(2)

	decl, err := ts.ParseVarDecl()
	if err != nil {
		ts.Restore()
		return ForLoop{}, err
	}

	semi := ts.ParseType(0, p.T_SEMICOLON)
	if semi == nil {
		ts.Restore()
		return ForLoop{}, errors.New("Missing semicolon")
	}

	cond, err := ts.ParseExpr(nil)
	if err != nil {
		ts.Restore()
		return ForLoop{}, err
	}

	semi = ts.ParseType(0, p.T_SEMICOLON)
	if semi == nil {
		ts.Restore()
		return ForLoop{}, errors.New("Missing semicolon")
	}

	oper, err := ts.ParseExpr(nil)
	if err != nil {
		ts.Restore()
		return ForLoop{}, err
	}

	ts.EatSemicolon()

	var body Node
	b, err := ts.ParseBlock()
	if err != nil {
		body, err = ts.ParseExpr(nil)
		if err != nil {
			ts.Restore()
			return ForLoop{}, err
		}
	} else {
		body = &b
	}

	return ForLoop{
		Base:   Base{AST_FOR},
		kwd:    kwd,
		Assign: decl,
		Cond:   cond,
		Oper:   oper,
		Body:   body,
	}, nil
}

func (ts *Scanner) ParseWhileLoop() (BlockStmt, error) {
	return ts.ParseBlockStmt("while", AST_WHILE)
}
