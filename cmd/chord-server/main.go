package main

import "x.io/xrpc/app/chord"

func main() {
	chord.Server(chord.DefaultAddr)
}
