package result

import (
	"k8s.io/klog/v2/klogr"
)

var L = klogr.New().WithName("result")

const (
	EnvTestStoreEndpoint = "TEST_STORE_ENDPOINT"
	EnvTestResultID      = "TEST_RESULT_ID"
	EnvTestName          = "TEST_NAME"
	EnvTestLabels        = "TEST_LABELS"
	EnvTestLabelPrefix   = "TEST_LABEL__"
)
