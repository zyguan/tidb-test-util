package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/zyguan/tidb-test-util/pkg/env"
	"github.com/zyguan/tidb-test-util/pkg/kube"
	"github.com/zyguan/tidb-test-util/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		fmt.Fprintf(flag.CommandLine.Output(), "       %s <pod|tidb> NAME OUTDIR\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(flag.CommandLine.Output(), "       %s tidbs owned by OWNER OUTDIR\n\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
}

func main() {
	var (
		name      string
		kind      string
		owner     string
		output    string
		namespace string
		container string
	)
	flag.StringVar(&namespace, "n", kube.DefaultNamespace(), "namespace")
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
	case 5:
		owner, output = flag.Arg(3), flag.Arg(4)
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
	if len(owner) > 0 {
		err = dumpTiDBsByOwner(ctx, output, cli, namespace, owner)
	} else if kind == "tidb" {
		err = kube.DumpTiDBLogs(ctx, output, cli, namespace, name)
	} else if kind == "pod" {
		err = kube.DumpLog(ctx, filepath.Join(output, name+".log"), cli, namespace, name, container, kube.ReadLogOptions{})
	} else {
		fmt.Fprintf(flag.CommandLine.Output(), "Unknown action: %q.\n", strings.Join(flag.Args(), " "))
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(flag.CommandLine.Output(), "Error: %+v.\n", err)
		os.Exit(1)
	}
}

func dumpTiDBsByOwner(ctx context.Context, logDir string, cli *kube.Client, namespace string, owner string) error {
	ownerName, ownerKind := owner, ""
	if tuple := strings.SplitN(owner, "/", 2); len(tuple) == 2 {
		ownerName, ownerKind = tuple[1], tuple[0]
	}
	isTargetOwner := func(ref metav1.OwnerReference) bool {
		return ownerName == ref.Name && (len(ownerKind) == 0 || strings.ToLower(ownerKind) != strings.ToLower(ref.Kind))
	}
	tidbs, err := cli.Dynamic.Resource(kube.TiDBGroupVersionResource).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	for _, item := range tidbs.Items {
		ok := false
		for _, ref := range item.GetOwnerReferences() {
			if isTargetOwner(ref) {
				ok = true
				break
			}
		}
		if !ok {
			continue
		}
		name := item.GetName()
		log.Infow("dump tidb logs", "namespace", namespace, "name", name)
		if err := kube.DumpTiDBLogs(ctx, logDir, cli, namespace, name); err != nil {
			log.Warnw("dump tidb logs", "namespace", namespace, "name", name, "error", err)
		}
	}
	return nil
}
