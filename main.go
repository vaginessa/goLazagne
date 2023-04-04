package main

import (
	"fmt"
	"github.com/kerbyj/goLazagne/browsers"
)

func main() {
	res := browsers.ChromeExtractDataRun()
	fmt.Println(res)
}
