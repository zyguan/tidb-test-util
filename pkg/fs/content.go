package fs

import (
	"fmt"
	"strings"
	"unicode"
)

var ReleaseBranches = []string{"master", "release-5.1", "release-5.0", "release-4.0", "release-3.1", "release-3.0"}

func (fs *Client) WhereIsComponent(name string, ref string) string {
	file := name + ".tar.gz"
	switch name {
	case "tidb", "tikv", "pd":
		file = name + "-server.tar.gz"
	case "ticdc":
		file = "ticdc-linux-amd64.tar.gz"
	}

	remote := fmt.Sprintf("builds/pingcap/%s/%s/centos7/%s", name, ref, file)
	if fs.Exists(remote) {
		return remote
	}
	remote = fmt.Sprintf("builds/pingcap/%s/pr/%s/centos7/%s", name, ref, file)
	if fs.Exists(remote) {
		return remote
	}
	if name == "tiflash" || name == "br" {
		for _, branch := range ReleaseBranches {
			remote = fmt.Sprintf("builds/pingcap/%s/%s/%s/centos7/%s", name, branch, ref, file)
			if fs.Exists(remote) {
				return remote
			}
		}
	}

	if !isCommit(ref) {
		refBytes, err := fs.ReadAll(fmt.Sprintf("refs/pingcap/%s/%s/sha1", name, ref))
		if err != nil {
			return ""
		}
		if ref := strings.TrimSpace(string(refBytes)); len(ref) > 0 {
			return fs.WhereIsComponent(name, ref)
		}
	}

	return ""
}

func isCommit(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, r := range s {
		if !unicode.In(r, unicode.ASCII_Hex_Digit) || unicode.IsUpper(r) {
			return false
		}
	}
	return true
}
