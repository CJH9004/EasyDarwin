package main

import (
	"fmt"
	// "net/http"
	// _ "net/http/pprof"

	"github.com/EasyDarwin/EasyDarwin/rtsp"
)

func main() {
	// go func() {
	// 	fmt.Println(http.ListenAndServe(":6060", nil))
	// }()
	s := rtsp.GetServer()
	fmt.Println(s.Start())
}
