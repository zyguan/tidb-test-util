package kube

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/zyguan/tidb-test-util/pkg/fs"
	"github.com/zyguan/tidb-test-util/pkg/log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/exec"
)

func AsExecError(err error) (exec.ExitError, bool) {
	e, ok := errors.Cause(err).(exec.ExitError)
	return e, ok
}

type ReadLogOptions struct {
	Follow       bool
	Previous     bool
	Timestamps   bool
	TailLines    *int64
	LimitBytes   *int64
	SinceSeconds *int64
	SinceTime    *metav1.Time
}

func (opts ReadLogOptions) AsPodLogOptions(container string) *corev1.PodLogOptions {
	return &corev1.PodLogOptions{
		Container:    container,
		Follow:       opts.Follow,
		Previous:     opts.Previous,
		SinceSeconds: opts.SinceSeconds,
		SinceTime:    opts.SinceTime,
		Timestamps:   opts.Timestamps,
		TailLines:    opts.TailLines,
		LimitBytes:   opts.LimitBytes,
	}
}

func ReadLog(ctx context.Context, cli *Client, namespace string, name string, container string, options ReadLogOptions) (io.ReadCloser, error) {
	req := cli.CoreV1().Pods(namespace).GetLogs(name, options.AsPodLogOptions(container))
	out, err := req.Stream(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return out, nil
}

func DumpLog(ctx context.Context, logFile string, cli *Client, namespace string, name string, container string, options ReadLogOptions) error {
	r, err := ReadLog(ctx, cli, namespace, name, container, options)
	if err != nil {
		log.Warnw("read container log", "namespace", namespace, "name", name, "container", container, "error", err)
		return err
	}
	err = fs.DumpStream(logFile, r)
	if err != nil {
		log.Warnw("dump container log", "namespace", namespace, "name", name, "container", container, "error", err)
		return err
	}
	return nil
}

type ExecOptions struct {
	Command []string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	TTY     bool
}

func Exec(_ context.Context, cli *Client, namespace string, name string, container string, options ExecOptions) error {
	opts := &corev1.PodExecOptions{
		Container: container,
		Command:   options.Command,
		TTY:       options.TTY,
		Stdin:     options.Stdin != nil,
		Stdout:    options.Stdout != nil,
		Stderr:    options.Stderr != nil,
	}
	req := cli.CoreV1().RESTClient().Post().
		Resource("pods").SubResource("exec").
		Namespace(namespace).Name(name).
		VersionedParams(opts, scheme.ParameterCodec)
	executor, err := remotecommand.NewSPDYExecutor(cli.Config, http.MethodPost, req.URL())
	if err != nil {
		return errors.WithStack(err)
	}
	err = executor.Stream(remotecommand.StreamOptions{
		Stdin:  options.Stdin,
		Stdout: options.Stdout,
		Stderr: options.Stderr,
		Tty:    options.TTY,
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func ExecWithOutput(ctx context.Context, cli *Client, namespace string, name string, container string, tty bool, command ...string) (string, string, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	err := Exec(ctx, cli, namespace, name, container, ExecOptions{
		Command: command,
		Stdout:  stdout,
		Stderr:  stderr,
		TTY:     tty,
	})
	return stdout.String(), stderr.String(), err
}

func GetFile(ctx context.Context, cli *Client, namespace string, name string, container string, path string) (io.ReadCloser, error) {
	r, w := io.Pipe()
	go func() {
		err := Exec(ctx, cli, namespace, name, container, ExecOptions{Command: []string{"cat", path}, Stdout: w})
		if err != nil {
			w.CloseWithError(err)
		}
		w.Close()
	}()
	return r, nil
}

func DumpFile(ctx context.Context, logFile string, cli *Client, namespace string, name string, container string, path string) error {
	r, err := GetFile(ctx, cli, namespace, name, container, path)
	if err != nil {
		log.Warnw("read container file", "namespace", namespace, "name", name, "container", container, "path", path, "error", err)
		return err
	}
	err = fs.DumpStream(logFile, r)
	if err != nil {
		log.Warnw("dump container file", "namespace", namespace, "name", name, "container", container, "path", path, "error", err)
		return err
	}
	return nil
}

func ListFiles(ctx context.Context, cli *Client, namespace string, name string, container string, path string) ([]string, error) {
	stdout, stderr, err := ExecWithOutput(ctx, cli, namespace, name, container, false, "ls", "-a", path)
	if err != nil {
		errmsg := err.Error()
		if len(stderr) > 0 {
			errmsg += ": " + stderr
		}
		return nil, errors.New(errmsg)
	}
	all := strings.Split(stdout, "\n")
	selected := make([]string, 0, len(all))
	for _, f := range all {
		if len(f) == 0 || f == "." || f == ".." {
			continue
		}
		selected = append(selected, f)
	}
	return selected, nil
}
