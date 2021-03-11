package main

import (
	"fmt"

	"github.com/EasyDarwin/EasyDarwin/rtsp"
)

func main() {
	s := rtsp.GetServer()
	fmt.Println(s.Start())
}
