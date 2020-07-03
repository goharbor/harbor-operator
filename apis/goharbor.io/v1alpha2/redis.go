package v1alpha2

//go:generate stringer -type=ComponentWithRedis -linecomment
type ComponentWithRedis int

const (
	// Order matters since the iota value
	// is the default redis index to use.
	CoreRedis        ComponentWithRedis = iota // core
	JobServiceRedis                            // jobservice
	RegistryRedis                              // registry
	ChartMuseumRedis                           // chartmuseum
	ClairRedis                                 // clair
	TrivyRedis                                 // trivy
)

func (r ComponentWithRedis) Index() int {
	return int(r)
}
