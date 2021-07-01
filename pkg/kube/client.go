package kube

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/zyguan/tidb-test-util/pkg/env"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	kubernetes.Interface
	Dynamic dynamic.Interface
	Config  *rest.Config
}

func DefaultClient() (*Client, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", env.Get(clientcmd.RecommendedConfigPathEnvVar))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	cli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &Client{cli, dyn, cfg}, nil
}

func DefaultNamespace() string {
	if ns, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		return string(ns)
	}
	return env.Get(env.TestNamespace, "default")
}
