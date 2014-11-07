package expression

import (
	"errors"
	"fmt"
	"strconv"
	. "yuh/sequence"
)

// 式
type Expression interface {
	Eval() (Expression, error)
}

// 二引数
type BinaryFunction interface {
	Expression
	GetLhs() (Expression, error)
	setLhs(e Expression) error
	GetRhs() (Expression, error)
	setRhs(e Expression) error
}

type UnaryFunction interface {
	Expression
	GetValue() (Expression, error)
}

//S ::= <式>
//<式>   ::= <項>(('+'|'-')<式>)?
//<項>   ::= <因子>(('*'|'/')<項>)?
//<因子> ::= '('<式>')'|<数>
//<数>   ::= [1-9][0-9]*

func Analyze(x []Expression) (ex Expression, rest []Expression, err error) {
	return expr(x)
}

func expr(arg []Expression) (ex Expression, rest []Expression, err error) {
	lhs, r, e := term(arg)
	if e != nil {
		rest = arg
		err = fmt.Errorf("expr: unmatch (<term>)\n%s", e.Error())
		return
	}
	for {
		ex = lhs
		rest = r
		if len(r) <= 0 {
			return // 終端
		}

		_, oka := r[0].(*PlusOp)
		_, oks := r[0].(*MinusOp)
		if oka || oks {
			rhs, r_, e := term(r[1:])
			if e != nil {
				return // <式>不一致
			}
			op, _ := r[0].(BinaryFunction)
			op.setLhs(lhs)
			op.setRhs(rhs)

			lhs = op
			r = r_
		} else {
			return
		}
	}
}

func term(arg []Expression) (ex Expression, rest []Expression, err error) {
	lhs, r, e := factor(arg)
	if e != nil {
		rest = arg
		err = fmt.Errorf("term: unmatch (<factor>)\n%s", e.Error())
		return
	}
	for {
		ex = lhs
		rest = r
		if len(r) <= 0 {
			return // 終端
		}

		_, oka := r[0].(*MultOp)
		_, oks := r[0].(*DivOp)
		if oka || oks {
			rhs, r_, e := factor(r[1:])
			if e != nil {
				return // <項>不一致
			}
			op, _ := r[0].(BinaryFunction)
			op.setLhs(lhs)
			op.setRhs(rhs)

			lhs = op
			r = r_
		} else {
			return
		}
	}
}

func factor(arg []Expression) (ex Expression, rest []Expression, err error) {
	rest = arg
	if len(arg) <= 0 {
		err = fmt.Errorf("factor: empty argument")
		return
	}
	if _, ok := arg[0].(*OpenParen); ok {
		x, r, e := expr(arg[1:])
		if e != nil {
			err = fmt.Errorf("factor: unmatch (<expr>)\n%s", e.Error())
			return // 内部の<式>が不整合
		}
		if len(r) > 0 {
			if _, ok := r[0].(*CloseParen); ok {
				ex = &Parenthesis{expr: x}
				rest = r[1:]
				return // '('<式>')'
			}
		}
		err = fmt.Errorf("factor: Inconsistent parentheses\n%s", e.Error())
		return // 括弧が閉じていない
	} else {
		x, r, e := number(arg)
		if e != nil {
			err = fmt.Errorf("factor: unmatch <number>\n%s", e.Error())
			return // 内部の<数>が不整合
		} else {
			return x, r, e // <数>
		}
	}
}

// <数>
func number(arg []Expression) (ex Expression, rest []Expression, err error) {
	if len(arg) <= 0 {
		rest = arg
		err = fmt.Errorf("number: empty argument")
	} else if x, ok := arg[0].(*NumberExpr); !ok {
		rest = arg
		err = fmt.Errorf("number: invalid type Expression")
	} else {
		ex = x
		rest = arg[1:]
	}
	return
}

// Map用
func scanToken(str string) (Anything, error) {
	switch str {
	case "+":
		return &PlusOp{}, nil
	case "-":
		return &MinusOp{}, nil
	case "*":
		return &MultOp{}, nil
	case "/":
		return &DivOp{}, nil
	case "(":
		return &OpenParen{}, nil
	case ")":
		return &CloseParen{}, nil
	default:
		val, err := strconv.ParseFloat(str, 64)
		return &NumberExpr{Val: val}, err
	}
	return nil, nil
}

// RuneSrc  Readerを(runeの)シーケンスとしてみなし、
// Map      各要素を変換し、
// RuneSink Writerに流し込む

// トークン分割を行うシーケンス
func TokenizeSrc(sep string, symbol []string, input string) (e Enumerable) {
	ret := Enumerable{
		Out:     make(chan Anything),
		Done:    make(chan error),
		Deferee: func() {},
	}

	go func() {
		st := 0
		cur := 0
		l := len(input)
		for cur < l {
			prev := input[st:cur]       // 直前までの文字列
			focus := input[cur : cur+1] // カーソル位置
			if focus == sep {           // 区切り文字
				if len(prev) > 0 {
					ret.Out <- prev
				}
				cur++
				st = cur
			} else { // 何らかのトークン
				if s := func() string {
					for _, s := range symbol {
						if focus == s {
							return s
						}
					}
					return ""
				}(); s != "" { // 演算子類
					if len(prev) > 0 { // 直前までの物が1トークン
						ret.Out <- prev
					}
					ret.Out <- focus
					cur++
					st = cur
				} else { // 区切られない
					cur++
				}
			}
		}
		if st != cur {
			ret.Out <- input[st:cur]
		}
		ret.Done <- nil
	}()

	return ret
}

func ParseExprToken(lhs Anything) (ret Anything, rerr error) {
	if t, ok := lhs.(string); ok {
		if x, err := scanToken(t); err != nil {
			rerr = err
			return
		} else {
			ret = x
			return
		}
	} else {
		panic(errors.New("value is NOT [string]"))
	}
}

// 数
type NumberExpr struct {
	Val float64
}

type BinaryExpression struct {
	lhs Expression
	rhs Expression
}

type PlusOp struct {
	BinaryExpression
}
type MinusOp struct {
	BinaryExpression
}
type MultOp struct {
	BinaryExpression
}
type DivOp struct {
	BinaryExpression
}

type OpenParen struct{}
type CloseParen struct{}

type Parenthesis struct {
	expr Expression
}

// implements for fmt.Stringer

func toString(arg Anything) string {
	if x, ok := arg.(fmt.Stringer); ok {
		return x.String()
	} else {
		return ""
	}
}

func (x NumberExpr) String() string {
	return strconv.FormatFloat(x.Val, 'f', -1, 64)
}

func (op PlusOp) String() string {
	if op.lhs == nil && op.rhs == nil {
		return "+"
	} else {
		return "(+ " + toString(op.lhs) + " " + toString(op.rhs) + ")"
	}
}
func (op MinusOp) String() string {
	if op.lhs == nil && op.rhs == nil {
		return "+"
	} else {
		return "(- " + toString(op.lhs) + " " + toString(op.rhs) + ")"
	}
}
func (op MultOp) String() string {
	if op.lhs == nil && op.rhs == nil {
		return "+"
	} else {
		return "(* " + toString(op.lhs) + " " + toString(op.rhs) + ")"
	}
}
func (op DivOp) String() string {
	if op.lhs == nil && op.rhs == nil {
		return "+"
	} else {
		return "(/ " + toString(op.lhs) + " " + toString(op.rhs) + ")"
	}
}
func (p OpenParen) String() string {
	return "("
}
func (p CloseParen) String() string {
	return ")"
}
func (p Parenthesis) String() string {
	return "( " + toString(p.expr) + " )"
}

// implements for Expression

func (expr *NumberExpr) Eval() (Expression, error) {
	return expr, nil
}

func evaluate(lhs Expression, rhs Expression) (retLhs *NumberExpr, retRhs *NumberExpr, err error) {
	if le, e := lhs.Eval(); e != nil {
		err = e
		return
	} else {
		r, ok := le.(*NumberExpr)
		if !ok {
			err = fmt.Errorf("can't evaluate to Number")
			return
		}
		retLhs = r
	}

	if re, e := rhs.Eval(); e != nil {
		err = e
		return
	} else {
		r, ok := re.(*NumberExpr)
		if !ok {
			err = fmt.Errorf("can't evaluate to Number")
			return
		}
		retRhs = r
	}

	return
}

func (op *PlusOp) Eval() (Expression, error) {
	lhs, rhs, err := evaluate(op.lhs, op.rhs)
	if err != nil {
		return nil, err
	}
	return &NumberExpr{
		Val: lhs.Val + rhs.Val,
	}, nil
}
func (op *MinusOp) Eval() (Expression, error) {
	lhs, rhs, err := evaluate(op.lhs, op.rhs)
	if err != nil {
		return nil, err
	}
	return &NumberExpr{
		Val: lhs.Val - rhs.Val,
	}, nil
}
func (op *MultOp) Eval() (Expression, error) {
	lhs, rhs, err := evaluate(op.lhs, op.rhs)
	if err != nil {
		return nil, err
	}
	return &NumberExpr{
		Val: lhs.Val * rhs.Val,
	}, nil
}
func (op *DivOp) Eval() (Expression, error) {
	lhs, rhs, err := evaluate(op.lhs, op.rhs)
	if err != nil {
		return nil, err
	}
	return &NumberExpr{
		Val: lhs.Val / rhs.Val,
	}, nil
}

// Implements for BinaryOperator

func (f *BinaryExpression) GetLhs() (Expression, error) {
	return f.lhs, nil
}
func (f *BinaryExpression) GetRhs() (Expression, error) {
	return f.rhs, nil
}
func (f *BinaryExpression) setLhs(e Expression) error {
	f.lhs = e
	return nil
}
func (f *BinaryExpression) setRhs(e Expression) error {
	f.rhs = e
	return nil
}
func (f *BinaryExpression) Eval() (Expression, error) {
	panic("not implemented")
}

// Parentheses

func (p *Parenthesis) Eval() (Expression, error) {
	return p.expr.Eval()
}

func (*OpenParen) Eval() (e Expression, err error) {
	err = errors.New("'(' is not expression")
	return
}
func (*CloseParen) Eval() (e Expression, err error) {
	err = errors.New("')' is not expression")
	return
}
