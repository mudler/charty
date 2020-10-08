BINDIR      := $(CURDIR)/bin
BINNAME     ?= charty

# go option
PKG        := ./...
TAGS       :=
TESTS      := .
TESTFLAGS  :=
LDFLAGS    := -w -s
GOFLAGS    :=
SRC        := $(shell find . -type f -name '*.go' -print)

GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_SHA    = $(shell git rev-parse --short HEAD)
GIT_TAG    = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
GIT_DIRTY  = $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")

override LDFLAGS += -X "github.com/mudler/charty/cmd/charty.BuildTime=$(shell date -u '+%Y-%m-%d %I:%M:%S %Z')"
override LDFLAGS += -X "github.com/mudler/charty/cmd/charty.BuildCommit=$(GIT_SHA)-$(GIT_DIRTY)"

.PHONY: all
all: build

# ------------------------------------------------------------------------------
#  build

.PHONY: build
build: $(BINDIR)/$(BINNAME)

$(BINDIR)/$(BINNAME): $(SRC)
	GO111MODULE=on go build $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o '$(BINDIR)'/$(BINNAME)
	chmod +x '$(BINDIR)'/$(BINNAME)


.PHONY: clean
clean:
	@rm -rf '$(BINDIR)'


.PHONY: info
info:
	 @echo "Version:           ${VERSION}"
	 @echo "Git Tag:           ${GIT_TAG}"
	 @echo "Git Commit:        ${GIT_COMMIT}"
	 @echo "Git Tree State:    ${GIT_DIRTY}"