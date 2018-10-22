package main

import (
	"flag"
	"github.com/chenhg5/ecsDeploy/deploy"
)

func main() {

	var (
		op         string
		number     int
		configPath string
	)
	flag.StringVar(&op, "op", "deploy", "operation type")
	flag.IntVar(&number, "number", 1, "number")
	flag.StringVar(&configPath, "cfg", "./config.ini", "config path")
	flag.Parse()

	config, _ := deploy.GetConfig(configPath, "project")
	project := deploy.NewProject(config)

	// 部署新机器
	if op == "deploy" {
		project.DeployNew(number)
	}

	// 全量更新
	if op == "update" {
		project.UpdateAll()
	}
}
