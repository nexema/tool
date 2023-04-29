package rules

func GetTypeStatementRules(resolver ObjectResolver) *AnalyzerRuleCollection {
	group := &AnalyzerRuleCollection{resolver, make([]BaseAnalyzerRule, 5)}
	group.Rules[0] = ValidBaseType{group}
	group.Rules[1] = DuplicatedDefaultValue{group}
	group.Rules[2] = DefaultValueValidField{group}
	group.Rules[3] = DuplicatedFieldName{group}

	return group
}

type RuleType int8

const (
	TypeStatementRules RuleType = 1
)
