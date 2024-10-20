package main

import (
	"KeepAccount/global/constant"
	"KeepAccount/global/nats/manager"
	_ "KeepAccount/service"
	"context"
	"fmt"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"os"
	"strings"
)

func main() {
	vectors := manager.EventManage.GetVector(context.TODO())
	nodes := make(map[string]graph.Node)
	nodeIDToName := make(map[int64]string)
	g := simple.NewDirectedGraph()
	for _, vector := range vectors {
		fmt.Println(vector)
		if _, exist := nodes[vector.From]; !exist {
			nodes[vector.From] = g.NewNode()
			nodeIDToName[nodes[vector.From].ID()] = vector.From
			g.AddNode(nodes[vector.From])
		}
		if _, exist := nodes[vector.To]; !exist {
			nodes[vector.To] = g.NewNode()
			nodeIDToName[nodes[vector.To].ID()] = vector.To
			g.AddNode(nodes[vector.To])
		}
		fromNode, toNode := nodes[vector.From], nodes[vector.To]
		g.SetEdge(simple.Edge{F: fromNode, T: toNode})
	}
	file, err := os.Create(constant.WORK_PATH + "/docs/graph.dot")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	fmt.Fprintln(file, "digraph G {")
	it := g.Edges()
	for it.Next() {
		e := it.Edge()

		fmt.Fprintf(
			file, "  %s -> %s;\n",
			strings.Replace(nodeIDToName[e.From().ID()], ".", "_", -1),
			strings.Replace(nodeIDToName[e.To().ID()], ".", "_", -1),
		)
	}
	fmt.Fprintln(file, "}")
}
