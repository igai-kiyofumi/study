package main

import (
	"fmt"
	. "yuh/expression"
	. "yuh/sequence"
)

func main() {
	// 入力
	input := []struct{ title, input string }{
		{"2.1", "(1 + 3) * 10"},
		{"2.2", "1 ERROR 3"},
		{"2.3", "sin(pi/6)"}, // 度数法めんどくさいので弧度法対・分数表記は無理。
		{"ex1", "(5-3-1)"},
		{"ex2", "(5-6/2/3-1*2)"},
	}
	for _, v := range input {
		fmt.Println(v.title)
		err := do(v.input)
		if err != nil {
			fmt.Printf("ERROR cause: \n" + err.Error())
		}
	}
}

func do(input string) (err error) {
	// トークン解析
	fmt.Printf("Parse [%s]\n", input)
	res, err := TokenizeSrc(
		" ", 
		[]string{"+", "-", "*", "/", "(", ")", ","}, 
		input).
		Map(TokenParser(ScanToken)).
		SliceSink(10)

	if err != nil {
		return
	}

	// 式木
	fmt.Printf("Construct [%v]\n", res)
	ex, _, err := Parse(*convertSliceToExpression(&res))
	if err != nil {
		return
	}

	// 評価
	fmt.Printf("Evaluate [%v][]\n", ex)
	val, err := ex.Eval()
	if err != nil {
		return
	}

	fmt.Printf("%s = %v [Expression: %v]\n", input, val, ex)
	return
}

// []Anything -> []Expression
// Generics欲しい…
func convertSliceToExpression(arg *[]Anything) (ret *[]Expression) {
	r := make([]Expression, 0, len(*arg))

	for _, v := range *arg {
		switch t := v.(type) {
		case Expression:
			r = append(r, t)
		}
	}

	return &r
}
