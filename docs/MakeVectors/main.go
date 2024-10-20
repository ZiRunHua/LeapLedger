package main

import (
	"KeepAccount/global/nats/manager"
	"context"
	"fmt"
)

func main() {
	vector := manager.EventManage.GetVector(context.TODO())
	for _, s := range vector {
		fmt.Println(s)
	}

}
