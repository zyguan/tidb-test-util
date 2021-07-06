GO := CGO_ENABLED=0 go
PKGS := $(shell go list ./pkg/... | grep -v pkg/fs)

LDFLAGS += -X "main.Version=$(shell git describe --tags --dirty --always)"
LDFLAGS += -X "main.BuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%S')"

.PHONY: FORCE build clean test test-all

build: bin/dodo bin/stmtflow bin/klogdump bin/testexec

clean:
	@rm -rfv bin

test:
	go test -v $(PKGS)

test-all:
	go test -v ./pkg/...

bin/%: FORCE
	cd ./cmd/$* && $(GO) build -o ../../$@ -ldflags '$(LDFLAGS)' ./

upload-%: bin/%
	bin/dodo put /pingcap/qa/archives/util/$* bin/$*
