package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhereIsComponent(t *testing.T) {
	cli := Default()
	// no need to auth for `whereis`

	for _, tt := range []struct {
		name string
		ref  string
	}{
		{"tidb", "master"},
		{"tikv", "master"},
		{"pd", "master"},
		{"br", "master"},
		{"ticdc", "master"},
		{"tiflash", "master"},
		{"tidb-lightning", "master"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			url := cli.WhereIsComponent(tt.name, tt.ref)
			assert.NotEmpty(t, url)
			t.Log(t.Name(), url)
		})
	}
}
