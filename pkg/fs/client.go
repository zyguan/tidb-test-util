package fs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	ExtPartial  = ".partial"
	ExtChecksum = ".sha256"

	TokenFile = "/tmp/.fbtoken"
)

type SizedReader interface {
	io.Reader
	Size() int64
}

func String(s string) SizedReader { return Buffer(bytes.NewBufferString(s)) }

type BufferReader struct{ *bytes.Buffer }

func Buffer(buffer *bytes.Buffer) *BufferReader { return &BufferReader{Buffer: buffer} }

func (r *BufferReader) Size() int64 { return int64(r.Len()) }

type FileReader struct{ *os.File }

func File(file *os.File) *FileReader { return &FileReader{File: file} }

func (r *FileReader) Size() int64 {
	if stat, err := r.Stat(); err != nil || stat.IsDir() {
		return -1
	} else {
		return stat.Size()
	}
}

type FileInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
	Ext  string `json:"extension"`
	Dir  bool   `json:"isDir"`
	Size int    `json:"size"`
	Mode int    `json:"mode"`

	Modified time.Time `json:"modified"`

	Items []FileInfo `json:"items"`

	Checksum struct {
		SHA256 string `json:"sha256"`
	} `json:"checksums"`
}

func Default() *Client {
	token, _ := ioutil.ReadFile(TokenFile)
	return &Client{
		Host:    "fileserver.pingcap.net",
		Port:    80,
		FBHost:  "filebrowser.pingcap.net",
		FBPort:  443,
		FBUser:  "guest",
		FBPass:  "guest",
		FBToken: string(token),

		Http: http.DefaultClient,
	}
}

type Client struct {
	Host    string
	FBHost  string
	FBUser  string
	FBPass  string
	FBToken string

	Port   int
	FBPort int

	Http *http.Client
}

func (fs *Client) Auth() error {
	resp, err := fs.doFBReq(http.MethodPost, "/api/renew", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		token := string(body)
		if token == fs.FBToken {
			return nil
		}
		fs.FBToken = token
	} else if resp.StatusCode == http.StatusForbidden {
		fs.FBToken = ""
		buf := new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(map[string]string{
			"username":  fs.FBUser,
			"password":  fs.FBPass,
			"recaptcha": "",
		})
		if err != nil {
			return err
		}
		resp, err := fs.doFBReq(http.MethodPost, "/api/login", buf)
		if err != nil {
			return err
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("login failed: %s", strings.TrimSpace(string(body)))
		}
		fs.FBToken = string(body)
	} else {
		return errors.New("renew failed: " + string(body))
	}
	ioutil.WriteFile(TokenFile, []byte(fs.FBToken), 0644)
	return nil
}

func (fs *Client) GetFile(remote string, local string) error {
	f, err := os.OpenFile(local, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if fs.Exists(remote + ExtPartial) {
		return fmt.Errorf("skip download because %s exists", remote+ExtPartial)
	}
	checksum, err := fs.ReadAll(remote + ExtChecksum)
	if err != nil {
		return err
	}
	r, err := fs.Read(remote)
	if err != nil {
		return err
	}
	defer r.Close()

	h := sha256.New()
	_, err = io.Copy(io.MultiWriter(h, f), r)
	if err != nil {
		return err
	}
	if s := hex.EncodeToString(h.Sum(nil)); s != string(checksum) {
		return fmt.Errorf("checksum mismatch: expect %s but got %s", checksum, s)
	}
	return nil
}

func (fs *Client) PutFile(remote string, local SizedReader) error {
	hash, err := fs.Write(remote, local)
	if err != nil {
		return err
	}
	_, err = fs.Write(remote+ExtChecksum, String(hash))
	if err != nil {
		return err
	}
	return nil
}

func (fs *Client) DelFile(remote string, unsafe bool) error {
	fs.Delete(remote + ExtPartial)
	if err := fs.Delete(remote + ExtChecksum); err != nil && !unsafe {
		return err
	}
	return fs.Delete(remote)
}

func (fs *Client) MoveFile(from string, to string) error {
	return fs.relocateFile(fs.Rename, from, to)
}

func (fs *Client) CopyFile(from string, to string) error {
	return fs.relocateFile(fs.Copy, from, to)
}

func (fs *Client) Exists(remote string) bool {
	resp, err := fs.Http.Head(fs.DownloadURL(remote))
	if err != nil {
		return false
	}
	return resp.StatusCode == http.StatusOK
}

func (fs *Client) Write(remote string, content SizedReader) (string, error) {
	// file-server use resty.upload to parse request body, which doesn't support a chunked request,
	// thus the content is required to be sized (otherwise golang http will send a chunked request)
	name := filepath.Base(remote)
	done := make(chan error, 1)
	r, w := io.Pipe()
	hw := sha256.New()
	mw := multipart.NewWriter(w)
	go func() {
		var err error
		defer func() {
			if err != nil {
				w.CloseWithError(err)
				done <- err
			} else {
				err = mw.Close()
				if err != nil {
					w.CloseWithError(err)
					done <- err
				} else {
					w.Close()
					done <- nil
				}
			}
		}()
		pw, err := mw.CreateFormFile(remote, name)
		if err != nil {
			return
		}
		_, err = io.Copy(io.MultiWriter(hw, pw), content)
	}()

	req, err := http.NewRequest(http.MethodPost, fs.UploadURL(), r)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.ContentLength = inferContentLength(remote, name, content.Size())
	resp, err := fs.Http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if err = <-done; err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("failed to write %s: %s", remote, strings.TrimSpace(string(body)))
	}
	return hex.EncodeToString(hw.Sum(nil)), nil
}

func (fs *Client) Read(remote string) (io.ReadCloser, error) {
	resp, err := fs.Http.Get(fs.DownloadURL(remote))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to read %s: %s", remote, strings.TrimSpace(string(body)))
	}
	return resp.Body, nil
}

func (fs *Client) ReadAll(remote string) ([]byte, error) {
	r, err := fs.Read(remote)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return ioutil.ReadAll(r)
}

func (fs *Client) Stat(remote string, checksum bool) (*FileInfo, error) {
	if checksum {
		remote += "?checksum=sha256"
	}
	resp, err := fs.doFBResReq(http.MethodGet, remote, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to stat %s: %s", remote, strings.TrimSpace(string(body)))
	}

	var infos FileInfo
	if err = json.NewDecoder(resp.Body).Decode(&infos); err != nil {
		return nil, err
	}
	infos.Path = strings.TrimPrefix(infos.Path, "/download")
	for i := range infos.Items {
		infos.Items[i].Path = strings.TrimPrefix(infos.Items[i].Path, "/download")
	}
	return &infos, nil
}

func (fs *Client) Delete(remote string) error {
	resp, err := fs.doFBResReq(http.MethodDelete, remote, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to delete %s: %s", remote, strings.TrimSpace(string(body)))
	}
	return nil
}

func (fs *Client) Rename(from string, to string) error {
	return fs.relocate("rename", from, to)
}

func (fs *Client) Copy(from string, to string) error {
	return fs.relocate("copy", from, to)
}

func (fs *Client) relocate(action string, from string, to string) error {
	vals := url.Values{}
	vals.Set("action", action)
	vals.Set("destination", fbPath(to))
	resp, err := fs.doFBResReq(http.MethodPatch, from+"?"+vals.Encode(), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to %s %s to %s: %s", action, from, to, strings.TrimSpace(string(body)))
	}
	return nil
}

func (fs *Client) relocateFile(action func(string, string) error, from string, to string) error {
	info, err := fs.Stat(from, false)
	if err != nil {
		return err
	}
	if info.Dir {
		return action(from, to)
	}
	if err = action(from+ExtChecksum, to+ExtChecksum); err != nil {
		return err
	}
	return action(from, to)
}

func (fs *Client) UploadURL() string {
	return mkURL(fs.Host, fs.Port, "/upload", false)
}

func (fs *Client) DownloadURL(remote string) string {
	return mkURL(fs.Host, fs.Port, pathJoin("download", remote), false)
}

func (fs *Client) Format(fi *FileInfo, t string) string {
	switch t {
	case "name":
		return fi.Name
	case "path":
		return fi.Path
	case "detail":
		if fi.Dir {
			return fmt.Sprintf("D %11d\t%s\t%s", fi.Size, fi.Modified.Format(time.RFC3339), fi.Path)
		} else {
			return fmt.Sprintf("F %11d\t%s\t%s", fi.Size, fi.Modified.Format(time.RFC3339), fi.Path)
		}
	case "url":
		if fi.Dir {
			return mkURL(fs.FBHost, fs.FBPort, pathJoin("files", fi.Path), true)
		} else {
			return fs.DownloadURL(fi.Path)
		}
	default:
		return fi.Name
	}
}

func (fs *Client) doFBResReq(method string, remote string, body io.Reader) (*http.Response, error) {
	return fs.doFBReq(method, pathJoin("/api/resources", fbPath(remote)), body)
}

func (fs *Client) doFBReq(method string, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, mkURL(fs.FBHost, fs.FBPort, path, true), body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if len(fs.FBToken) > 0 {
		req.Header.Set("X-Auth", fs.FBToken)
	}
	return fs.Http.Do(req)
}

func mkURL(host string, port int, path string, tls bool) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if port != 80 {
		host += ":" + strconv.Itoa(port)
	}
	scheme := "http"
	if tls {
		scheme = "https"
	}
	return scheme + "://" + host + path
}

func pathJoin(elems ...string) string {
	prefix := ""
	if len(elems) > 0 && strings.HasPrefix(elems[0], "/") {
		prefix = "/"
	}
	for i := range elems {
		elems[i] = strings.Trim(elems[i], "./")
	}
	return prefix + strings.Join(elems, "/")
}

func fbPath(path string) string {
	return pathJoin("/download", path)
}

func inferContentLength(remote string, name string, size int64) int64 {
	if size < 0 {
		return 0
	}
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	w.CreateFormFile(remote, name)
	w.Close()
	return int64(buf.Len()) + size
}
