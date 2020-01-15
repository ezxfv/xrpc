protoc-gen-go:
		@echo build protoc-gen-go
		cd cmd/protoc-gen-go
		/opt/soft/go/bin/go build -o protoc-gen-go
		mv protoc-gen-go $GOPATH/bin

greeter:
		@echo generate greeter
		cd rpc/greeter
		protoc --go_out=plugins=xrpc:. greeter.proto