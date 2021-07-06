package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/zyguan/tidb-test-util/pkg/env"
	"github.com/zyguan/tidb-test-util/pkg/log"
	"github.com/zyguan/tidb-test-util/pkg/result"
)

var (
	Version   = "latest"
	BuildTime = "unknown"
)

func command() (*exec.Cmd, *os.File) {
	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	var out *os.File
	if outPath := env.Get(env.TestOutput); len(outPath) > 0 {
		os.MkdirAll(filepath.Base(outPath), 0755)
		f, err := os.OpenFile(outPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err == nil {
			out = f
			cmd.Stdout = io.MultiWriter(cmd.Stdout, out)
			cmd.Stderr = io.MultiWriter(cmd.Stderr, out)
		} else {
			log.Warnw("failed to open out file", "path", outPath, "error", err)
		}
	}
	return cmd, out
}

func output(err error, out *os.File) string {
	const size, sep = 1023, "\n---\n"
	msg := ""
	if err != nil {
		msg += err.Error()
	}
	if out != nil && len(msg)+len(sep) < size {
		bs, err := tail(out, int64(size-len(msg)-len(sep)))
		if err != nil {
			log.Warnw("failed to tail log", "error", err)
		} else if len(bs) > 0 {
			if len(msg) > 0 {
				msg += sep
			}
			msg += string(bs)
		}
	}
	return msg
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "USAGE: %s <command> [args]\n", filepath.Base(os.Args[0]))
		os.Exit(2)
	}
	log.Infow("test executor", "version", Version, "build-time", BuildTime, "result-store", env.Get(env.TestResultEndpoint))
	r, err := result.InitDefault()
	if err != nil {
		log.Warnw("failed init default result", "error", err)
	}
	cmd, out := command()
	exitCode := 0
	log.Infow("run test command", "cmd", cmd.String(), "env", env.ListTestVars())
	if err = cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			exitCode = 2
		}
	}
	if r.Report(result.ExitConclusion(exitCode), output(err, out)) == nil {
		log.Infow("report result", "data", r)
	}
	out.Close()
	os.Exit(exitCode)
}

func init() {
	env.LoadDotEnvOnce()
	log.UseGLog()
}

func tail(f *os.File, n int64) ([]byte, error) {
	fi, err := f.Stat()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	size := fi.Size()
	if n < size {
		size = n
	}
	if _, err = f.Seek(-size, io.SeekEnd); err != nil {
		return nil, errors.WithStack(err)
	}
	buf := make([]byte, size)
	if _, err = f.Read(buf); err != nil {
		return nil, errors.WithStack(err)
	}
	return buf, nil
}
