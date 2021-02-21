package result

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type Conclusion string

const (
	Unknown  Conclusion = "unknown"
	Success  Conclusion = "success"
	Failure  Conclusion = "failure"
	Canceled Conclusion = "canceled"
	TimedOut Conclusion = "timed_out"
	Skipped  Conclusion = "skipped"
)

type Result struct {
	ID          string            `json:"id,omitempty"`
	Name        string            `json:"schedID,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	StartedAt   int64             `json:"startedAt,omitemptyt"`
	CompletedAt int64             `json:"completedAt,omitempty"`
	Conclusion  Conclusion        `json:"conclusion,omitempty"`
	Output      string            `json:"output,omitempty"`
}

func New(name string, labels map[string]string) *Result {
	r := Result{
		Name:      name,
		Labels:    make(map[string]string),
		StartedAt: time.Now().Unix(),
	}
	for k, v := range labels {
		r.Labels[k] = v
	}
	r.mergeEnvLabels()
	return &r
}

func Get(id string) (*Result, error) {
	var r Result
	apiURL := makeURL("/results/" + id)
	if len(apiURL) == 0 {
		return nil, errors.New("cannot get result from an empty endpoint")
	}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, errors.Wrapf(err, "do request %s %s", http.MethodGet, apiURL)
	}
	defer resp.Body.Close()
	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read response")
	}
	if resp.StatusCode >= 300 {
		return nil, errors.Errorf("unexpected response %s %q", resp.Status, string(raw))
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, errors.Wrap(err, "decode response body")
	}
	r.mergeEnvLabels()
	return &r, nil
}

func (r *Result) Report(conclusion Conclusion, output string) error {
	if r.CompletedAt == 0 {
		r.CompletedAt = time.Now().Unix()
	}
	r.Conclusion = conclusion
	r.Output = output
	err := r.Update()
	if err != nil {
		if isEmptyEndpointError(err) {
			L.V(1).Info(err.Error())
		} else {
			L.Error(err, "update result", "instance", r)
		}
	}
	return err
}

func (r *Result) Update() error {
	method, apiPath := http.MethodPost, "/results"
	if len(r.ID) > 0 {
		method, apiPath = http.MethodPatch, "/results/"+r.ID
	}
	apiURL := makeURL(apiPath)
	if len(apiURL) == 0 {
		return errors.New("skip report due to empty endpoint")
	}
	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(r); err != nil {
		return errors.Wrap(err, "encode result")
	}
	req, err := http.NewRequest(method, apiURL, body)
	if err != nil {
		return errors.Wrap(err, "create request")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "do request %s %s", method, apiURL)
	}
	defer resp.Body.Close()
	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "read response")
	}
	if resp.StatusCode >= 300 {
		return errors.Errorf("unexpected response %s %q", resp.Status, string(raw))
	}
	if err := json.Unmarshal(raw, r); err != nil {
		return errors.Wrap(err, "decode response body")
	}
	return nil
}

func (r *Result) mergeEnvLabels() {
	if r.Labels == nil {
		r.Labels = make(map[string]string)
	}
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, EnvTestLabelPrefix) {
			continue
		}
		kv := strings.SplitN(env, "=", 2)
		r.Labels[strings.TrimPrefix(kv[0], EnvTestLabelPrefix)] = kv[1]
	}
}

func makeURL(path string) string {
	if len(testStoreEndpoint) > 0 {
		return testStoreEndpoint + path
	}
	endpoint := os.Getenv(EnvTestStoreEndpoint)
	if len(endpoint) == 0 {
		return ""
	}
	return endpoint + path
}

func isEmptyEndpointError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "empty endpoint")
}
