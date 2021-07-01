package result

import (
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/zyguan/tidb-test-util/pkg/env"
)

var defaultResult *Result

func InitDefault() (*Result, error) {
	if defaultResult != nil {
		return defaultResult, nil
	}
	if id, ok := os.LookupEnv(env.TestResultID); ok {
		r, err := Get(id)
		if err != nil {
			return nil, err
		}
		defaultResult = r
		return defaultResult, nil
	}
	name, ok := os.LookupEnv(env.TestName)
	if !ok {
		name = "test-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	defaultResult = New(name, nil)
	return defaultResult, nil
}

func Report(conclusion Conclusion, output string) error {
	if defaultResult == nil {
		return errors.New("default result is nil")
	}
	return defaultResult.Report(conclusion, output)
}
