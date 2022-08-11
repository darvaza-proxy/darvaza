.PHONY: all generate fmt install

GO ?= go
GOFMT ?= gofmt
GOFMT_FLAGS = -w -l -s
GOGENERATE_FLAGS = -v

PROJECTS = acme agent server shared

TEMPDIR ?= .tmp

all: get generate fmt build

$(TEMPDIR)/gen.mk: scripts/gen_mk.sh Makefile
	mkdir -p $(@D)
	$< $(PROJECTS) > $@~
	if cmp $@ $@~ 2> /dev/null; then rm $@~; else mv $@~ $@; fi

include $(TEMPDIR)/gen.mk

fmt: tidy
	@find . -name '*.go' | xargs -r $(GOFMT) $(GOFMT_FLAGS)

generate:
	@git grep -l '^//go:generate' | sed -n -e 's|\(.*\)/[^/]\+\.go$$|\1|p' | sort -u | while read d; do \
		git grep -l '^//go:generate' "$$d"/*.go | xargs -r $(GO) generate $(GOGENERATE_FLAGS); \
	done

install:
	$(GO) install -v ./...
