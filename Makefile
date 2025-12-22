.PHONY: build run clean deps

deps:
	go mod tidy
	go get github.com/charmbracelet/huh

build: deps
	go build -o blueprint_gen ./cmd/blueprint

run: deps
	go run ./cmd/blueprint

clean:
	rm -f blueprint_gen
