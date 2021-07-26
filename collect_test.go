package main

import (
	"context"
	"testing"
)

func TestNewCollector(t *testing.T) {
	InitCfg()
	_ = NewCollector(context.Background(), CollectorOptMySQL(globalCfg.Collector.Mysql))
}
