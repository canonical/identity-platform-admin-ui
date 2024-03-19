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

# Run tests with coverage
test-coverage: mocks
	go test ./... -cover -coverprofile coverage_source.out $(ARGS)
	# this will be cached, just needed to get the test.json
	go test ./... -cover -coverprofile coverage_source.out  $(ARGS) -json > test_source.json
	cat coverage_source.out | grep -v "mock_*" | tee coverage.out
	cat test_source.json | grep -v "mock_*" | tee test.json
.PHONY: test-coverage

# Run tests
test: mocks
	go test ./... $(ARGS)
.PHONY: test
