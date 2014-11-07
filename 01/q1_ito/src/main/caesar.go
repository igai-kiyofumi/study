package main
 
import (
	//"fmt"
	"github.com/docopt/docopt-go"
	//"reflect"
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
)
 
// entry point
func main() {
	major_ver := "0"
	minor_ver := "1"
 
	usage := `
Encrypt/Decrypt using Caesar-Cipher.
 
Usage:
  caesar (e|encrypt) (-f | --file)<input-file> <key>
  caesar (e|encrypt) <key> <input>
  caesar (d|decrypt) (-f | --file) <input-file> <key>
  caesar (d|decrypt) <key> <input>
  caesar --version
  caesar -h|--help
 
Options:
  <input>      input text.
  <input-file> input file path.
  <key>        encryption key.
  e encrypt    encrypt input text.
  d decript    decrypt input text.
`
	// docoptでパラメータ処理
	arg, err := docopt.Parse(usage, nil, true, major_ver+"."+minor_ver, false)
	if err != nil {
		panic(err)
	}
 
	param, pdefer, perr := makeParam(arg)
	defer pdefer()
	if perr != nil {
		panic(perr)
	}
 
	// 入出力を整理
	reader := param["text"].(io.Reader)
	key := param["key"].(int)
	var stdout io.Writer = os.Stdout
 
	// 実処理部
	res := caesarEncrypt(&reader, &stdout, key)
	if res != nil {
		panic(res)
	}
}
 
// (rune, int) -> rune
// 一文字のカエサル暗号化 デコードは負数の鍵を与える
// r rune: 入力ルーン
// key int: 鍵
// out rune: 出力ルーン
func caesar(key int, r rune) (out rune) {
	// 範囲内に収まるようにスライドさせる
	f := func(st rune, ed rune) rune {
		span := (ed - st + 1)
		ret := rune(int(r) + key)
		for ret < st {
			ret = ret + span
		}
		for ed < ret {
			ret = ret - span
		}
		return ret
	}
 
	// 小文字・大文字両対応
	if unicode.IsLower(r) {
		return f('a', 'z')
	}
	if unicode.IsUpper(r) {
		return f('A', 'Z')
	}
	return r
}
 
// readerから1文字ずつ読み、caesar暗号を施してwriterに送る
// reader *io.Reader 入力
// reader *io.Writer 出力
func caesarEncrypt(reader *io.Reader, writer *io.Writer, key int) error {
	caesarAny := func(in AnyValue) AnyValue {
		return caesar(key, in.(rune))
	}
 
	return <-RuneSrc(reader, func() {}).Map(caesarAny).RuneSink(writer)
}
 
// Generics欲しい
type AnyValue interface{}
 
// シーケンスメソッドチェイン試作
type Enumerable struct {
	// 出力チャネル
	out chan AnyValue
	// 終了/エラー通知
	done chan error
	// deferすべき処理を積む
	deferee func()
}
 
// RuneSrc  Readerを(runeの)シーケンスとしてみなし、
// Map      各要素を変換し、
// RuneSink Writerに流し込む
 
// Readerをruneのシーケンスとして返す: bufioのReadRuneを利用
// reader *io Reader: 入力
// deferee func()   : defer対象としてセットするfunction
// e Enumerable     : 出力
func RuneSrc(reader *io.Reader, deferee func()) (e Enumerable) {
	ret := Enumerable{
		out:     make(chan AnyValue),
		done:    make(chan error),
		deferee: deferee,
	}
	go func() {
		rb := bufio.NewReaderSize(*reader, 4096)
		r, _, err := rb.ReadRune()
		for ; err == nil; r, _, err = rb.ReadRune() {
			ret.out <- r
		}
		if err != io.EOF {
			ret.done <- err
		} else {
			ret.done <- nil
		}
	}()
	return ret
}
 
// Enumerableの出力を渡したfunctionでマッピングする
// f func(AnyValue)AnyValue   : 変換するfunction
func (e Enumerable) Map(f func(AnyValue) AnyValue) Enumerable {
	ret := Enumerable{
		out:     make(chan AnyValue),
		done:    make(chan error),
		deferee: e.deferee,
	}
 
	go func() {
		for {
			select {
			case val := <-e.out:
				ret.out <- f(val)
			case d := <-e.done:
				ret.done <- d
				return
			}
		}
	}()
 
	return ret
}
 
// EnumerableをWriterに出力する
// writer *io.Writer: 出力先
// done chan error  : 完了/エラー通知チャネル
func (e Enumerable) RuneSink(writer *io.Writer) (done chan error) {
	ret := make(chan error)
	wb := bufio.NewWriterSize(*writer, 4096)
 
	go func() {
		defer wb.Flush()
		defer e.deferee()
		for {
			select {
			case val := <-e.out:
				switch t := val.(type) {
				case rune:
					wb.WriteRune(t)
				}
			case done := <-e.done:
				wb.Flush()
				ret <- done
				return
			}
		}
	}()
	return ret
}
 
// 引数フラグを整理して詰め直す
// arg map[string]interface{}: 引数オブジェクト
// param map[string]interface{}: パラメータ
// deferee func() deferすべき関数
// err error: エラー
func makeParam(arg map[string]interface{}) (param map[string]interface{}, deferee func(), err error) {
	key, kerr := strconv.Atoi(arg["<key>"].(string))
	if kerr != nil {
		return nil, func() {}, kerr
	}
 
	text, textd, texterr := prepareTextReader(arg)
	if texterr != nil {
		return nil, func() {}, texterr
	}
 
	enc := arg["e"].(bool) || arg["encrypt"].(bool)
	dec := arg["d"].(bool) || arg["decrypt"].(bool)
 
	if dec && !enc {
		key = -key
	}
 
	return map[string]interface{}{
			"enc":  enc,
			"dec":  dec,
			"text": text,
			"key":  key,
		}, func() {
			textd()
		}, nil
}
 
// ファイル名指定ならopen,直に渡されたらstringsのReaderを準備する
// arg map[string]interface{} 引数オブジェクト
// reader io.Reader 作成したreader
// deferee func() deferすべき関数
// err error: エラー
func prepareTextReader(arg map[string]interface{}) (reader io.Reader, deferree func(), err error) {
	return func() (io.Reader, func(), error) {
		if arg["-f"].(bool) || arg["--file"].(bool) {
			fp, ferr := os.Open(arg["<input-file>"].(string))
 
			return fp, func() {
				fp.Close()
			}, ferr
		} else {
			return strings.NewReader(arg["<input>"].(string)), func() {}, nil
		}
	}()
}
 