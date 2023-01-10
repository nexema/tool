package internal

type ErrorCollection []error

func NewErrorCollection() *ErrorCollection {
	return &ErrorCollection{}
}

func (e *ErrorCollection) Report(err error) {
	(*e) = append((*e), err)
}
