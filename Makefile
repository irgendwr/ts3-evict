.PHONY: all
all: test build

.PHONY: test
test:
	go vet ./...
	go test -v -vet=off ./...

.PHONY: build
build:
	GOVERSION=$(go version | awk '{print $3 " on " $4;}') goreleaser release --rm-dist --snapshot
	cp ./dist/ts3-evict_linux_amd64/ts3-evict ./ts3-evict

.PHONY: clean
clean:
	@rm -rf dist
	@rm ts3-evict

.PHONY: update
update:
	go get -u
	go mod tidy
