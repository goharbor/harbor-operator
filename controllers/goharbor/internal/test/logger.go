package test

import (
	"io"

	"github.com/sirupsen/logrus"
	"k8s.io/klog"
)

func ConfigureLoggers(output io.Writer) {
	logrus.SetOutput(output)
	klog.SetOutput(output)
}
