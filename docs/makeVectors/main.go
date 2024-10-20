package main

import (
	"KeepAccount/global/constant"
	"KeepAccount/global/nats/manager"
	_ "KeepAccount/service"
	"context"
	"fmt"
	"os"
	"strings"
)

func main() {
	file, err := os.Create(constant.WORK_PATH + "/docs/eventVectorGraph.dot")
	if err != nil {
		panic(err)
	}
	defer func() {
		err = file.Close()
		if err != nil {
			panic(err)
		}
	}()
	_, err = fmt.Fprintln(file, "digraph G {")
	if err != nil {
		panic(err)
	}
	for _, vector := range manager.EventManage.GetVector(context.TODO()) {
		_, err = fmt.Fprintf(
			file, "  %s -> %s;\n",
			strings.Replace(vector.From, ".", "_", -1),
			strings.Replace(vector.To, ".", "_", -1),
		)
		if err != nil {
			panic(err)
		}
	}
	_, err = fmt.Fprintln(file, "}")
	if err != nil {
		panic(err)
	}
}
