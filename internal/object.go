package internal

// Object represents an evaluated TypeStmt
type Object struct {
	ID   string    // The object's id
	Stmt *TypeStmt // The TypeStmt used to evaluate the current object
}

type ObjectCollection map[string]*Object

func NewObjectCollection() *ObjectCollection {
	m := make(ObjectCollection)
	return &m
}

// append adds a new Object. It will return false if the object is already added
func (o *ObjectCollection) append(name string, obj *Object) bool {
	_, ok := (*o)[name]
	if ok {
		return false
	}

	(*o)[name] = obj
	return true
}

func (o *ObjectCollection) find(name string) *Object {
	return (*o)[name]
}

func (o *ObjectCollection) exists(name string) bool {
	_, ok := (*o)[name]
	return ok
}
