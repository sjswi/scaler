package main

import (
	"conserver/pkg/config"
	"conserver/pkg/controller"
	"conserver/pkg/global"
	"conserver/pkg/k8s"
	"flag"
	"fmt"
	"time"
)

func main() {
	global.DbConfig = &config.DatabaseConfig{
		Strategy:          "table",
		ClusterConnConfig: make(map[string]*config.ClusterConfig),
		ScaleFactor:       0.1,
		ScaleRuleType:     "",
		Min:               0,
		Max:               0,
		TimeInterval:      0,
	}
	global.RedisConfig = make(map[string]map[string]string)

	global.ScalerStartTime = time.Now()
	// 解析命令行参数
	flag.Parse()
	client := k8s.GetK8sClient()

	go client.Listen()
	controller.RunController()
	fmt.Println("Asadsass")
}
