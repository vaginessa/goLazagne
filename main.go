package main

import (
	"fmt"
	"github.com/kerbyj/goLazagne/sysadmin"
	//"github.com/kerbyj/goLazagne/browsers"
	//"github.com/kerbyj/goLazagne/sysadmin"
)

func main() {
	//res := browsers.ChromeExtractDataRun()
	//res := browsers.MozillaExtractDataRun("browser")
	//res, _ := Run()
	res, _ := sysadmin.RDPManagerRun()
	fmt.Println(res)
}
