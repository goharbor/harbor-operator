package v1alpha2

type ComponentWithRedis Component

const (
	CoreRedis        = ComponentWithRedis(CoreComponent)
	JobServiceRedis  = ComponentWithRedis(JobServiceComponent)
	RegistryRedis    = ComponentWithRedis(RegistryComponent)
	ChartMuseumRedis = ComponentWithRedis(ChartMuseumComponent)
	ClairRedis       = ComponentWithRedis(ClairComponent)
	TrivyRedis       = ComponentWithRedis(TrivyComponent)
)

func (r ComponentWithRedis) Index() int {
	return map[ComponentWithRedis]int{
		CoreRedis:        0,
		JobServiceRedis:  1,
		RegistryRedis:    2, // nolint:gomnd
		ChartMuseumRedis: 3, // nolint:gomnd
		ClairRedis:       4, // nolint:gomnd
		TrivyRedis:       5, // nolint:gomnd
	}[r]
}

func (r ComponentWithRedis) String() string {
	return Component(r).String()
}
