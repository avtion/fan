package main

import (
	"github.com/google/gops/agent"
	"go.uber.org/zap"
)

func init() {
	if err := agent.Listen(agent.Options{}); err != nil {
		log.Error("gops agent run failed", zap.Error(err))
		return
	}
}

func main() {
	InitCfg()
	InitRobots(globalCfg)
}
