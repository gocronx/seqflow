package seqflow

import "fmt"

// handlerNode 表示一个命名的 handler 及其可选依赖
type handlerNode struct {
	name      string   // handler 名称
	handler   Handler  // 用户回调
	dependsOn []string // 依赖的 handler 名称列表
}

// dag 保存经过验证的拓扑结构
type dag struct {
	nodes     []handlerNode // 所有节点
	order     []string      // 拓扑排序顺序
	terminals []string      // 终端 handler（无下游依赖者）
}

// buildDAG 验证 handler 图并计算拓扑排序
func buildDAG(nodes []handlerNode) (*dag, error) {
	// 检查名称唯一性
	nameSet := make(map[string]struct{}, len(nodes))
	for _, n := range nodes {
		if _, exists := nameSet[n.name]; exists {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateHandler, n.name)
		}
		nameSet[n.name] = struct{}{}
	}

	// 验证所有依赖存在
	for _, n := range nodes {
		for _, dep := range n.dependsOn {
			if _, exists := nameSet[dep]; !exists {
				return nil, fmt.Errorf("%w: %s 依赖 %s", ErrUnknownDependency, n.name, dep)
			}
		}
	}

	// 拓扑排序（Kahn 算法）
	inDegree := make(map[string]int, len(nodes))
	children := make(map[string][]string, len(nodes))
	for _, n := range nodes {
		if _, ok := inDegree[n.name]; !ok {
			inDegree[n.name] = 0
		}
		for _, dep := range n.dependsOn {
			children[dep] = append(children[dep], n.name)
			inDegree[n.name]++
		}
	}

	var queue []string
	for _, n := range nodes {
		if inDegree[n.name] == 0 {
			queue = append(queue, n.name)
		}
	}

	var order []string
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		order = append(order, name)

		for _, child := range children[name] {
			inDegree[child]--
			if inDegree[child] == 0 {
				queue = append(queue, child)
			}
		}
	}

	if len(order) != len(nodes) {
		return nil, fmt.Errorf("%w", ErrCyclicDependency)
	}

	// 找到终端 handler（不被任何 DependsOn 引用的 handler）
	referenced := make(map[string]struct{})
	for _, n := range nodes {
		for _, dep := range n.dependsOn {
			referenced[dep] = struct{}{}
		}
	}
	var terminals []string
	for _, name := range order {
		if _, isReferenced := referenced[name]; !isReferenced {
			terminals = append(terminals, name)
		}
	}

	return &dag{
		nodes:     nodes,
		order:     order,
		terminals: terminals,
	}, nil
}
