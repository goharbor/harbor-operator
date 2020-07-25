package v1alpha2

type ComponentWithDatabase Component

const (
	CoreDatabase         = ComponentWithDatabase(CoreComponent)
	NotaryServerDatabase = ComponentWithDatabase(NotaryServerComponent)
	NotarySignerDatabase = ComponentWithDatabase(NotarySignerComponent)
	ClairDatabase        = ComponentWithDatabase(ClairComponent)
)

func (r ComponentWithDatabase) DBName() string {
	return r.String()
}

func (r ComponentWithDatabase) String() string {
	return Component(r).String()
}
