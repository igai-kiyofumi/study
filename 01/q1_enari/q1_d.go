package main

import(
 "./crypt"
 "fmt"
 "io/ioutil"
)
 
const ENCRYPTED_FILE = "encrypted.txt"
const DECRYPTED_FILE = "decrypted.txt"

func main(){
    // ファイルの読み込み
    contents,err := ioutil.ReadFile(ENCRYPTED_FILE) // ReadFileの戻り値は []byte
    if err != nil {
        fmt.Println(contents, err)
        return
    }

    // 暗号化
    crypt.DecryptBinary(contents, 10)
 
    // ファイルに書き込み
    ioutil.WriteFile(DECRYPTED_FILE, contents, 0644) // 0644はpermission
 
}
