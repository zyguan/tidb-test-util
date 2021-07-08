package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/zyguan/tidb-test-util/pkg/kube"
	"github.com/zyguan/tidb-test-util/pkg/log"
	"github.com/zyguan/tidb-test-util/pkg/result"

	"github.com/spf13/pflag"
	"github.com/valyala/fasttemplate"
	"golang.org/x/crypto/ssh"

	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	keyPath    = "/tmp/key"
	jepsenPath = "jepsen.jar"
)

var global struct {
	name      string
	namespace string
	owner     string

	nodes int
	image string
	user  string

	jarURL string
	jobURL string
	labels map[string]string
}

func init() {
	pflag.StringVar(&global.name, "name", "jepsen", "")
	pflag.StringVar(&global.namespace, "namespace", kube.DefaultNamespace(), "")
	pflag.StringVar(&global.owner, "owner", "", "")

	pflag.IntVar(&global.nodes, "nodes", 5, "")
	pflag.StringVar(&global.image, "image", "hub.pingcap.net/qa/jepsen-node", "")
	pflag.StringVar(&global.user, "user", "root", "")

	pflag.StringVar(&global.jarURL, "jar-url", "http://fileserver.pingcap.net/download/pingcap/qa/tests/jepsen/jepsen-tidb.jar", "")
	pflag.StringVar(&global.jobURL, "job-url", "", "")
	pflag.StringToStringVarP(&global.labels, "label", "l", map[string]string{}, "")
}

type Job struct {
	Name        string            `json:"name"`
	ReportTo    string            `json:"reportTo"`
	FailureHook string            `json:"failureHook"`
	Labels      map[string]string `json:"labels"`
	FailFast    bool              `json:"failFast"`
	Tests       []Test            `json:"tests"`
}

type Test struct {
	Name       string            `json:"name"`
	Labels     map[string]string `json:"labels"`
	Resources  map[string]string `json:"resources"`
	Args       []string          `json:"args"`
	Timeout    int               `json:"timeout"`
	RetryLimit int               `json:"retryLimit"`
}

func (j *Job) output(ctx context.Context, t Test, r result.Conclusion, start int64, end int64) string {
	if r == result.Success || len(j.FailureHook) == 0 {
		return ""
	}
	tmpl, err := fasttemplate.NewTemplate(j.FailureHook, "{{", "}}")
	if err != nil {
		log.Warnf("invalid failure hook: %v", err)
		return "invalid failure hook"
	}
	buf := new(bytes.Buffer)
	_, err = tmpl.Execute(buf, map[string]interface{}{
		"name":       fmt.Sprintf("%s-%s-%d-%d", j.Name, t.Name, start, end),
		"store-path": "store",
	})
	if err != nil {
		log.Warnf("invalid failure hook: %v", err)
		return "invalid failure hook"
	}
	hookCmd := exec.CommandContext(ctx, "/bin/bash", "-c", buf.String())
	buf.Reset()
	hookCmd.Stdout = buf
	hookCmd.Stderr = buf
	log.Infof("[%s] failure hook: %s", t.Name, hookCmd.String())
	err = hookCmd.Run()
	if err != nil {
		log.Warnf("[%s] failure hook: %+v", t.Name, err)
		return "failed to execute hook command"
	}
	return buf.String()
}

func (j *Job) run(ctx context.Context) {
	for _, t := range j.Tests {
		log.Infof("[%s] start", t.Name)
		r := result.New(fmt.Sprintf("%s::%s", j.Name, t.Name), nil)
		c := t.run(ctx)
		log.Infof("[%s] %v", t.Name, c)
		output := j.output(ctx, t, c, r.StartedAt, time.Now().Unix())
		if len(j.ReportTo) > 0 {
			for k, v := range j.Labels {
				r.Labels[k] = v
			}
			for k, v := range t.Labels {
				r.Labels[k] = v
			}
			for k, v := range global.labels {
				r.Labels[k] = v
			}
			r.Report(c, output)
		}
		if j.FailFast {
			break
		}
	}
}

func (t *Test) run(ctx context.Context) result.Conclusion {
	logPrefix := "[" + t.Name + "] "
	log.Info(logPrefix + "init resources")
	try(os.RemoveAll("resources"))
	try(os.Mkdir("resources", 0755))
	for name, content := range t.Resources {
		try(ioutil.WriteFile("resources/"+name, []byte(content), 0644))
	}
	if err := os.RemoveAll("store"); err != nil && !os.IsNotExist(err) {
		log.Warnf(logPrefix+"clean store: %+v", err)
	}

	nodes := make([]string, global.nodes)
	for i := 0; i < global.nodes; i++ {
		nodes[i] = fmt.Sprintf("%s-%d.%s.%s", global.name, i, global.name, global.namespace)
	}
	args := []string{
		"-cp", "resources:" + jepsenPath,
		"tidb.core", "test",
		"--username=" + global.user,
		"--ssh-private-key=" + keyPath,
		"--nodes=" + strings.Join(nodes, ","),
	}
	args = append(args, t.Args...)
	if t.RetryLimit < 0 {
		t.RetryLimit = 0
	}
	for i := 0; i < t.RetryLimit+1; i++ {
		time.Sleep(time.Duration(i) * time.Second)
		testCtx := ctx
		if t.Timeout > 0 {
			testCtx, _ = context.WithTimeout(ctx, time.Duration(t.Timeout)*time.Second)
		}
		testCmd := exec.CommandContext(testCtx, "java", args...)
		log.Info(logPrefix + "exec: " + testCmd.String())
		err := testCmd.Run()
		if err == nil {
			return result.Success
		}
		log.Errorf(logPrefix+"%+v", err)
		if testCmd.ProcessState.ExitCode() == 1 {
			return result.Failure
		}
	}
	return result.Cancelled
}

func main() {
	pflag.Parse()

	ctx := context.Background()
	k := try(kube.DefaultClient()).(*kube.Client)

	genSSHKey()
	setupNodes(ctx, k)
	defer teardownNodes(ctx, k)
	waitNodesReady()
	downloadJepsen()
	runJob(ctx)
}

func genSSHKey() {
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		log.Info("generate ssh key")
		try(exec.Command("ssh-keygen", "-b", "2048", "-t", "rsa", "-m", "pem", "-f", keyPath, "-q", "-N", "").Run())
	} else {
		log.Info("reuse ssh key: " + keyPath)
	}
}

func readPublicKey() string {
	pk := try(ioutil.ReadFile(keyPath + ".pub")).([]byte)
	return string(pk)
}

func setupNodes(ctx context.Context, k *kube.Client) {
	yes := true
	labels := map[string]string{"name": global.name}
	replicas := int32(global.nodes)

	container := corev1.Container{
		Name:            "node",
		Image:           global.image,
		ImagePullPolicy: corev1.PullAlways,
		SecurityContext: &corev1.SecurityContext{
			Privileged: &yes,
		},
		Ports: []corev1.ContainerPort{
			{Name: "ssh", ContainerPort: 22},
		},
		Env: []corev1.EnvVar{
			{Name: "AUTHORIZED_KEYS", Value: readPublicKey()},
		},
	}
	nodes := appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: global.name,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    &replicas,
			ServiceName: global.name,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{container},
				},
			},
		},
	}
	ports := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: global.name,
		},
		Spec: corev1.ServiceSpec{
			Selector:  labels,
			ClusterIP: corev1.ClusterIPNone,
			Ports: []corev1.ServicePort{
				{Name: "ssh", Port: 22, TargetPort: intstr.FromInt(22)},
			},
		},
	}

	setOwner(&nodes.ObjectMeta)
	setOwner(&ports.ObjectMeta)
	try(k.AppsV1().StatefulSets(global.namespace).Create(ctx, &nodes, metav1.CreateOptions{}))
	try(k.CoreV1().Services(global.namespace).Create(ctx, &ports, metav1.CreateOptions{}))
}

func teardownNodes(ctx context.Context, k *kube.Client) {
	log.Info("teardown nodes")
	if err := k.AppsV1().StatefulSets(global.namespace).Delete(ctx, global.name, metav1.DeleteOptions{}); err != nil {
		log.Warnf("failed to destroy nodes: %v", err)
	}
	if err := k.CoreV1().Services(global.namespace).Delete(ctx, global.name, metav1.DeleteOptions{}); err != nil {
		log.Warnf("failed to destroy service: %v", err)
	}
}

func waitNodesReady() {
	privateKey := try(ioutil.ReadFile(keyPath)).([]byte)
	signer := try(ssh.ParsePrivateKey(privateKey)).(ssh.Signer)

	config := &ssh.ClientConfig{
		User:            global.user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	for i := 0; i < global.nodes; i++ {
		target := fmt.Sprintf("%s-%d.%s.%s:22", global.name, i, global.name, global.namespace)
		try(wait.PollImmediate(3*time.Second, 5*time.Minute, func() (done bool, err error) {
			log.Infof("ping ssh://%s@%s", global.user, target)
			cli, err := ssh.Dial("tcp", target, config)
			if err != nil {
				log.Infof("%s may be not ready: %+v", target, err)
				return false, nil
			}
			cli.Close()
			return true, nil
		}))
	}
}

func setOwner(meta *metav1.ObjectMeta) {
	if len(global.owner) == 0 {
		return
	}
	var ref metav1.OwnerReference
	try(json.Unmarshal([]byte(global.owner), &ref))
	meta.OwnerReferences = append(meta.OwnerReferences, ref)
}

func downloadJepsen() {
	log.Info("download jepsen.jar from " + global.jarURL)
	resp := try(http.Get(global.jarURL)).(*http.Response)
	defer resp.Body.Close()
	out := try(os.OpenFile(jepsenPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)).(*os.File)
	defer out.Close()
	try(io.Copy(out, resp.Body))
}

func runJob(ctx context.Context) {
	log.Info("download job spec from " + global.jobURL)
	var (
		input io.ReadCloser
		job   Job
	)
	if strings.HasPrefix(global.jobURL, "http://") || strings.HasPrefix(global.jobURL, "https://") {
		input = try(http.Get(global.jobURL)).(*http.Response).Body
	} else {
		input = try(os.Open(global.jobURL)).(*os.File)
	}
	defer input.Close()
	rawJob := try(ioutil.ReadAll(input)).([]byte)
	log.Info("job spec: " + string(rawJob))
	log.Info("runtime labels: " + string(try(json.Marshal(global.labels)).([]byte)))
	try(json.Unmarshal(rawJob, &job))
	job.run(ctx)
}

func try(xs ...interface{}) interface{} {
	if len(xs) == 0 {
		return nil
	}
	if err, ok := xs[len(xs)-1].(error); ok && err != nil {
		panic(err)
	}
	return xs[0]
}
