package main

import (
	"time"

	reuseport "github.com/libp2p/go-reuseport"
)

func main() {
	println("xrpc cli")
	listener, _ := reuseport.Listen("tcp", ":8080")
	go func() {
		for i := 0; i < 50; i++ {
			conn, err := reuseport.Dial("tcp", ":8080", "120.27.242.169:8080")
			if err != nil {
				println(err.Error())
				time.Sleep(time.Millisecond)
				continue
			}
			n, err := conn.Write([]byte("hello"))
			if err != nil {
				time.Sleep(time.Millisecond)
				continue
			}
			if n != len("hello") {
				time.Sleep(time.Millisecond)
				continue
			}
			conn.Close()
			time.Sleep(time.Second)
		}
	}()
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		data := make([]byte, 5)
		n, err := conn.Read(data)
		if err == nil && n == 5 {
			println("get:", string(data))
			break
		}
	}
}
