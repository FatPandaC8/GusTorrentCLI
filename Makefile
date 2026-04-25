# CONFIG 
APP_NAME=gustorrent
MAIN=./cmd/main.go

# BUILD 
build:
	go build -o bin/$(APP_NAME) $(MAIN)

run:
	go run $(MAIN)

# TEST 
test:
	go test ./...

test-parser:
	go test ./internal/parser

# COVERAGE 
cover:
	go test -cover ./...

cover-parser:
	go test -cover ./internal/parser

cover-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# FUZZ 
fuzz:
	go test -fuzz=Fuzz -fuzztime=20s ./internal/parser

# CLEAN 
clean:
	rm -rf bin coverage.out

# FORMAT & LINT 
fmt:
	go fmt ./...

vet:
	go vet ./...

# ALL 
all: fmt vet test build