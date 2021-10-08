package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhereIsComponent(t *testing.T) {
	cli := Default()
	// no need to auth for `whereis`
	for _, name := range []string{"tidb", "tikv", "pd", "br", "ticdc", "tiflash", "tidb-lightning"} {
		for _, branch := range []string{"master", "release-5.2", "release-5.1", "release-4.0"} {
			t.Run(name+"-"+branch, func(t *testing.T) {
				url := cli.WhereIsComponent(name, branch)
				assert.NotEmpty(t, url)
				t.Log(t.Name(), url)
			})
		}
	}
}
