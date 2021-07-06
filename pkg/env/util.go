package env

import (
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

var loadDotEnvOnce sync.Once

func LoadDotEnvOnce(cbs ...func()) { loadDotEnvOnce.Do(func() { LoadDotEnv(cbs...) }) }

func LoadDotEnv(cbs ...func()) bool {
	d, err := os.Getwd()
	if err != nil {
		return false
	}
	for len(d) > 0 {
		if err := godotenv.Load(path.Join(d, ".env")); err == nil {
			for _, cb := range cbs {
				cb()
			}
			return true
		}
		lastLen := len(d)
		d = path.Dir(d)
		if len(d) == lastLen {
			break
		}
	}
	return false
}

func Get(name string, optDefault ...string) string {
	if val := os.Getenv(name); len(val) > 0 {
		return val
	}
	if len(optDefault) > 0 {
		return optDefault[0]
	}
	return ""
}

func GetInt(name string, optDefault ...int) int {
	if val := os.Getenv(name); len(val) > 0 {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	if len(optDefault) > 0 {
		return optDefault[0]
	}
	return 0
}

func GetBool(name string, optDefault ...bool) bool {
	if val := strings.ToLower(os.Getenv(name)); len(val) > 0 {
		if val == "true" || val == "yes" || val == "on" || val == "1" {
			return true
		}
	}
	if len(optDefault) > 0 {
		return optDefault[0]
	}
	return false
}

func ListTestVars() []string {
	var lst []string
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "TEST_") {
			lst = append(lst, kv)
		}
	}
	return lst
}
