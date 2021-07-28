package kube

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	TiDBGroupVersionResource = schema.GroupVersionResource{Group: "pingcap.com", Version: "v1alpha1", Resource: "tidbclusters"}
)

func isLogFile(name string) bool {
	return strings.Contains(name, ".log") || name == "slowlog"
}

func DiscoverTiDBLogFiles(ctx context.Context, cli *Client, namespace string, name string, container string, dirs ...string) []string {
	if len(dirs) == 0 {
		switch container {
		case "tidb":
			dirs = []string{"/var/log/tidblog", "/var/log/tidb"}
		case "tikv":
			dirs = []string{"/var/lib/tikv/tikvlog", "/var/lib/tikv"}
		case "pd":
			dirs = []string{"/var/log/pdlog", "/var/lib/pd"}
		case "tiflash":
			dirs = []string{"/data0/logs"}
		}
	}
	var lst []string
	for _, dir := range dirs {
		files, err := ListFiles(ctx, cli, namespace, name, container, dir)
		if err != nil {
			continue
		}
		for _, file := range files {
			if isLogFile(file) {
				lst = append(lst, filepath.Join(dir, file))
			}
		}
	}
	return lst
}

func ListTiDBPods(ctx context.Context, cli *Client, namespace string, name string) ([]corev1.Pod, error) {
	selector := labels.SelectorFromSet(labels.Set{
		"app.kubernetes.io/instance":   name,
		"app.kubernetes.io/managed-by": "tidb-operator",
		"app.kubernetes.io/name":       "tidb-cluster",
	})
	lst, err := cli.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return lst.Items, nil
}

func DumpTiDBLogs(ctx context.Context, logDir string, cli *Client, namespace string, name string) error {
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		return errors.WithStack(err)
	}
	pods, err := ListTiDBPods(ctx, cli, namespace, name)
	if err != nil {
		return err
	}
	if len(pods) == 0 {
		return errors.New("no pod found for " + namespace + "/" + name)
	}

	for _, pod := range pods {
		component := pod.Labels["app.kubernetes.io/component"]
		if len(component) == 0 {
			continue
		}
		files := DiscoverTiDBLogFiles(ctx, cli, namespace, pod.Name, component)
		if len(files) == 0 {
			DumpLog(ctx, filepath.Join(logDir, pod.Name, component+".log"), cli, namespace, pod.Name, component, ReadLogOptions{})
		} else {
			dumpLog := true
			for _, file := range files {
				DumpFile(ctx, filepath.Join(logDir, pod.Name, filepath.Base(file)), cli, namespace, pod.Name, component, file)
				if filepath.Base(file) == component+".log" {
					dumpLog = false
				}
			}
			if dumpLog {
				DumpLog(ctx, filepath.Join(logDir, pod.Name, component+".log"), cli, namespace, pod.Name, component, ReadLogOptions{})
			}
		}
		if component == "tidb" {
			uid, err := GetUserID(ctx, cli, namespace, pod.Name, component)
			if err != nil {
				continue
			}
			DumpTarball(ctx, filepath.Join(logDir, pod.Name, "tmp-storage.tar.gz"), cli, namespace, pod.Name, component, fmt.Sprintf("/tmp/%d_tidb", uid))
		}
	}
	return nil
}
