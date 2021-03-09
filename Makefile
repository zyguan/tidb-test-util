GO := CGO_ENABLED=0 go

LDFLAGS += -X "github.com/zyguan/tidb-test-util/cmd/stmtflow/command.Version=$(shell git describe --tags --dirty --always)"
LDFLAGS += -X "github.com/zyguan/tidb-test-util/cmd/stmtflow/command.BuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%S')"

.PHONY: FORCE build clean

build: bin/stmtflow

clean:
	@rm -rfv bin

bin/stmtflow: FORCE
	$(GO) build -ldflags '$(LDFLAGS)' -o $@ ./cmd/stmtflow
