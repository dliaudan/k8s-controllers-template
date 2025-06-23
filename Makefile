APP = k8s-controller-tutorial
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_FLAGS = -v -o $(APP) -ldflags "-X=main.version=$(VERSION)"

.PHONY: all build test run docker-build clean deps

all: build

deps:
	go mod download
	go mod tidy

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) main.go

test:
	go test ./...

run:
	go run main.go

docker-build:
	docker build --build-arg VERSION=$(VERSION) -t $(APP):latest .

clean:
	rm -f $(APP)
	go clean 