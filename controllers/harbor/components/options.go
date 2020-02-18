package components

const PriorityBase = 100

type OptionGetter interface {
	GetPriority() *int32
}

type OptionSetter interface {
	SetPriority(*int32)
}

type Option struct {
	priority *int32
}

func (o *Option) SetPriority(priority *int32) {
	o.priority = priority
}

func (o *Option) GetPriority() *int32 {
	return o.priority
}
