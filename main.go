package main

import (
	"github.com/lmxia/nightwatcher/routers"
	"k8s.io/klog/v2"
)

func main() {
	routersInit := routers.InitRouter()
	if err := routersInit.Run(":8900"); err != nil {
		klog.Errorf("Failed to run, error: %v", err)
	}
}
