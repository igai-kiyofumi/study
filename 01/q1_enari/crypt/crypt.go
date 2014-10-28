package crypt

import "math"

func EncryptBinary(data []byte, key int) []byte {
  sz := len(data);     // サイズの設定
  mkey := byte(math.Mod(float64(key), 256)) // keyの範囲を0～255に修正

  for i := 0; i < sz; i++ {
    data[i] = data[i] + mkey // 暗号化
  }
  return data
}

func DecryptBinary(data []byte, key int) []byte {
  sz := len(data);     // サイズの設定
  mkey := byte(math.Mod(float64(key), 256)) // keyの範囲を0～255に修正

  for i := 0; i < sz; i++ {
    data[i] = data[i] - mkey // 復号化
  }
  return data
}

func EncryptString(str string, key int) string {
  data := []byte(str)  // バイナリデータに変換
  sz := len(data);     // サイズの設定
  mkey := byte(math.Mod(float64(key), 256)) // keyの範囲を0～255に修正

  for i := 0; i < sz; i++ {
    data[i] = data[i] + mkey // 暗号化
  }
  return string(data[:sz])
}

func DecryptString(str string, key int) string {
  data := []byte(str)  // バイナリデータに変換
  sz := len(data);     // サイズの設定
  mkey := byte(math.Mod(float64(key), 256)) // keyの範囲を0～255に修正

  for i := 0; i < sz; i++ {
    data[i] = data[i] - mkey // 復号化
  }
  return string(data[:sz])
}
