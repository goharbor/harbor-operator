package controllers

import (
	"fmt"
	"strings"
)

//go:generate controller-gen rbac:roleName="harbor-operator-role" output:artifacts:config="../config/rbac" paths="./..."

//go:generate stringer -type=Controller -linecomment
type Controller int

const (
	Core                      Controller = iota // core
	JobService                                  // jobservice
	Portal                                      // portal
	Registry                                    // registry
	RegistryController                          // registryctl
	ChartMuseum                                 // chartmuseum
	Exporter                                    // exporter
	NotaryServer                                // notaryserver
	NotarySigner                                // notarysigner
	Trivy                                       // trivy
	Harbor                                      // harbor
	HarborCluster                               // harborcluster
	HarborConfigurationCm                       // harborconfigurationcm
	HarborConfiguration                         // harborconfiguration
	HarborServerConfiguration                   // harborserverconfiguration
	PullSecretBinding                           // pullsecretbinding
	Namespace                                   // namespace
)

func (c Controller) GetFQDN() string {
	return fmt.Sprintf("%s.goharbor.io", strings.ToLower(c.String()))
}

func (c Controller) Label(suffix ...string) string {
	return c.LabelWithPrefix("", suffix...)
}

func (c Controller) LabelWithPrefix(prefix string, suffix ...string) string {
	var suffixString string
	if len(suffix) > 0 {
		suffixString = "/" + strings.Join(suffix, "-")
	}

	if prefix != "" {
		prefix = "." + prefix
	}

	return fmt.Sprintf("%s%s%s", prefix, c.GetFQDN(), suffixString)
}
