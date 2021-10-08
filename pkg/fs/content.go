package fs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"unicode"
)

func (fs *Client) WhereIsComponent(name string, ref string, branches ...string) string {
	file := name + ".tar.gz"
	switch name {
	case "tidb", "tikv", "pd":
		file = name + "-server.tar.gz"
	case "ticdc":
		file = "ticdc-linux-amd64.tar.gz"
	}

	if isCommit(ref) {
		if name == "tiflash" || name == "br" {
			for _, branch := range branches {
				remote := fmt.Sprintf("builds/pingcap/%s/%s/%s/centos7/%s", name, branch, ref, file)
				if fs.Exists(remote) {
					return remote
				}
			}
		}
		remote := fmt.Sprintf("builds/pingcap/%s/%s/centos7/%s", name, ref, file)
		if fs.Exists(remote) {
			return remote
		}
		remote = fmt.Sprintf("builds/pingcap/%s/pr/%s/centos7/%s", name, ref, file)
		if fs.Exists(remote) {
			return remote
		}
	} else {
		branch := ref
		if name == "br" || name == "tidb-lightning" {
			if isBRInTiDB(fs.Http, ref) {
				refBytes, err := fs.ReadAll(fmt.Sprintf("refs/pingcap/tidb/%s/sha1", ref))
				if ref := trimSHA1(refBytes); err == nil && len(ref) > 0 {
					return fs.WhereIsComponent("br", ref, branch)
				}
			}
			if name == "br" || isLightningInBR(fs.Http, ref) {
				refBytes, err := fs.ReadAll(fmt.Sprintf("refs/pingcap/br/%s/sha1", ref))
				if ref := trimSHA1(refBytes); err == nil && len(ref) > 0 {
					return fs.WhereIsComponent("br", ref, branch)
				}
			}
		}
		sha1Path := fmt.Sprintf("refs/pingcap/%s/%s/sha1", name, ref)
		refBytes, err := fs.ReadAll(sha1Path)
		if err != nil {
			return ""
		}
		if ref := trimSHA1(refBytes); len(ref) > 0 {
			return fs.WhereIsComponent(name, ref, branch)
		}
	}

	return ""
}

func trimSHA1(raw []byte) string {
	return string(bytes.TrimPrefix(bytes.TrimSpace(raw), []byte(`pr/`)))
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

func isBRInTiDB(cli *http.Client, ref string) bool {
	return existsOnGitHub(cli, ref, "pingcap/tidb", "/br/README.md")
}

func isLightningInBR(cli *http.Client, ref string) bool {
	return existsOnGitHub(cli, ref, "pingcap/br", "/cmd/tidb-lightning/main.go")
}

func existsOnGitHub(cli *http.Client, ref string, repo string, path string) bool {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s%s", repo, processRef(cli, ref, repo), path)
	resp, err := cli.Head(url)
	return err == nil && resp.StatusCode == http.StatusOK
}

func processRef(cli *http.Client, ref string, repo string) string {
	if !strings.HasPrefix(ref, "pr/") {
		return ref
	}
	resp, err := cli.Get("https://api.github.com/repos/" + repo + "/pulls/" + ref[3:])
	if err != nil {
		return ref
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ref
	}
	var body struct {
		Head struct {
			SHA string `json:"sha"`
		} `json:"head"`
	}
	if json.NewDecoder(resp.Body).Decode(&body) != nil {
		return ref
	}
	return body.Head.SHA
}
