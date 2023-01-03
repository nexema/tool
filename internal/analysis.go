package internal

// Analyzer takes an Ast built with a parser and run rules to check if their matches
// the mpack reference
type Analyzer struct {
	ast *Ast
}

func NewAnalyzer(ast *Ast) *Analyzer {
	return &Analyzer{ast: ast}
}

// Analyze analyzes the input ast and returns a FileDefinition
// if its a valid Ast or an error explaining what is wrong
func Analyze() (*FileDefinition, error) {
	return nil, nil
}
