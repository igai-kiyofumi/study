package main

import(
 "fmt"
 "math"
)

const LIST_SIZE = 10000

func main(){
  // 変数の宣言
  var primeList [LIST_SIZE] int
  // 変数初期化
  primeList[0] = 2 // 最初の素数は2

  for cnt := 1; cnt < LIST_SIZE; cnt++ {
    primeList[cnt] = primeList[cnt - 1] + 1  // ひとつ前の要素に1加えた値を次の要素に代入
    loop := -1
    for loop == -1 {
      flg := true  // primeList[cnt]を素数として扱う(仮定)
      for idx := 0; idx < cnt; idx++ {
        if math.Mod(float64(primeList[cnt]), float64(primeList[idx])) == 0 {
          flg = false  // primeList[cnt]を素数ではない(矛盾)
          break
        }
      }
      if flg == true {
         loop = 0  // primeList[cnt]が素数であるためループ終了
      } else {
        primeList[cnt] = primeList[cnt] + 1
      }
    }
  }
  fmt.Printf("%d番目の素数 : %d\n", LIST_SIZE, primeList[LIST_SIZE - 1])
}
