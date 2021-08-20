BIN_NAME 	= containerd-healthcheck
ROOT_DIR 	= $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
BIN_DIR 	= $(ROOT_DIR)/bin
BIN_PATH 	= $(ROOT_DIR)/bin/$(BIN_NAME)
CMD_PATH 	= $(ROOT_DIR)/cmd/$(BIN_NAME)
VERSION 	= $(shell cat $(ROOT_DIR)/VERSION)
GIT_BRANCH 	= $(shell git rev-parse --abbrev-ref HEAD)
GIT_COMMIT 	= $(shell git rev-parse HEAD)
BUILD_DATE 	= $(shell date +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS 	= "-X main.commit=$(GIT_COMMIT) -X main.version=$(VERSION) -X main.date=$(BUILD_DATE)"
PORT 		= "9891"
ENV 		= development

export PATH := $(PATH):$(ROOT_DIR)/.bin

.PHONY: tidy
tidy:
	@(go mod tidy)

.PHONY: setup
setup: tidy
	@(./scripts/install-go-release.sh "goreleaser/goreleaser")

.PHONY: build
build: setup
	@(echo "-> Compiling packages")
	GOOS=linux go build -ldflags $(LDFLAGS) -o $(BIN_PATH) $(CMD_PATH)/main.go
	@(echo "-> Binary created at '$(BIN_PATH)'")

.PHONY: run
run:
	@($(BIN_PATH) --addr ":$(PORT)" --env $(ENV))

.PHONY: test -v
test:
	@(cd test && go test -v)

.PHONY: test-update
test-update:
	@(cd test && go test -update)

.PHONY: docker-build
docker-build:
	@(docker build -t $(BIN_NAME) .)

.PHONY: docker-run
docker-run:
	@(docker run -p 8080:$(PORT) -e "ENV=$(ENV)" -it $(BIN_NAME):latest)

.PHONY: release
release:
ifeq ($(GIT_BRANCH),master)
	@(echo "-> Creating tag '$(VERSION)'")
	@(git tag $(VERSION))
	@(echo "-> Pushing tag '$(VERSION)'")
	@(git push origin $(VERSION))
	@(echo "-> Releasing to remote repository")
	@(goreleaser --rm-dist)
else
	@echo "You need to be in branch master"
endif

.PHONY: release-snapshot
release-snapshot:
	@(goreleaser release --skip-publish --snapshot --rm-dist)
