package harbor

import (
	"sync"

	"github.com/goharbor/harbor/src/lib/config/metadata"
	corev1 "k8s.io/api/core/v1"
)

type ConfigValue interface {
	GetValue() string
	GetValueFrom() *corev1.EnvVarSource
}

var (
	once    sync.Once
	envKeys map[string]string
)

func refreshEnvKeys() {
	envKeys = map[string]string{}

	for _, config := range metadata.ConfigList {
		envKeys[config.Name] = config.EnvKey
	}
}

func EnvVar(configName string, value ConfigValue) (corev1.EnvVar, error) {
	once.Do(refreshEnvKeys)

	envKey, ok := envKeys[configName]
	if !ok {
		return corev1.EnvVar{}, errNoConfigFound(configName)
	}

	return corev1.EnvVar{
		Name:      envKey,
		Value:     value.GetValue(),
		ValueFrom: value.GetValueFrom(),
	}, nil
}

func EnvVars(configs map[string]ConfigValue) ([]corev1.EnvVar, error) {
	envVars := []corev1.EnvVar{}

	for name, value := range configs {
		envVar, err := EnvVar(name, value)
		if err != nil {
			return nil, err
		}

		envVars = append(envVars, envVar)
	}

	return envVars, nil
}

type Value string

func (v Value) GetValue() string {
	return string(v)
}

func (v Value) GetValueFrom() *corev1.EnvVarSource {
	return nil
}

type ValueFrom corev1.EnvVarSource

func (v ValueFrom) GetValue() string {
	return ""
}

func (v ValueFrom) GetValueFrom() *corev1.EnvVarSource {
	source := corev1.EnvVarSource(v)

	return &source
}
