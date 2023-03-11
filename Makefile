build:
	go build -o bin/main main.go

run: build
	./bin/main

test:
	go test -v ./...

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o bin/main-arm main.go

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o bin/main-amd main.go

scp: build-linux-arm64
	scp ./bin/main-arm mos@192.168.64.8:/home/data/
