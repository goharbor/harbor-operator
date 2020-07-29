package controllers

//go:generate controller-gen rbac:roleName="manager-role" output:artifacts:config="../config/rbac" paths="./..."

//go:generate stringer -type=Controller -linecomment
type Controller int

const (
	Core               Controller = iota // core
	JobService                           // jobservice
	Portal                               // portal
	Registry                             // registry
	RegistryController                   // registryctl
	ChartMuseum                          // chartmuseum
	NotaryServer                         // notaryserver
	NotarySigner                         // notarysigner
	Clair                                // clair
	Trivy                                // trivy
	Harbor                               // harbor
)
