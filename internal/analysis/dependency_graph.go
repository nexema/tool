package analysis

type dependencyGraph map[string][]string

func (g *dependencyGraph) addDependency(node, dependency string) {
	(*g)[node] = append((*g)[node], dependency)
}

func (g *dependencyGraph) findCyclicDependencies() [][]string {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	cyclicDependencies := [][]string{}

	var findCycle func(node string, stack []string)
	findCycle = func(node string, stack []string) {
		visited[node] = true
		recStack[node] = true
		stack = append(stack, node)

		for _, neighbor := range (*g)[node] {
			if !visited[neighbor] {
				findCycle(neighbor, stack)
			} else if recStack[neighbor] {
				cycle := append(stack, neighbor)
				cyclicDependencies = append(cyclicDependencies, cycle)
			}
		}

		recStack[node] = false
	}

	for node := range *g {
		if !visited[node] {
			findCycle(node, nil)
		}
	}

	return cyclicDependencies
}
