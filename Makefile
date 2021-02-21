GO := CGO_ENABLED=0 go

.PHONY: FORCE build clean

build: bin/stmtflow

clean:
	@rm -rfv bin

bin/stmtflow: FORCE
	$(GO) build -o $@ ./cmd/stmtflow
