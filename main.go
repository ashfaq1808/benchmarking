package main

import (
    "benchmark_tool/core"
    "flag"
    "fmt"
    "os"
)

func main() {
    var configFile string
    flag.StringVar(&configFile, "config", "", "Path to YAML configuration file")
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

    core.RunBenchmark(engine, config)
}
