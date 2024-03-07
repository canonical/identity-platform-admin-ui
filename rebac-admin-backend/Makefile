# Copyright 2024 Canonical Ltd.

# Show this help
help:
	@echo 'Usage:'
	@echo 'make <command>'
	@echo ''
	@echo 'Available commands:'
	@awk '/^#/{c=substr($$0,3);next}c&&/^[[:alpha:]][[:alnum:]_-]+:/{print substr($$1,1,index($$1,":")),c}1{c=0}' $(MAKEFILE_LIST) | column -s: -t
.PHONY: help

# Pull the latest stable OpenAPI spec and generate Go types
pull-spec:
	./script/generate-resources-from-spec.sh
.PHONY: pull-spec

# Generate test mocks
mocks:
	go install go.uber.org/mock/mockgen@v0.3.0
	go generate ./...
.PHONY: mocks

# Run tests
test: mocks
	go test ./...
.PHONY: test
