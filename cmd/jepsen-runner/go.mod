module github.com/zyguan/tidb-test-util/cmd/jepsen-runner

go 1.16

require (
	github.com/spf13/pflag v1.0.5
	github.com/valyala/fasttemplate v1.2.1
	github.com/zyguan/tidb-test-util v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0
	k8s.io/api v0.20.0
	k8s.io/apimachinery v0.20.0
)

replace github.com/zyguan/tidb-test-util => ../..
