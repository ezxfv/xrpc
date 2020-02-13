package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"x.io/xrpc/pkg/generator/parser"
)

var (
	idl = flag.String("idl", "", "service description file")
)

func parseIdl(file string) error {
	meta := parser.NewMetaData()
	meta.Parse(file)
	stub := strings.ReplaceAll(file, ".go", ".stub.go")
	f, err := os.Create(stub)
	if err != nil {
		return err
	}
	return meta.Print(parser.NewXrpcStubBuilder(), f)
}

func main() {
	flag.Parse()
	if *idl != "" {
		if err := parseIdl(*idl); err != nil {
			log.Fatalln(err.Error())
		}
	}
}
