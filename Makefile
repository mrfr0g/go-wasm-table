.PHONY: clean prepare copy-assets build

GOENVROOT=`go env GOROOT`
MYPATH=`pwd`
MYGOPATH=$(MYPATH):$(HOME)/go/

clean:
	rm -rf ./build

prepare: clean
	mkdir ./build

copy-assets:
	cp ./client-src/index.html ./build
	cp $(GOENVROOT)/misc/wasm/wasm_exec.js ./build

build: prepare copy-assets
	GOARCH=wasm GOOS=js GOPATH=$(MYGOPATH) go build -o ./build/main.wasm main.go
