# CONFIG 
APP_NAME=gustorrent
MAIN=./cmd/main.go

# BUILD 
build:
	go build -o bin/$(APP_NAME) $(MAIN)

run:
	go run $(MAIN)

docker-build:
	sudo docker build -t gustorrent-client .

docker-run:
	sudo docker run gustorrent-client

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
bench-metadata:
	go test -bench=. -benchmem -cpuprofile=metadata_cpu.out ./internal/metadata
	go tool pprof -http=:8080 metadata_cpu.out

bench-bencode:
	go test -bench=. -benchmem -cpuprofile=bencode_cpu.out ./internal/bencode
	go tool pprof -http=:8080 bencode_cpu.out

bench-tracker:
	go test -bench=. -benchmem -cpuprofile=tracker_cpu.out ./internal/tracker
	go tool pprof -http=:8080 tracker_cpu.out

bench-utils:
	go test -bench=. -benchmem -cpuprofile=utils_cpu.out ./internal/utils
	go tool pprof -http=:8080 utils_cpu.out

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