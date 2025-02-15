DOCS_DIR := ./docs
OUT_DIR := ./out
TESTS_DIR := ./tests

GO_FILES := $(shell find . -type f \( -iname '*.go' \))

PB_BUILD_VERSION ?= $(shell git describe --tags)
ifeq ($(PB_BUILD_VERSION),)
	_ := $(error Cannot determine build version)
endif

.PHONY: build
build:
	mkdir -p $(OUT_DIR)
	go build -ldflags="-w -s -X main.version=$(PB_BUILD_VERSION)" -o $(OUT_DIR)/pushbits ./cmd/pushbits

.PHONY: clean
clean:
	rm -rf $(DOCS_DIR)
	rm -rf $(OUT_DIR)

.PHONY: test
test:
	stdout=$$(gofumpt -l . 2>&1); if [ "$$stdout" ]; then exit 1; fi
	go vet ./...
	misspell -error $(GO_FILES)
	gocyclo -over 10 $(GO_FILES)
	staticcheck ./...
	errcheck -exclude errcheck_excludes.txt ./...
	gocritic check -disable='#experimental,#opinionated' -@ifElseChain.minThreshold 3 ./...
	revive -set_exit_status -exclude ./docs ./...
	nilaway ./...
	go test -v -cover ./...
	gosec -exclude-dir=tests ./...
	govulncheck ./...
	@printf '\n%s\n' "> Test successful"

.PHONY: setup
setup:
	go install github.com/client9/misspell/cmd/misspell@latest
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	go install github.com/go-critic/go-critic/cmd/gocritic@latest
	go install github.com/kisielk/errcheck@latest
	go install github.com/mgechev/revive@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install go.uber.org/nilaway/cmd/nilaway@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install mvdan.cc/gofumpt@latest

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

.PHONY: run_postgres_debug
	podman run \
		--rm \
		--name=postgres \
		--network=host \
		--env-file \
		postgres-debug.env docker.io/library/postgres:15
