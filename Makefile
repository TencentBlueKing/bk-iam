.PHONY: init dep migrations mock lint lint-dupl test bench build build-linux build-aarch64 clean all serve cov

VERSION = `head -1 VERSION`

init:
	pip install pre-commit
	pre-commit install
	# go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.43.0
	# for make doc
	go install github.com/swaggo/swag/cmd/swag@v1.7.6
	# for make mock
	go install github.com/golang/mock/mockgen@v1.4.4
	# for ginkgo
	go install github.com/onsi/ginkgo/v2/ginkgo@latest

dep:
	go mod tidy
	go mod vendor

doc:
	swag init

godoc:
	godoc -http=127.0.0.1:6060 -goroot="."

# migrations:
# 	sh pkg/database/migrations/template.sh pkg/database/migrations

mock:
	go generate ./...

lint:
	export GOFLAGS=-mod=vendor
	golangci-lint run

lint-dupl:
	export GOFLAGS=-mod=vendor
	golangci-lint run --no-config --disable-all --enable=dupl

test:
	go test -mod=vendor -gcflags=all=-l $(shell go list ./... | grep -v mock | grep -v docs) -covermode=count -coverprofile .coverage.cov

cov:
	go tool cover -html=.coverage.cov

bench:
	go test -run=nonthingplease -benchmem -bench=. $(shell go list ./... | grep -v /vendor/)

build:
	# go build .
	go build -mod=vendor -tags=jsoniter -ldflags "-X iam/pkg/version.Version=${VERSION} -X iam/pkg/version.Commit=`git rev-parse HEAD` -X iam/pkg/version.BuildTime=`date +%Y-%m-%d_%I:%M:%S` -X 'iam/pkg/version.GoVersion=`go version`'" .

build-linux:
	# GOOS=linux GOARCH=amd64 go build .
	GOOS=linux GOARCH=amd64 go build -mod=vendor -tags=jsoniter -ldflags "-X iam/pkg/version.Version=${VERSION} -X iam/pkg/version.Commit=`git rev-parse HEAD` -X iam/pkg/version.BuildTime=`date +%Y-%m-%d_%I:%M:%S` -X 'iam/pkg/version.GoVersion=`go version`'" .

build-aarch64:
	GOOS=linux GOARCH=arm64 go build -mod=vendor -tags=jsoniter -ldflags "-X iam/pkg/version.Version=${VERSION} -X iam/pkg/version.Commit=`git rev-parse HEAD` -X iam/pkg/version.BuildTime=`date +%Y-%m-%d_%I:%M:%S` -X 'iam/pkg/version.GoVersion=`go version`'" .

all: lint test build

serve: build
	./iam -c config.yaml
