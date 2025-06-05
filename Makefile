CGO_ENABLED?=0
GOOS?=linux
GO_BIN?=app
GO?=go
GOFLAGS?=-ldflags=-w -ldflags=-s -a -buildvcs
UI_FOLDER?=
MICROK8S_REGISTRY_FLAG?=SKAFFOLD_DEFAULT_REPO=localhost:32000
SKAFFOLD?=skaffold
CONFIGMAP?=deployments/kubectl/configMap.yaml


.EXPORT_ALL_VARIABLES:

mocks: vendor
	$(GO) install go.uber.org/mock/mockgen@v0.3.0
	# generate gomocks
	$(GO) generate ./...
.PHONY: mocks

test: mocks vet
	$(GO) test ./... -cover -coverprofile coverage_source.out
	# this will be cached, just needed to the test.json
	$(GO) test ./... -cover -coverprofile coverage_source.out -json > test_source.json
	cat coverage_source.out | grep -v "mock_*" | tee coverage.out
	cat test_source.json | grep -v "mock_*" | tee test.json
.PHONY: test

vet: cmd/ui/dist
	$(GO) vet ./...
.PHONY: vet

vendor:
	$(GO) mod vendor
.PHONY: vendor

build: cmd/ui/dist
	$(GO) build -o $(GO_BIN) ./
.PHONY: build

# plan is to use this as a probe, if folder is there target wont run and npm-build will skip
# but not working atm
cmd/ui/dist:
	@echo "copy dist npm files into cmd/ui folder"
	mkdir -p cmd/ui/dist
	cp -r $(UI_FOLDER)ui/dist cmd/ui/
.PHONY: cmd/ui/dist

npm-build:
	$(MAKE) -C ui/ build
.PHONY: npm-build

dev:
	@echo "after job admin-ui-openfga-setup has run, restart the identity-platform-admin-ui deployment"
	@echo "to make sure the changes in the configmap are picked up"
	@$(MICROK8S_REGISTRY_FLAG) $(SKAFFOLD) run \
		--port-forward \
		--no-prune=false \
		--tail \
		--cache-artifacts=false
.PHONY: dev


GOOSE=goose
GOOSE_DRIVER=postgres
GOOSE_DBSTRING?=postgresql://user:user@localhost:5432/admin-service?sslmode=disable
GOOSE_MIGRATION_DIR?=./migrations

db-status:
	@GOOSE_DRIVER=$(GOOSE_DRIVER) \
	GOOSE_DBSTRING=$(GOOSE_DBSTRING) \
	$(GOOSE) -dir $(GOOSE_MIGRATION_DIR) status
.PHONY: db-status

db:
	@GOOSE_DRIVER=$(GOOSE_DRIVER) \
	GOOSE_DBSTRING=$(GOOSE_DBSTRING) \
	$(GOOSE) -dir $(GOOSE_MIGRATION_DIR) up
.PHONY: migrate

db-down:
	@GOOSE_DRIVER=$(GOOSE_DRIVER) \
	GOOSE_DBSTRING=$(GOOSE_DBSTRING) \
	$(GOOSE) -dir $(GOOSE_MIGRATION_DIR) down
.PHONY: migrate-down

install-goose:
	go install github.com/pressly/goose/v3/cmd/goose@v3.24.3
.PHONY: install-goose
