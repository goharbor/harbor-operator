package exit

import (
	"github.com/sirupsen/logrus"
	"k8s.io/klog"
)

const SuccessExitCode int = 0

var exitCode = SuccessExitCode

func SetCode(value int) {
	exitCode = value
}

func GetCode() int {
	return exitCode
}

func Exit() {
	klog.Flush()
	logrus.Exit(GetCode())
}
