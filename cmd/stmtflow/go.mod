module github.com/zyguan/tidb-test-util/cmd/stmtflow

go 1.16

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/antlr/antlr4/runtime/Go/antlr v0.0.0-20211028231423-7b32c9b169a2
	github.com/go-sql-driver/mysql v1.6.0
	github.com/google/go-jsonnet v0.17.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	github.com/zyguan/tidb-test-util v0.0.0-00010101000000-000000000000
)

replace github.com/zyguan/tidb-test-util => ../..
