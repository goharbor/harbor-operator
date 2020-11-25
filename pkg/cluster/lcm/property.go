package lcm

const (
	DatabasePropertyName string = "database"
	StoragePropertyName  string = "storage"
	CachePropertyName    string = "cache"
)

//Property is the current property of component.
type Property struct {
	//Property name, e.p: Connection,Port.
	Name string
	//Property value, e.p: "rfs-harborcluster-sample.svc"
	Value interface{}
}

type Properties []*Property

//Add append a new property to properties
func (ps *Properties) Add(Name string, Value interface{}) {
	p := &Property{
		Name:  Name,
		Value: Value,
	}
	*ps = append(*ps, p)
}

//Update updates properties according to the given arguments
func (ps *Properties) Update(Name string, Value interface{}) {
	for _, p := range *ps {
		if p.Name == Name {
			p.Value = Value
			return
		}
	}
}

//Get retrieves properties according to the given name
func (ps *Properties) Get(Name string) *Property {
	for _, p := range *ps {
		if p.Name == Name {
			return p
		}
	}
	return nil
}

//ToInt parse properties value to int type
func (p *Property) ToInt() int {
	if p.Value != nil {
		if v, ok := p.Value.(int); ok {
			return v
		}
	}

	return 0
}

//ToString parse properties value to string type
func (p *Property) ToString() string {
	if p.Value != nil {
		if v, ok := p.Value.(string); ok {
			return v
		}
	}

	return ""
}

//ToFloat64 parse properties value to float64 type
func (p *Property) ToFloat64() float64 {
	if p.Value != nil {
		if v, ok := p.Value.(float64); ok {
			return v
		}
	}

	return 0
}
