# CONFIG 
APP_NAME=gustorrent
MAIN=./cmd/main.go

# BUILD 
build:
	go build -o bin/$(APP_NAME) $(MAIN)

run:
	go run $(MAIN)

# TEST 
.PHONY: test
test:
	go test ./...

# COVERAGE 
cover:
	go test -cover ./...

# the explorer.exe is because of the wsl
# NOTE: remember to change this if change dev env
cover-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	explorer.exe coverage.html 

# BENCH
bench:
	go test -bench=. -benchmem ./...

# FUZZ 
fuzz-bencode:
	go test -fuzz=Fuzz -fuzztime=20s ./internal/bencode

fuzz-metadata:
	go test -fuzz=Fuzz -fuzztime=20s ./internal/metadata

# CLEAN 
clean:
	rm -rf bin coverage.html

# FORMAT & LINT 
fmt:
	go fmt ./...

vet:
	go vet ./...

# ALL 
all: fmt vet test build