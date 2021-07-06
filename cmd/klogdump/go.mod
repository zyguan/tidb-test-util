module github.com/zyguan/tidb-test-util/cmd/klogdump

go 1.16

require (
	github.com/pkg/errors v0.9.1
	github.com/zyguan/tidb-test-util v0.0.0-00010101000000-000000000000
	k8s.io/apimachinery v0.20.0
)

replace github.com/zyguan/tidb-test-util => ../..
