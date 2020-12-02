package lcm

const (
	DatabasePropertyName string = "database"
	StoragePropertyName  string = "storage"
	CachePropertyName    string = "cache"
)

// Property is the current property of component.
type Property struct {
	// Name, e.p: Connection,Port.
	Name string
	// Value, e.p: "rfs-harborcluster-sample.svc"
	Value interface{}
}

type Properties []*Property

// Add append a new property to properties.
func (ps *Properties) Add(name string, value interface{}) {
	p := &Property{
		Name:  name,
		Value: value,
	}
	*ps = append(*ps, p)
}

// Update updates properties according to the given arguments.
func (ps *Properties) Update(name string, value interface{}) {
	for _, p := range *ps {
		if p.Name == name {
			p.Value = value

			return
		}
	}
}

// Get retrieves properties according to the given name.
func (ps *Properties) Get(name string) *Property {
	for _, p := range *ps {
		if p.Name == name {
			return p
		}
	}

	return nil
}

// ToInt parse properties value to int type.
func (p *Property) ToInt() int {
	if p.Value != nil {
		if v, ok := p.Value.(int); ok {
			return v
		}
	}

	return 0
}

// ToString parse properties value to string type.
func (p *Property) ToString() string {
	if p.Value != nil {
		if v, ok := p.Value.(string); ok {
			return v
		}
	}

	return ""
}

// ToFloat64 parse properties value to float64 type.
func (p *Property) ToFloat64() float64 {
	if p.Value != nil {
		if v, ok := p.Value.(float64); ok {
			return v
		}
	}

	return 0
}
