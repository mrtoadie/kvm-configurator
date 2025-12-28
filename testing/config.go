package main

import (
    "fmt"
    "log"

    "gopkg.in/ini.v1"
)

type ServerConfig struct {
    CPU       int    `ini:"cpu"`
    RAM       int `ini:"ram"`
    DISK      string   `ini:"disk"`
}

func main() {
    cfg, err := ini.Load("config.cfg")
    if err != nil {
        log.Fatalf("Failed to read config: %v", err)
    }

    var sCfg ServerConfig
    err = cfg.Section("debian").MapTo(&sCfg)
    if err != nil {
        log.Fatalf("Mapping failed: %v", err)
    }

    fmt.Printf("%+v\n", sCfg)
}