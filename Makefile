
HAS_LINT := $(shell command -v golangci-lint;)
HAS_IMPORTS := $(shell command -v goimports;)

PROJECT = github.com/simonnik/GB_Best_CourseWork_GO
GO_PKG = $(shell go list $(PROJECT)/hw1/...)

bootstrap:
ifndef HAS_LINT
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.43.0
endif
ifndef HAS_IMPORTS
	go install golang.org/x/tools/cmd/goimports
endif

init: bootstrap pre-commit-install

build: test
	go build -o bin/sql ./cmd/sql/sql.go

run:
	go run ./cmd/sql/sql.go -config="config.yaml"

test:
	@echo "+ $@"
	@go list -f '"go test -race -cover -v {{.Dir}}"' $(GO_PKG) | xargs -L 1 sh -c

fmt:
	@echo "+ $@"
	@go list -f '"gofmt -w -s -l {{.Dir}}"' $(GO_PKG) | xargs -L 1 sh -c

imports:
	@echo "+ $@"
	@go list -f '"goimports -w {{.Dir}}"' ${GO_PKG} | xargs -L 1 sh -c

check: bootstrap
	@echo "+ $@"
	@golangci-lint run ./...

pre-commit:
	@echo "+ $@"
	pre-commit run --all-files

pre-commit-install:
	@echo "+ $@"
	pre-commit install

.PHONY: bootstrap \
	build \
	test \
	fmt \
	imports \
	check \
	init \
	run