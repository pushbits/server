.PHONY: build
build:
	mkdir -p ./out
	go build -ldflags="-w -s" -o ./out/pushbits ./cmd/pushbits

.PHONY: test
test:
	stdout=$$(gofmt -l . 2>&1); if [ "$$stdout" ]; then exit 1; fi
	go vet ./...
	gocyclo -over 10 $(shell find . -iname '*.go' -type f)
	staticcheck ./...
	go test -v -cover ./...
	@printf '\n%s\n' "> Test successful"

.PHONY: setup
setup:
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest

.PHONY: swag
swag:
	swag init --parseDependency=true -d . -g cmd/pushbits/main.go
