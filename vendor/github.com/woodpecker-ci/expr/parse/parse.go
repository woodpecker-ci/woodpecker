package parse

import (
	"bytes"
	"fmt"
)

// Tree is the representation of a single parsed SQL statement.
type Tree struct {
	Root BoolExpr

	lex   *lexer
	depth int
}

// Parse parses the SQL statement and returns a Tree.
func Parse(buf []byte) (*Tree, error) {
	t := new(Tree)
	t.lex = new(lexer)
	return t.Parse(buf)
}

// Parse parses the SQL statement buffer to construct an ast
// representation for execution.
func (t *Tree) Parse(buf []byte) (tree *Tree, err error) {
	defer t.recover(&err)
	t.lex.init(buf)
	t.Root = t.parseExpr()
	return t, nil
}

// recover is the handler that turns panics into returns.
func (t *Tree) recover(err *error) {
	if e := recover(); e != nil {
		*err = e.(error)
	}
}

// errorf formats the error and terminates processing.
func (t *Tree) errorf(format string, args ...interface{}) {
	t.Root = nil
	format = fmt.Sprintf("selector: parse error:%d: %s", t.lex.start, format)
	panic(fmt.Errorf(format, args...))
}

func (t *Tree) parseExpr() BoolExpr {
	switch t.lex.peek() {
	case tokenLparen:
		t.lex.scan()
		return t.parseGroup()
	case tokenNot:
		t.lex.scan()
		return t.parseNot()
	}

	left := t.parseVal()
	node := t.parseComparison(left)

	switch t.lex.scan() {
	case tokenOr:
		return t.parseOr(node)
	case tokenAnd:
		return t.parseAnd(node)
	case tokenRparen:
		if t.depth == 0 {
			t.errorf("unexpected token")
			return nil
		}
		return node
	default:
		return node
	}
}

func (t *Tree) parseGroup() BoolExpr {
	t.depth++
	node := t.parseExpr()
	t.depth--

	switch t.lex.scan() {
	case tokenOr:
		return t.parseOr(node)
	case tokenAnd:
		return t.parseAnd(node)
	case tokenEOF:
		return node
	default:
		t.errorf("unexpected token")
		return nil
	}
}

func (t *Tree) parseAnd(left BoolExpr) BoolExpr {
	node := new(AndExpr)
	node.Left = left
	node.Right = t.parseExpr()
	return node
}

func (t *Tree) parseOr(left BoolExpr) BoolExpr {
	node := new(OrExpr)
	node.Left = left
	node.Right = t.parseExpr()
	return node
}

func (t *Tree) parseNot() BoolExpr {
	node := new(NotExpr)
	node.Expr = t.parseExpr()
	return node
}

func (t *Tree) parseComparison(left ValExpr) BoolExpr {
	var negate bool
	if t.lex.peek() == tokenNot {
		t.lex.scan()
		negate = true
	}

	op := t.parseOperator()

	if negate {
		switch op {
		case OperatorIn:
			op = OperatorNotIn
		case OperatorGlob:
			op = OperatorNotGlob
		case OperatorRe:
			op = OperatorNotRe
		case OperatorBetween:
			op = OperatorNotBetween
		}
	}

	switch op {
	case OperatorBetween:
		return t.parseBetween(left)
	case OperatorNotBetween:
		return t.parseNotBetween(left)
	}

	node := new(ComparisonExpr)
	node.Left = left
	node.Operator = op

	switch node.Operator {
	case OperatorIn, OperatorNotIn:
		node.Right = t.parseList()
	case OperatorRe, OperatorNotRe:
		// TODO we should use a custom regexp node here that parses and
		// compiles the regexp, insteam of recompiling on every evaluation.
		node.Right = t.parseVal()
	default:
		node.Right = t.parseVal()
	}
	return node
}

func (t *Tree) parseNotBetween(value ValExpr) BoolExpr {
	node := new(NotExpr)
	node.Expr = t.parseBetween(value)
	return node
}

func (t *Tree) parseBetween(value ValExpr) BoolExpr {
	left := new(ComparisonExpr)
	left.Left = value
	left.Operator = OperatorGte
	left.Right = t.parseVal()

	if t.lex.scan() != tokenAnd {
		t.errorf("unexpected token, expecting AND")
		return nil
	}

	right := new(ComparisonExpr)
	right.Left = value
	right.Operator = OperatorLte
	right.Right = t.parseVal()

	node := new(AndExpr)
	node.Left = left
	node.Right = right
	return node
}

func (t *Tree) parseOperator() (op Operator) {
	switch t.lex.scan() {
	case tokenEq:
		return OperatorEq
	case tokenGt:
		return OperatorGt
	case tokenGte:
		return OperatorGte
	case tokenLt:
		return OperatorLt
	case tokenLte:
		return OperatorLte
	case tokenNeq:
		return OperatorNeq
	case tokenIn:
		return OperatorIn
	case tokenRegexp:
		return OperatorRe
	case tokenGlob:
		return OperatorGlob
	case tokenBetween:
		return OperatorBetween
	default:
		t.errorf("illegal operator")
		return
	}
}

func (t *Tree) parseVal() ValExpr {
	switch t.lex.scan() {
	case tokenIdent:
		node := new(Field)
		node.Name = t.lex.bytes()
		return node
	case tokenText:
		return t.parseText()
	case tokenReal, tokenInteger, tokenTrue, tokenFalse:
		node := new(BasicLit)
		node.Value = t.lex.bytes()
		return node
	default:
		t.errorf("illegal value expression")
		return nil
	}
}

func (t *Tree) parseList() ValExpr {
	if t.lex.scan() != tokenLparen {
		t.errorf("unexpected token, expecting (")
		return nil
	}
	node := new(ArrayLit)
	for {
		next := t.lex.peek()
		switch next {
		case tokenEOF:
			t.errorf("unexpected eof, expecting )")
		case tokenComma:
			t.lex.scan()
		case tokenRparen:
			t.lex.scan()
			return node
		default:
			child := t.parseVal()
			node.Values = append(node.Values, child)
		}
	}
}

func (t *Tree) parseText() ValExpr {
	node := new(BasicLit)
	node.Value = t.lex.bytes()

	// this is where we strip the starting and ending quote
	// and unescape the string. On the surface this might look
	// like it is subject to index out of bounds errors but
	// it is safe because it is already verified by the lexer.
	node.Value = node.Value[1 : len(node.Value)-1]
	node.Value = bytes.Replace(node.Value, quoteEscaped, quoteUnescaped, -1)
	return node
}

var (
	quoteEscaped   = []byte("\\'")
	quoteUnescaped = []byte("'")
)
