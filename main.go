package main

import (
	"benchmark_tool/core"
	"flag"
	"fmt"
	"os"
)

func main() {
	var configFile string
	var validate bool
	flag.StringVar(&configFile, "config", "", "Path to YAML configuration file")
	flag.BoolVar(&validate, "validate", false, "Validate read operations after benchmark")
	flag.Parse()

	var config *core.Config
	var err error
	if configFile != "" {
		config, err = core.LoadConfigFromYAML(configFile)
	} else {
		config = core.LoadConfigFromFlags()
	}
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}
	engine := &core.CassandraEngine{Config: config}
	if err := engine.Connect(); err != nil {
		panic(fmt.Sprintf("Failed to connect to Cassandra nodes: %v", err))
	}
	defer engine.Close()
	go core.RunOpenLoopBenchmark(engine, config)
	go core.RunClosedLoopBenchmark(engine, config)
	go core.RunLoadTest(engine, config)
	if validate {
		core.ValidateReads(engine, 10)
	}
	fmt.Println("Benchmark complete.")
}
