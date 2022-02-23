GO := CGO_ENABLED=0 go
PKGS := $(shell go list ./pkg/... | grep -v pkg/fs)
MINIO_CLIENT ?= mc
MINIO_UPLOAD_PATH ?= idc/tp-team/tools

LDFLAGS += -X "main.Version=$(shell git describe --tags --dirty --always)"
LDFLAGS += -X "main.BuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%S')"

.PHONY: FORCE build clean test test-all

build: bin/dodo bin/jepsen-runner bin/klogdump bin/stmtflow bin/testexec

clean:
	@rm -rfv bin

test:
	go test -v $(PKGS)

test-all:
	go test -v ./pkg/...

bin/%: FORCE
	cd ./cmd/$* && $(GO) mod tidy
	cd ./cmd/$* && $(GO) build -o ../../$@ -ldflags '$(LDFLAGS)' ./

upload-%: bin/%
	$(MINIO_CLIENT) cp bin/$* $(MINIO_UPLOAD_PATH)/$*
