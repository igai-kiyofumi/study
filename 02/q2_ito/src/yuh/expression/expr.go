package expression

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	. "yuh/sequence"
)

// トークン定義
func ScanToken(str string) (Anything, error) {
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
	case "sin":
		return &UnaryFunc {
			name: "sin",
			f: WrapUnaryFunc(math.Sin),
		}, nil
	case "cos":
		return &UnaryFunc {
			name: "sin",
			f: WrapUnaryFunc(math.Cos),
		}, nil
	case "tan":
		return &UnaryFunc {
			name: "sin",
			f: WrapUnaryFunc(math.Tan),
		}, nil
	case "pi":
		return &NumberExpr{Val: math.Pi}, nil
	default:
		val, err := strconv.ParseFloat(str, 64)
		return &NumberExpr{Val: val}, err
	}
	return nil, nil
}

func WrapUnaryFunc(f func(float64)float64) func(Expression)(Expression, error) {
	return func (arg Expression) (ret Expression, err error) {
		ret, err = arg.Eval()
		if err != nil {
			return
		}
		if t, ok :=ret.(*NumberExpr); ok {
			return &NumberExpr{ Val: f(t.Val) }, nil
		}
		return
	}
}

// 式
type Expression interface {
	Eval() (Expression, error)
}

type Function interface {
	Expression
	ArgumentNum() int
}


type UnaryFunction interface {
	Function
	GetValue() (Expression, error)
	setValue() (Expression, error)
}

// 二引数
type BinaryFunction interface {
	Function
	GetLhs() (Expression, error)
	setLhs(e Expression) error
	GetRhs() (Expression, error)
	setRhs(e Expression) error
}

//S ::= <式>
//<式>   ::= <項>(('+'|'-')<項>)*
//<項>   ::= <因子>(('*'|'/')<因子>)*
//<因子> ::= '('<式>')'|<関数>'('<式>(,<式>)*')'|<数>
//<数>   ::= [1-9][0-9]*
//<関数> ::= [a-zA-Z][a-zA-Z0-9]*

func Parse(x []Expression) (ex Expression, rest []Expression, err error) {
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
	} else if f, ok := arg[0].(Function); ok {
		if _, ok := arg[1].(*OpenParen); !ok {
			if s,ok := f.(fmt.Stringer); ok { 
				err = fmt.Errorf("factor: missing '(' after <" + s.String() + ">")
			} else {
				err = fmt.Errorf("factor: missing '('")
			}
			return // 括弧(がない
		}
		x,r,e := expr(arg[2:]) // <式>
		if e != nil {
			err = fmt.Errorf("factor: unmatch <func>'('<expr>(,<expr>)*')'\n%s", e.Error())
			return // 内部の<式>が不整合
		}
		for i:=1; i<f.ArgumentNum(); i++ {
			err = fmt.Errorf("factor: not implemented (n-ary functions)\n%s", e.Error())
			return // 内部の<式>が不整合
		}
		if _, ok := r[0].(*CloseParen); !ok {
			if s,ok := f.(fmt.Stringer); ok { 
				err = fmt.Errorf("factor: missing ')' after <" + s.String() + ">")
			} else {
				err = fmt.Errorf("factor: missing ')'")
			}
			return // 括弧)がない
		}
		ex = f
		rest = r[1:]
		switch f.ArgumentNum() {
		case 1:
			u, _ := f.(*UnaryFunc)
			u.setValue(x)
		default:
			err = fmt.Errorf("factor: not implemented (n-ary functions)\n%s", e.Error())
		}
		return
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


// トークン分割を行うシーケンス
func TokenizeSrc(sep string, operators []string, input string) (e Enumerable) {
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
					for _, s := range operators {
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

func TokenParser(f func (str string) (Anything, error)) func (lhs Anything) (ret Anything, rerr error) {
	return func (lhs Anything) (ret Anything, rerr error) {
		if t, ok := lhs.(string); ok {
			if x, err := f(t); err != nil {
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
}


// 数
type NumberExpr struct {
	Val float64
}

type BinaryExpression struct {
	lhs Expression
	rhs Expression
}

type UnaryExpression struct {
	expr Expression
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


type UnaryFunc struct {
	UnaryExpression
	f func (Expression) (Expression, error)
	name string
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
func (f *BinaryExpression) ArgumentNum() int {
	return 2
}



func (f *UnaryExpression) GetValue() (Expression, error) {
	return f.expr, nil
}
func (f *UnaryExpression) setValue(e Expression) error {
	f.expr = e
	return nil
}
func (f *UnaryExpression) Eval() (Expression, error) {
	panic("not implemented")
}
func (f *UnaryExpression) ArgumentNum() int {
	return 1
}

func (f *UnaryFunc) String() string {
	if f.expr == nil {
		return f.name
	} else if e, ok := f.expr.(fmt.Stringer); ok {
		return "(" + f.name + " " + e.String() + ")"
	} else {
		return "(" + f.name + " *)"
	}
}

func (f *UnaryFunc) Eval() (ret Expression, err error) {
	ret, err = f.expr.Eval()
	if err != nil {
		return
	}
	
	ret, err = f.f(ret)
	return
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
