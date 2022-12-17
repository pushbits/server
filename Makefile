# References:
# [1] Needed so the Go files of semgrep-rules do not interfere with static analysis

DOCS_DIR := ./docs
OUT_DIR := ./out
TESTS_DIR := ./tests

GO_FILES := $(shell find . -type f \( -iname '*.go' ! -path "./tests/semgrep-rules/*" \))

PB_BUILD_VERSION ?= $(shell git describe --tags)
ifeq ($(PB_BUILD_VERSION),)
	_ := $(error Cannot determine build version)
endif

SEMGREP_MODFILE := $(TESTS_DIR)/semgrep-rules/go.mod

.PHONY: build
build:
	mkdir -p $(OUT_DIR)
	go build -ldflags="-w -s -X main.version=$(PB_BUILD_VERSION)" -o $(OUT_DIR)/pushbits ./cmd/pushbits

.PHONY: clean
clean:
	rm -rf $(DOCS_DIR)
	rm -rf $(OUT_DIR)
	rm -rf $(SEMGREP_MODFILE)

.PHONY: test
test:
	touch $(SEMGREP_MODFILE) # See [1].
	stdout=$$(gofumpt -l $(GO_FILES) 2>&1); if [ "$$stdout" ]; then exit 1; fi
	go vet ./...
	gocyclo -over 10 $(GO_FILES)
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
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install mvdan.cc/gofumpt@latest
	poetry install

.PHONY: fmt
fmt:
	gofumpt -l -w $(GO_FILES)

.PHONY: swag
swag: build
	swag init --parseDependency=true --exclude $(TESTS_DIR) -g cmd/pushbits/main.go

.PHONY: docker_build_dev
docker_build_dev:
	podman build \
		--build-arg=PB_BUILD_VERSION=dev \
		-t local/pushbits .
