package main

import(
 "fmt"
 "math"
)

func main(){

  // 変数初期化
  value := 10001
  result := -1

  // 背理的な考え方で素数を求める
  for result == -1 {
    flg := true  // valueを素数として扱う(仮定)
    for i := 2; i < (value / 2); i++ {
      if math.Mod(float64(value), float64(i)) == 0 {
        flg = false  // valueが素数でない(矛盾)⇒false
        break;
      }
    }
    if flg == true {
      result = value // valueが素数(結論)
    }
    value = value + 1
  }
  fmt.Printf("10000より大きい素数 : %d\n", result)
}
