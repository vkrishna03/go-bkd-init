.PHONY: run build clean

run:
	go run ./cmd/app

build:
	go build -o bin/app ./cmd/app

clean:
	rm -rf bin/
