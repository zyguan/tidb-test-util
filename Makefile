GO := CGO_ENABLED=0 go

LDFLAGS += -X "main.Version=$(shell git describe --tags --dirty --always)"
LDFLAGS += -X "main.BuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%S')"

.PHONY: FORCE build clean

build: bin/stmtflow bin/klogdump

clean:
	@rm -rfv bin

bin/%: FORCE
	cd ./cmd/$* && $(GO) build -o ../../$@ -ldflags '$(LDFLAGS)' ./
