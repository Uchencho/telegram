OUTPUT = main

main: 
	go build -o $(OUTPUT) main.go

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:
	rm -f $(OUTPUT)

.PHONY: local
local:
	GOOS=linux GOARCH=amd64 $(MAKE) main

.PHONY: build
build: clean local

run-local:
	go run main.go
	