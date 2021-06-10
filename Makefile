IMAGE := eikendev/pushbits

.PHONY: build
build:
	mkdir -p ./out
	go build -ldflags="-w -s" -o ./out/pushbits ./cmd/pushbits

.PHONY: test
test:
	stdout=$$(gofmt -l . 2>&1); \
	if [ "$$stdout" ]; then \
		exit 1; \
	fi
	go vet ./...
	gocyclo -over 10 $(shell find . -iname '*.go' -type f)
	staticcheck ./...
	go test -v -cover ./...

.PHONY: setup
setup:
	go get -u github.com/fzipp/gocyclo/cmd/gocyclo
	go get -u honnef.co/go/tools/cmd/staticcheck

.PHONY: build_image
build_image:
	docker build -t ${IMAGE}:latest .

.PHONY: push_image
push_image:
	docker push ${IMAGE}:latest
