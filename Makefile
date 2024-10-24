run: fmt
	cd cmd && go run main.go

build: fmt
	go build -o jekill cmd/main.go && cp cmd/jekill .

fmt:
	gofmt -s -w **/*.go
