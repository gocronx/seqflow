package seqflow

import (
	"errors"
	"testing"
)

func TestDAG_LinearPipeline(t *testing.T) {
	nodes := []handlerNode{
		{name: "A", handler: handlerFunc(func(l, u int64) {})},
		{name: "B", handler: handlerFunc(func(l, u int64) {}), dependsOn: []string{"A"}},
		{name: "C", handler: handlerFunc(func(l, u int64) {}), dependsOn: []string{"B"}},
	}
	d, err := buildDAG(nodes)
	if err != nil {
		t.Fatalf("buildDAG 错误: %v", err)
	}
	if len(d.terminals) != 1 || d.terminals[0] != "C" {
		t.Errorf("终端节点 = %v, 期望 [C]", d.terminals)
	}
}

func TestDAG_Diamond(t *testing.T) {
	nodes := []handlerNode{
		{name: "A", handler: handlerFunc(func(l, u int64) {})},
		{name: "B", handler: handlerFunc(func(l, u int64) {}), dependsOn: []string{"A"}},
		{name: "C", handler: handlerFunc(func(l, u int64) {}), dependsOn: []string{"A"}},
		{name: "D", handler: handlerFunc(func(l, u int64) {}), dependsOn: []string{"B", "C"}},
	}
	d, err := buildDAG(nodes)
	if err != nil {
		t.Fatalf("buildDAG 错误: %v", err)
	}
	if len(d.terminals) != 1 || d.terminals[0] != "D" {
		t.Errorf("终端节点 = %v, 期望 [D]", d.terminals)
	}
}

func TestDAG_FanOut(t *testing.T) {
	nodes := []handlerNode{
		{name: "A", handler: handlerFunc(func(l, u int64) {})},
		{name: "B", handler: handlerFunc(func(l, u int64) {}), dependsOn: []string{"A"}},
		{name: "C", handler: handlerFunc(func(l, u int64) {}), dependsOn: []string{"A"}},
		{name: "D", handler: handlerFunc(func(l, u int64) {}), dependsOn: []string{"A"}},
	}
	d, err := buildDAG(nodes)
	if err != nil {
		t.Fatalf("buildDAG 错误: %v", err)
	}
	if len(d.terminals) != 3 {
		t.Errorf("终端节点数 = %d, 期望 3（B, C, D）", len(d.terminals))
	}
}

func TestDAG_DuplicateName(t *testing.T) {
	nodes := []handlerNode{
		{name: "A", handler: handlerFunc(func(l, u int64) {})},
		{name: "A", handler: handlerFunc(func(l, u int64) {})},
	}
	_, err := buildDAG(nodes)
	if !errors.Is(err, ErrDuplicateHandler) {
		t.Errorf("期望 ErrDuplicateHandler, 实际: %v", err)
	}
}

func TestDAG_UnknownDependency(t *testing.T) {
	nodes := []handlerNode{
		{name: "A", handler: handlerFunc(func(l, u int64) {}), dependsOn: []string{"X"}},
	}
	_, err := buildDAG(nodes)
	if !errors.Is(err, ErrUnknownDependency) {
		t.Errorf("期望 ErrUnknownDependency, 实际: %v", err)
	}
}

func TestDAG_CyclicDependency(t *testing.T) {
	nodes := []handlerNode{
		{name: "A", handler: handlerFunc(func(l, u int64) {}), dependsOn: []string{"B"}},
		{name: "B", handler: handlerFunc(func(l, u int64) {}), dependsOn: []string{"A"}},
	}
	_, err := buildDAG(nodes)
	if !errors.Is(err, ErrCyclicDependency) {
		t.Errorf("期望 ErrCyclicDependency, 实际: %v", err)
	}
}

func TestDAG_SelfDependency(t *testing.T) {
	nodes := []handlerNode{
		{name: "A", handler: handlerFunc(func(l, u int64) {}), dependsOn: []string{"A"}},
	}
	_, err := buildDAG(nodes)
	if !errors.Is(err, ErrCyclicDependency) {
		t.Errorf("期望 ErrCyclicDependency, 实际: %v", err)
	}
}
