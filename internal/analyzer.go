package internal

// Analyzer takes an array of Ast and do validations in order to check if it matches the
// Nexema specification
type Analyzer struct {
	astArr []*Ast
}

// NewAnalyzer creates a new Analzer
func NewAnalyzer(input []*Ast) *Analyzer {
	return &Analyzer{
		astArr: input,
	}
}

// Analyze analyzes the given list of Ast when creating the Analyzer.
// It reports any error that is encountered or returns a NexemaDefinition
func (a *Analyzer) Analyze() (*NexemaDefinition, error) {
	return nil, nil
}
