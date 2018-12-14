# ----------------------------------------
# init
# ----------------------------------------
PROJECT_NAME=fabric8-env
PACKAGE_NAME:=github.com/fabric8-services/$(PROJECT_NAME)
CUR_DIR=$(shell pwd)
TMP_PATH=$(CUR_DIR)/tmp
INSTALL_PREFIX=$(CUR_DIR)/bin
VENDOR_DIR=vendor
SOURCE_DIR ?= .
SOURCES := $(shell find $(SOURCE_DIR) -path $(SOURCE_DIR)/vendor -prune -o -name '*.go' -print)
DESIGN_DIR=design
DESIGNS := $(shell find $(SOURCE_DIR)/$(DESIGN_DIR) -path $(SOURCE_DIR)/vendor -prune -o -name '*.go' -print)

# declares variable that are OS-sensitive
ifeq ($(OS),Windows_NT)
	include ./.make/Makefile.win
else
	include ./.make/Makefile.lnx
endif

# Find all required tools:
GIT_BIN := $(shell command -v $(GIT_BIN_NAME) 2> /dev/null)
DEP_BIN_DIR := $(TMP_PATH)/bin
DEP_BIN := $(DEP_BIN_DIR)/$(DEP_BIN_NAME)
DEP_VERSION=v0.5.0
GO_BIN := $(shell command -v $(GO_BIN_NAME) 2> /dev/null)
DOCKER_BIN := $(shell command -v $(DOCKER_BIN_NAME) 2> /dev/null)

# for test targets
include ./.make/test.mk

# for docker targets
ifneq ($(OS),Windows_NT)
	ifdef DOCKER_BIN
		include ./.make/docker.mk
	endif
endif

# This is a fix for a non-existing user in passwd file when running in a docker
# container and trying to clone repos of dependencies
GIT_COMMITTER_NAME ?= "user"
GIT_COMMITTER_EMAIL ?= "user@example.com"
export GIT_COMMITTER_NAME
export GIT_COMMITTER_EMAIL
COMMIT=$(shell git rev-parse HEAD 2>/dev/null)
GITUNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=no)
ifneq ($(GITUNTRACKEDCHANGES),)
	COMMIT := $(COMMIT)-dirty
endif
BUILD_TIME=`date -u '+%Y-%m-%dT%H:%M:%SZ'`
BUILD_DIR = bin

.DEFAULT_GOAL := help

# Call this function with $(call log-info,"Your message")
define log-info =
@echo "INFO: $(1)"
endef



# ----------------------------------------
# targets
# ----------------------------------------
.DEFAULT_GOAL := all

.PHONY: all
all: prebuild-check deps generate build

.PHONY: help
# Based on https://gist.github.com/rcmachado/af3db315e31383502660
## display this help text.
help:/
	$(info Available targets)
	$(info -----------------)
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		helpCommand = substr($$1, 0, index($$1, ":")-1); \
		if (helpMessage) { \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			gsub(/##/, "\n                                     ", helpMessage); \
			printf "%-35s -> %s\n", helpCommand, helpMessage; \
			lastLine = "" \
		} \
	} \
	{ hasComment = match(lastLine, /^## (.*)/); \
          if(hasComment) { \
            lastLine=lastLine$$0; \
	  } \
          else { \
	    lastLine = $$0 \
          } \
        }' $(MAKEFILE_LIST)

$(BUILD_DIR):
	mkdir $(BUILD_DIR)

.PHONY: build-linux
build-linux: prebuild-check deps generate ## Builds the Linux binary for the container image into bin/ folder
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -v $(LDFLAGS) -o $(BUILD_DIR)/$(PROJECT_NAME)

image: clean-artifacts build-linux
	docker build -t $(REGISTRY_URL) \
	  --build-arg BUILD_DIR=$(BUILD_DIR)\
	  --build-arg PROJECT_NAME=$(PROJECT_NAME)\
	  -f $(VENDOR_DIR)/github.com/fabric8-services/fabric8-common/makefile/$(DOCKERFILE) .


$(INSTALL_PREFIX):
	mkdir -p $(INSTALL_PREFIX)
$(TMP_PATH):
	mkdir -p $(TMP_PATH)

.PHONY: show-info
show-info:
	$(call log-info,"$(shell go version)")
	$(call log-info,"$(shell go env)")

# TODO add CHECK_GOPATH_BIN
.PHONY: prebuild-check
prebuild-check: $(TMP_PATH) $(INSTALL_PREFIX) show-info
# Check that all tools where found
ifndef GIT_BIN
	$(error The "$(GIT_BIN_NAME)" executable could not be found in your PATH)
endif
ifndef DEP_BIN
	$(error The "$(DEP_BIN_NAME)" executable could not be found in your PATH)
endif
ifndef GO_BIN
	$(error The "$(GO_BIN_NAME)" executable could not be found in your PATH)
endif


.PHONY: deps
deps: $(DEP_BIN) $(VENDOR_DIR) ## Download build dependencies.

# install dep in a the tmp/bin dir of the repo
$(DEP_BIN):
	@echo "Installing 'dep' $(DEP_VERSION) at '$(DEP_BIN_DIR)'..."
	mkdir -p $(DEP_BIN_DIR)
ifeq ($(UNAME_S),Darwin)
	@curl -L -s https://github.com/golang/dep/releases/download/$(DEP_VERSION)/dep-darwin-amd64 -o $(DEP_BIN) 
	@cd $(DEP_BIN_DIR) && \
	curl -L -s https://github.com/golang/dep/releases/download/$(DEP_VERSION)/dep-darwin-amd64.sha256 -o $(DEP_BIN_DIR)/dep-darwin-amd64.sha256 && \
	echo "1a7bdb0d6c31ecba8b3fd213a1170adf707657123e89dff234871af9e0498be2  dep" > dep-darwin-amd64.sha256 && \
	shasum -a 256 --check dep-darwin-amd64.sha256
else
	@curl -L -s https://github.com/golang/dep/releases/download/$(DEP_VERSION)/dep-linux-amd64 -o $(DEP_BIN)
	@cd $(DEP_BIN_DIR) && \
	echo "287b08291e14f1fae8ba44374b26a2b12eb941af3497ed0ca649253e21ba2f83  dep" > checksum && \
	sha256sum -c checksum
endif
	@chmod +x $(DEP_BIN)

$(VENDOR_DIR): Gopkg.toml
	@echo "checking dependencies with $(DEP_BIN_NAME)"
	@$(DEP_BIN) ensure -v 
		

# Code format/check
$(GOLINT_BIN):
	cd $(VENDOR_DIR)/github.com/golang/lint/golint && go build -v
$(GOCYCLO_BIN):
	cd $(VENDOR_DIR)/github.com/fzipp/gocyclo && go build -v

GOFORMAT_FILES := $(shell find  . -name '*.go' | grep -vEf .gofmt_exclude)

.PHONY: check-go-format
check-go-format: prebuild-check deps ## Exists with an error if there are files whose formatting differs from gofmt's
	@gofmt -s -l ${GOFORMAT_FILES} 2>&1 \
		| tee /tmp/gofmt-errors \
		| read \
	&& echo "ERROR: These files differ from gofmt's style (run 'make format-go-code' to fix this):" \
	&& cat /tmp/gofmt-errors \
	&& exit 1 \
	|| true


.PHONY: analyze-go-code 
analyze-go-code: deps golint gocyclo govet ## Run a complete static code analysis using the following tools: golint, gocyclo and go-vet.

golint: $(GOLINT_BIN)
	$(info >>--- RESULTS: GOLINT CODE ANALYSIS ---<<)
	@$(foreach d,$(GOANALYSIS_DIRS),$(GOLINT_BIN) $d 2>&1 | grep -vEf .golint_exclude || true;)

gocyclo: $(GOCYCLO_BIN)
	$(info >>--- RESULTS: GOCYCLO CODE ANALYSIS ---<<)
	@$(foreach d,$(GOANALYSIS_DIRS),$(GOCYCLO_BIN) -over 10 $d | grep -vEf .golint_exclude || true;)

govet:
	$(info >>--- RESULTS: GO VET CODE ANALYSIS ---<<)
	@$(foreach d,$(GOANALYSIS_DIRS),go tool vet --all $d/*.go 2>&1;)


.PHONY: format-go-code
format-go-code: prebuild-check ## Formats any go file that differs from gofmt's style
	@gofmt -s -l -w ${GOFORMAT_FILES}


# support for generating goa code
$(GOAGEN_BIN): $(VENDOR_DIR)
	cd $(VENDOR_DIR)/github.com/goadesign/goa/goagen && go build -v

.PHONY: generate-client
generate-client: $(GOAGEN_BIN)
	$(GOAGEN_BIN) client -d github.com/fabric8-services/fabric8-env/design --pkg env

# -------------------------------------------------------------------
# clean
# -------------------------------------------------------------------
# For the global "clean" target all targets in this variable will be executed
CLEAN_TARGETS =

CLEAN_TARGETS += clean-artifacts
.PHONY: clean-artifacts
clean-artifacts:
	-rm -rf $(INSTALL_PREFIX)

CLEAN_TARGETS += clean-object-files
.PHONY: clean-object-files
clean-object-files:
	go clean ./...

CLEAN_TARGETS += clean-generated
.PHONY: clean-generated
clean-generated:
	-rm -rf ./app
	-rm -rf ./swagger/
	-rm -f ./migration/sqlbindata.go
	-rm -f ./migration/sqlbindata_test.go
	-rm -rf ./env
	-rm -rf ./tool

CLEAN_TARGETS += clean-vendor
.PHONY: clean-vendor
clean-vendor:
	-rm -rf $(VENDOR_DIR)

.PHONY: clean-tmp
clean-tmp:
	-rm -rf $(TMP_PATH)

# Keep this "clean" target here after all `clean-*` sub tasks
.PHONY: clean
clean: $(CLEAN_TARGETS) ## Runs all clean-* targets.


# run in dev mode
.PHONY: dev
dev: prebuild-check deps generate docker-compose-up
	F8_DEVELOPER_MODE_ENABLED=true $(INSTALL_PREFIX)/env


.PHONY: docker-compose-up
docker-compose-up:
ifeq ($(UNAME_S),Darwin)
	@echo "Running docker-compose with macOS network settings"
	docker-compose -f docker-compose.macos.yml up -d db auth swagger_ui
else
	@echo "Running docker-compose with Linux network settings"
	docker-compose up -d db
endif


# build the binary executable (to ship in prod)
LDFLAGS=-ldflags "-X ${PACKAGE_NAME}/app.Commit=${COMMIT} -X ${PACKAGE_NAME}/app.BuildTime=${BUILD_TIME}"

.PHONY: build
build: prebuild-check deps generate $(BINARY_SERVER_BIN)

$(BINARY_SERVER_BIN): $(SOURCES)
ifeq ($(OS),Windows_NT)
	go build -v $(LDFLAGS) -o "$(shell cygpath --windows '$(BINARY_SERVER_BIN)')"
else
	go build -v $(LDFLAGS) -o $(BINARY_SERVER_BIN)
endif


.PHONY: migrate-database
## Compiles the server and runs the database migration with it
migrate-database: $(BINARY_SERVER_BIN)
	$(BINARY_SERVER_BIN) -migrateDatabase


.PHONY: generate
generate: app/controllers.go migration/sqlbindata.go migration/sqlbindata_test.go

app/controllers.go: prebuild-check $(DESIGNS) $(GOAGEN_BIN) $(VENDOR_DIR) ## Generate GOA sources. Only necessary after clean of if changed `design` folder.
	$(GOAGEN_BIN) app -d ${PACKAGE_NAME}/${DESIGN_DIR}
	$(GOAGEN_BIN) controller -d ${PACKAGE_NAME}/${DESIGN_DIR} -o controller/ --pkg controller --app-pkg ${PACKAGE_NAME}/app
	$(GOAGEN_BIN) gen -d ${PACKAGE_NAME}/${DESIGN_DIR} --pkg-path=github.com/fabric8-services/fabric8-common/goasupport/status --out app
	$(GOAGEN_BIN) gen -d ${PACKAGE_NAME}/${DESIGN_DIR} --pkg-path=github.com/fabric8-services/fabric8-common/goasupport/jsonapi_errors_helpers --out app
	$(GOAGEN_BIN) swagger -d ${PACKAGE_NAME}/${DESIGN_DIR}

migration/sqlbindata.go: $(GO_BINDATA_BIN)
	$(GO_BINDATA_BIN) \
		-o migration/sqlbindata.go \
		-pkg migration \
		-prefix migration/sql-files \
		-nocompress \
		migration/sql-files

migration/sqlbindata_test.go: $(GO_BINDATA_BIN)
	$(GO_BINDATA_BIN) \
		-o migration/sqlbindata_test.go \
		-pkg migration_test \
		-prefix migration/sql-test-files \
		-nocompress \
		migration/sql-test-files

$(GO_BINDATA_BIN): $(VENDOR_DIR)
	cd $(VENDOR_DIR)/github.com/jteeuwen/go-bindata/go-bindata && go build -v
