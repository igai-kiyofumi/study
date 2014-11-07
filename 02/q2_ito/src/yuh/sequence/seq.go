package sequence

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

// Generics欲しい
type Anything interface{}

// シーケンスメソッドチェイン試作
type Enumerable struct {
	// 出力チャネル
	Out chan Anything
	// 終了/エラー通知
	Done chan error
	// deferすべき処理を積む
	Deferee func()
}

// Readerをruneのシーケンスとして返す: bufioのReadRuneを利用
// reader *io Reader: 入力
// deferee func()   : defer対象としてセットするfunction
// e Enumerable     : 出力
func RuneSrc(reader *io.Reader, deferee func()) (e Enumerable) {
	ret := Enumerable{
		Out:     make(chan Anything),
		Done:    make(chan error),
		Deferee: deferee,
	}
	go func() {
		rb := bufio.NewReaderSize(*reader, 4096)
		r, _, err := rb.ReadRune()
		for ; err == nil; r, _, err = rb.ReadRune() {
			ret.Out <- r
		}
		if err != io.EOF {
			ret.Done <- err
		} else {
			ret.Done <- nil
		}
		close(ret.Out)
		close(ret.Done)
	}()
	return ret
}

// Enumerableの出力を渡したfunctionでマッピングする
// f func(AnyValue)AnyValue   : 変換するfunction
func (e Enumerable) Map(f func(Anything) (Anything, error)) Enumerable {
	ret := Enumerable{
		Out:     make(chan Anything),
		Done:    make(chan error),
		Deferee: e.Deferee,
	}

	go func() {
		var err error
		for {
			select {
			case val := <-e.Out:
				if x, err_ := f(val); err_ != nil {
					if err != nil {
						err = errors.New(err_.Error() + "\n" + err.Error())
					} else {
						err = err_
					}
				} else {
					ret.Out <- x
				}
			case d := <-e.Done:
				if err != nil {
					if d != nil {
						err = errors.New(err.Error() + "\n" + d.Error())
					} else {
						d = err
					}
				}
				ret.Done <- d
				close(ret.Out)
				close(ret.Done)
				return
			}
		}
	}()

	return ret
}

// EnumerableをWriterに出力する
// writer *io.Writer: 出力先
// done chan error  : 完了/エラー通知チャネル
func (e Enumerable) Sink(writer *io.Writer) (done chan error) {
	ret := make(chan error)
	wb := bufio.NewWriterSize(*writer, 4096)

	go func() {
		defer wb.Flush()
		defer e.Deferee()
		for {
			select {
			case val := <-e.Out:
				switch t := val.(type) {
				case rune:
					wb.WriteRune(t)
				case string:
					wb.WriteString("\"" + t + "\", ")
				}
			case done := <-e.Done:
				wb.Flush()
				ret <- done
				close(ret)
				return
			}
		}
	}()
	return ret
}

// EnumerableをWriterに出力する
// writer *io.Writer: 出力先
// done chan error  : 完了/エラー通知チャネル
func (e Enumerable) PrintSink() (done chan error) {
	ret := make(chan error)

	go func() {
		defer e.Deferee()
		for {
			select {
			case val := <-e.Out:
				fmt.Print(val)
				fmt.Print(" ")
			case done := <-e.Done:
				ret <- done
				return
			}
		}
	}()
	return ret
}

func (e Enumerable) SliceSink(n int) (ret []Anything, err error) {
	ret = make([]Anything, 0, n)
	defer e.Deferee()
	for {
		select {
		case val := <-e.Out:
			ret = append(ret, val)
		case done := <-e.Done:
			err = done
			return
		}
	}
}
