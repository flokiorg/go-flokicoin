PKG := github.com/flokiorg/go-flokicoin

LINT_PKG := github.com/golangci/golangci-lint/cmd/golangci-lint
GOACC_PKG := github.com/ory/go-acc
GOIMPORTS_PKG := golang.org/x/tools/cmd/goimports

GO_BIN := ${GOPATH}/bin
LINT_BIN := $(GO_BIN)/golangci-lint
GOACC_BIN := $(GO_BIN)/go-acc

LINT_COMMIT := v1.18.0
GOACC_COMMIT := 80342ae2e0fcf265e99e76bcc4efd022c7c3811b

DEPGET := cd /tmp && go get -v
GOBUILD := go build -v
GOINSTALL := go install -v 
DEV_TAGS := rpctest
GOTEST_DEV = go test -v  -tags=$(DEV_TAGS)
GOTEST := go test -v

GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

RM := rm -f
CP := cp
MAKE := make
XARGS := xargs -L 1

# Linting uses a lot of memory, so keep it under control by limiting the number
# of workers if requested.
ifneq ($(workers),)
LINT_WORKERS = --concurrency=$(workers)
endif

LINT = $(LINT_BIN) run -v $(LINT_WORKERS)

GREEN := "\\033[0;32m"
NC := "\\033[0m"
define print
	echo $(GREEN)$1$(NC)
endef

#? default: Run `make build`
default: build

#? all: Run `make build` and `make check`
all: build check

# ============
# DEPENDENCIES
# ============

$(LINT_BIN):
	@$(call print, "Fetching linter")
	$(DEPGET) $(LINT_PKG)@$(LINT_COMMIT)

$(GOACC_BIN):
	@$(call print, "Fetching go-acc")
	$(DEPGET) $(GOACC_PKG)@$(GOACC_COMMIT)

#? goimports: Install goimports
goimports:
	@$(call print, "Installing goimports.")
	$(DEPGET) $(GOIMPORTS_PKG)

# ============
# INSTALLATION
# ============

#? build: Build all binaries, place them in project directory
build:
	@$(call print, "Building all binaries")
	$(GOBUILD) $(PKG)
	$(GOBUILD) $(PKG)/cmd/lokid-cli
	$(GOBUILD) $(PKG)/cmd/gencerts
	$(GOBUILD) $(PKG)/cmd/findcheckpoint
	$(GOBUILD) $(PKG)/cmd/addblock

#? install: Install all binaries, place them in $GOPATH/bin
install:
	@$(call print, "Installing all binaries")
	$(GOINSTALL) $(PKG)
	$(GOINSTALL) $(PKG)/cmd/lokid-cli
	$(GOINSTALL) $(PKG)/cmd/gencerts
	$(GOINSTALL) $(PKG)/cmd/findcheckpoint
	$(GOINSTALL) $(PKG)/cmd/addblock

#? release-install: Install lokid and lokid-cli release binaries, place them in $GOPATH/bin
release-install:
	@$(call print, "Installing lokid and lokid-cli release binaries")
	env CGO_ENABLED=0 $(GOINSTALL) -trimpath -ldflags="-s -w -buildid=" $(PKG)
	env CGO_ENABLED=0 $(GOINSTALL) -trimpath -ldflags="-s -w -buildid=" $(PKG)/cmd/lokid-cli

# =======
# TESTING
# =======

#? check: Run `make unit`
check: unit

#? unit: Run unit tests
unit:
	@$(call print, "Running unit tests.")
	$(GOTEST_DEV) ./... -test.timeout=20m
	cd crypto; $(GOTEST_DEV) ./... -test.timeout=20m
	cd chainutil; $(GOTEST_DEV) ./... -test.timeout=20m
	cd chainutil/psbt; $(GOTEST_DEV) ./... -test.timeout=20m

#? unit-cover: Run unit coverage tests
unit-cover: $(GOACC_BIN)
	@$(call print, "Running unit coverage tests.")
	$(GOACC_BIN) ./...
	
	# We need to remove the /v2 pathing from the module to have it work
	# nicely with the CI tool we use to render live code coverage.
	cd crypto; $(GOACC_BIN) ./...; sed -i.bak 's/v2\///g' coverage.txt

	cd chainutil; $(GOACC_BIN) ./...

	cd chainutil/psbt; $(GOACC_BIN) ./...

#? unit-race: Run unit race tests
unit-race:
	@$(call print, "Running unit race tests.")
	env CGO_ENABLED=1 GORACE="history_size=7 halt_on_errors=1" $(GOTEST) -race -test.timeout=20m ./...
	cd crypto; env CGO_ENABLED=1 GORACE="history_size=7 halt_on_errors=1" $(GOTEST) -race -test.timeout=20m ./...
	cd chainutil; env CGO_ENABLED=1 GORACE="history_size=7 halt_on_errors=1" $(GOTEST) -race -test.timeout=20m ./...
	cd chainutil/psbt; env CGO_ENABLED=1 GORACE="history_size=7 halt_on_errors=1" $(GOTEST) -race -test.timeout=20m ./...

# =========
# UTILITIES
# =========

#? fmt: Fix imports and formatting source
fmt: goimports
	@$(call print, "Fixing imports.")
	goimports -w $(GOFILES_NOVENDOR)
	@$(call print, "Formatting source.")
	gofmt -l -w -s $(GOFILES_NOVENDOR)

#? lint: Lint source
lint: $(LINT_BIN)
	@$(call print, "Linting source.")
	$(LINT)

#? clean: Clean source
clean:
	@$(call print, "Cleaning source.$(NC)")
	$(RM) coverage.txt crypto/coverage.txt chainutil/coverage.txt chainutil/psbt/coverage.txt
	
#? tidy-module: Run 'go mod tidy' for all modules
tidy-module:
	echo "Running 'go mod tidy' for all modules"
	scripts/tidy_modules.sh

.PHONY: all \
	default \
	build \
	check \
	unit \
	unit-cover \
	unit-race \
	fmt \
	lint \
	clean

#? help: Get more info on make commands
help: Makefile
	@echo " Choose a command run in lokid:"
	@sed -n 's/^#?//p' $< | column -t -s ':' |  sort | sed -e 's/^/ /'

.PHONY: help


# =========
# docker
# =========

build-image:
	docker build . -f ./docker/Dockerfile -t flokiorg/go-flokicoin

test-image:
	docker run --rm -it flokiorg/go-flokicoin


# =========
# temp
# =========

dir ?= .
run ?= .

debug:
	dlv debug $(dir)  --headless --listen=:2345  --api-version=2 --  $(run)

debugtest:
	dlv test $(dir)  --headless --listen=:2345  --api-version=2 -- -test.run $(run)

test:
	go test -v -count=1 $(dir) -run $(run)

testexport:
	go run ./cmd/testexport -v \
		-d ./_test/network/blocks_ffldb \
		-o ./blockchain/testdata \
		-f $(file).dat.bz2 \
		-r 0-10000
