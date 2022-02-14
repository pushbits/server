OUTDIR := ./out

SEMGREP_MODFILE := ./tests/semgrep-rules/go.mod

.PHONY: build
build:
	mkdir -p $(OUTDIR)
	go build -ldflags="-w -s" -o $(OUTDIR)/pushbits ./cmd/pushbits

.PHONY: clean
clean:
	rm -rf $(OUTDIR)
	rm -rf $(SEMGREP_MODFILE)

.PHONY: test
test:
	touch $(SEMGREP_MODFILE) # Needed so the Go files of semgrep-rules do not interfere with static analysis
	go fmt ./...
	go vet ./...
	gocyclo -over 10 $(shell find . -type f \( -iname '*.go' ! -path "./tests/semgrep-rules/*" \))
	staticcheck ./...
	go test -v -cover ./...
	gosec -exclude-dir=tests ./...
	semgrep --lang=go --config=tests/semgrep-rules/go --metrics=off
	rm -rf $(SEMGREP_MODFILE)
	@printf '\n%s\n' "> Test successful"

.PHONY: setup
setup:
	git submodule update --init --recursive
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	poetry install

.PHONY: swag
swag:
	swag init --parseDependency=true -d . -g cmd/pushbits/main.go
