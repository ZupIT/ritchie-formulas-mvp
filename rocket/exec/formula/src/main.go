package main

import (
	"net"
	"os"
	"rocket/formula/pkg/formula"
	"strings"
)

func main() {
	formula.Inputs{
		Username: os.Getenv("USERNAME"),
		Password: os.Getenv("PASSWORD"),
		IPAddr:   localAddr(),
	}.Run()
}

func localAddr() string {
	conn, _ := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return strings.Split(localAddr.String(), ":")[0]
}
