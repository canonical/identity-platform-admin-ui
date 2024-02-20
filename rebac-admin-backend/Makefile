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
	./script/pull-spec.sh
.PHONY: pull-spec
