# References:
# [1] Needed so the Go files of semgrep-rules do not interfere with static analysis

DOCS_DIR := ./docs
OUT_DIR := ./out
TESTS_DIR := ./tests

VERSION := $(shell git describe --tags)
ifeq ($(VERSION),)
	_ := $(error Cannot determine build version)
endif

SEMGREP_MODFILE := $(TESTS_DIR)/semgrep-rules/go.mod

.PHONY: build
build:
	mkdir -p $(OUT_DIR)
	go build -ldflags="-w -s -X main.version=$(VERSION)" -o $(OUT_DIR)/pushbits ./cmd/pushbits

.PHONY: clean
clean:
	rm -rf $(DOCS_DIR)
	rm -rf $(OUT_DIR)
	rm -rf $(SEMGREP_MODFILE)

.PHONY: test
test:
	touch $(SEMGREP_MODFILE) # See [1].
	go fmt ./...
	go vet ./...
	gocyclo -over 10 $(shell find . -type f \( -iname '*.go' ! -path "./tests/semgrep-rules/*" \))
	staticcheck ./...
	go test -v -cover ./...
	gosec -exclude-dir=tests ./...
	semgrep --lang=go --config=tests/semgrep-rules/go --metrics=off
	rm -rf $(SEMGREP_MODFILE) # See [1].
	@printf '\n%s\n' "> Test successful"

.PHONY: setup
setup:
	git submodule update --init --recursive
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install honnef.co/go/tools/cmd/staticcheck@v0.2.2
	poetry install

.PHONY: swag
swag:
	swag init --parseDependency=true --exclude $(TESTS_DIR) -g cmd/pushbits/main.go
