package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

//コントローラ
func main() {
	//引数定義
	var mode   *string = flag.String("mode", "e", "Mode, e:Encode mode, d:Decode mode")
	var input  *string = flag.String("in", "input.txt", "Input filename")
	var output *string = flag.String("out", "output.txt", "Output filename")
	var key    *string = flag.String("key", "c", "Encode/Decode key")
	flag.Parse()
	if *mode != "e" && *mode != "d" {
		println("Invalid mode:", *mode)
		os.Exit(1)
	}

	//入力バッファ
	var input_buf []byte
	//変換値テキスト
	var converted_str []byte

	input_buf, err := ReadBinaryFile(*input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	os.Stdout.Write(input_buf)
	if *mode == "e" {
		converted_str = encode(input_buf, *key)
	} else if *mode == "d" {
		converted_str = decode(input_buf, *key)
	}
	os.Stdout.Write(converted_str)

	err = WriteBinaryFile(*output, converted_str)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

//ファイルをバイナリとして読み込む
func ReadBinaryFile(filename string) ([]byte, error) {
	buff, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return buff, nil
}

//バイナリファイルとして書き出し
func WriteBinaryFile(filename string, lines []byte) error {
	err := ioutil.WriteFile(filename, lines, 0644)
	if err != nil {
		return err
	}
	return nil
}

//暗号化
func encode(lines []byte, key string) []byte {
	var enc_str []byte
	enc_str = make([]byte, len(lines))

	var offset byte
	offset = calcOffset(key)

	for i := 0; i < len(lines); i++ {
		if lines[i] + offset > ^byte(0) {
			enc_str[i] = lines[i] + offset - ^byte(0)
		} else {
			enc_str[i] = lines[i] + offset
		}
	}
	return enc_str
}

//復号化
func decode(lines []byte, key string) []byte {
	var dec_str []byte
	dec_str = make([]byte, len(lines))

	var offset byte
	offset = calcOffset(key)

	for i := 0; i < len(lines); i++ {
		if lines[i] - offset < byte(0) {
			dec_str[i] = lines[i] - offset + ^byte(0)
		} else {
			dec_str[i] = lines[i] - offset
		}
	}
	return dec_str
}

//鍵をバイト値に変換
func calcOffset(key string) byte {
	byte_key := []byte(key)
	return byte_key[0]
}
