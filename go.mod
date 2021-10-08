module github.com/zyguan/tidb-test-util

go 1.16

require (
	github.com/go-sql-driver/mysql v1.6.0
	github.com/joho/godotenv v1.3.0
	github.com/olekukonko/tablewriter v0.0.5
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/zyguan/sqlz v0.0.0-20211008183028-44ff42cf1df2
	go.uber.org/zap v1.18.1
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	k8s.io/api v0.20.0
	k8s.io/apimachinery v0.20.0
	k8s.io/client-go v0.20.0
)

//replace github.com/zyguan/sqlz => ../sqlz
