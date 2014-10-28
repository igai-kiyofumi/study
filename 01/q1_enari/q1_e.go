package main

import(
 "./crypt"
 "fmt"
 "io/ioutil"
)
 
const INPUT_FILE = "input.txt"
const ENCRYPTED_FILE = "encrypted.txt"

func main(){
    // ファイルの読み込み
    contents,err := ioutil.ReadFile(INPUT_FILE) // ReadFileの戻り値は []byte
    if err != nil {
        fmt.Println(contents, err)
        return
    }

    // 暗号化
    crypt.EncryptBinary(contents, 10)
 
    // ファイルに書き込み
    ioutil.WriteFile(ENCRYPTED_FILE, contents, 0644) // 0644はpermission
 
}
