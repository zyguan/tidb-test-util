package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zyguan/tidb-test-util/pkg/env"
	"github.com/zyguan/tidb-test-util/pkg/kube"
	"github.com/zyguan/tidb-test-util/pkg/log"
)

var (
	Version   = "latest"
	BuildTime = "unknown"
)

func init() {
	env.LoadDotEnvOnce()
	log.UseGLog()
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s version\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(flag.CommandLine.Output(), "       %s <pod|tidb> [NAMESPACE] NAME OUTDIR\n\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
}

func main() {
	var (
		namespace string
		name      string
		kind      string
		output    string
		container string
	)
	flag.StringVar(&container, "c", "", "container name")
	flag.Parse()

	switch flag.NArg() {
	case 1:
		if action := flag.Arg(0); action == "version" {
			fmt.Printf("%s@%s (%s)\n", filepath.Base(os.Args[0]), Version, BuildTime)
			os.Exit(0)
		} else {
			fmt.Fprintf(flag.CommandLine.Output(), "Unknown command %q.\n", action)
			os.Exit(1)
		}
	case 3:
		kind, name, output = flag.Arg(0), flag.Arg(1), flag.Arg(2)
		namespace = kube.DefaultNamespace()
	case 4:
		kind, namespace, name, output = flag.Arg(0), flag.Arg(1), flag.Arg(2), flag.Arg(3)
	default:
		flag.Usage()
		os.Exit(1)
	}
	ctx := context.Background()
	cli, err := kube.DefaultClient()
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "Failed to construct kubernetes client.\n")
		os.Exit(1)
	}
	switch kind {
	case "pod":
		err = kube.DumpLog(ctx, filepath.Join(output, name+".log"), cli, namespace, name, container, kube.ReadLogOptions{})
	case "tidb":
		err = kube.DumpTiDBLogs(ctx, output, cli, namespace, name)
	default:
		fmt.Fprintf(flag.CommandLine.Output(), "Invalid kind of resource %q, please use one of [pod tidb].\n", kind)
		os.Exit(1)
	}
}
