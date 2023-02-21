package builder

// Builder is responsible of parsing, linking and analysing a Nexema project
type Builder struct {
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (self *Builder) Build() {}
