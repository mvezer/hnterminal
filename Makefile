# enabling gofmt and golint since by default they're not enabled by golangci-lint
VERSION           = $(shell go version)
LINTER			  = golangci-lint run -v $(LINTER_FLAGS) --exclude-use-default=false --timeout $(LINTER_DEADLINE)
LINTER_DEADLINE	  = 30s
LINTER_FLAGS ?=
UNIT_TEST_PACKAGES := $(shell go list ./... | grep -i challenges)

GO_FLAGS        ?=
GO_FLAGS        += --ldflags 'extldflags="-static"'

ifneq (,$(findstring darwin/arm,$(VERSION)))
    GO_FLAGS += -tags dynamic
endif
ifneq (,$(wildcard /etc/alpine-release))
    GO_FLAGS += -tags musl
LINTER_FLAGS += --build-tags=musl
endif
# ifeq (run,$(firstword $(MAKECMDGOALS)))
#   # RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
#   RUN_ARGS := $(wordlist)
#   $(eval $(RUN_ARGS):;@:)
# endif
ifeq (test,$(firstword $(MAKECMDGOALS)))
  TEST_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  $(eval $(TEST_ARGS):;@:)
endif
INTEGRATION_TEST_ROOT		= ./test/integration
UNIT_TEST_PACKAGES			= $(shell go list ./... | grep -v $(INTEGRATION_TEST_ROOT))

PROTO_DIR = protos
PROTO_FILES = $(wildcard $(PROTO_DIR)/*.proto)

build:
	go build $(GO_FLAGS) -v -o bin/hnterminal internal/main.go
	@echo "** build complete: bin/hnterminal **"

run:
	go run $(GO_FLAGS) internal/main.go

test:
	# go test -v -race -coverprofile=coverage.txt -covermode=atomic $(UNIT_TEST_PACKAGES)
	go test -v "hnterminal/challenges/day$(TEST_ARGS)"

.PHONY: build run
