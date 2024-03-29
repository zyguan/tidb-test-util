package core

import (
	"bytes"
	"encoding/json"
	stderr "errors"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"

	. "github.com/google/go-jsonnet"
	. "github.com/zyguan/tidb-test-util/pkg/stmtflow"
)

var (
	ErrNotAsserted = stderr.New("no assertion")
)

type Assertion interface {
	Assert(actual History) error
	ExpectedText() (string, bool)
}

type Test struct {
	Name   string            `json:"name"`
	Test   []Stmt            `json:"test"`
	Labels map[string]string `json:"labels"`
	Expect json.RawMessage   `json:"expect"`
	Repeat int               `json:"repeat"`

	VersionConstraint string `json:"versionConstraint"`

	AssertMethod string      `json:"assertMethod"`
	Assertions   []Assertion `json:"-"`
}

func (t *Test) Assert(actual History) error {
	if len(t.Assertions) == 0 {
		return errors.WithStack(ErrNotAsserted)
	}
	if t.Repeat < 1 {
		t.Repeat = 1
	}
	for _, a := range t.Assertions {
		if err := a.Assert(actual); err != nil {
			return err
		}
	}
	return nil
}

func (t *Test) ExpectedText() (string, bool) {
	for _, a := range t.Assertions {
		if out, ok := a.ExpectedText(); ok {
			return out, ok
		}
	}
	return "", false
}

func (t *Test) ValidateVersion(ver string) error {
	if len(t.VersionConstraint) == 0 {
		return nil
	}
	v, err := semver.NewVersion(ver)
	if err != nil {
		return errors.Wrap(err, "invalid version")
	}
	c, err := semver.NewConstraint(t.VersionConstraint)
	if err != nil {
		return errors.Wrap(err, "invalid version constraint")
	}
	ok, errs := c.Validate(v)
	if len(errs) > 0 {
		return errors.Wrap(errs[0], "version mismatch")
	}
	if !ok {
		return errors.New("version mismatch")
	}
	return nil
}

type matchText struct {
	expect string
}

func (a *matchText) Assert(actual History) error {
	buf := new(bytes.Buffer)
	if err := actual.DumpText(buf, TextDumpOptions{Verbose: true}); err != nil {
		return errors.Wrap(err, "dump actual output")
	}
	if strings.TrimSpace(a.expect) != strings.TrimSpace(buf.String()) {
		return errors.New("result mismatch")
	}
	return nil
}

func (a *matchText) ExpectedText() (string, bool) {
	return a.expect, true
}

type matchHistory struct {
	expect History
}

func (a *matchHistory) Assert(actual History) error {
	if len(a.expect) != len(actual) {
		return errors.Errorf("expect %d events, got %d", len(a.expect), len(actual))
	}
	for i := range a.expect {
		if ok, msg := a.expect[i].EqualTo(actual[i], DefaultDigestOptions); !ok {
			return errors.Errorf("event#%d mismatch: %s", i, msg)
		}
	}
	return nil
}

func (a *matchHistory) ExpectedText() (string, bool) {
	buf := new(bytes.Buffer)
	if err := a.expect.DumpText(buf, TextDumpOptions{Verbose: true}); err != nil {
		return "", false
	}
	return buf.String(), true
}

type customAssertFn struct {
	path string
	name string
}

func (a *customAssertFn) Assert(actual History) error {
	vm := initVM(MakeVM())
	h, err := json.Marshal(actual)
	if err != nil {
		return errors.WithStack(err)
	}
	vm.ExtCode("actual", string(h))

	src := srcExpect
	src = strings.Replace(src, "__PATH__", a.path, 1)
	src = strings.Replace(src, "__NAME__", a.name, 1)

	js, err := vm.EvaluateAnonymousSnippet(":assert:", src)
	if err != nil {
		return errors.Wrap(err, "assert error")
	}
	msg := ""
	if err := json.Unmarshal([]byte(js), &msg); err != nil {
		return errors.Wrap(err, "unexpected assert outcome: "+js)
	}
	if len(msg) > 0 {
		return errors.New("assert message: " + msg)
	}
	return nil
}

func (a *customAssertFn) ExpectedText() (string, bool) {
	return "", false
}
