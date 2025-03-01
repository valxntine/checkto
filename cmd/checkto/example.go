package main

import (
	"log"
	"net/http"
	"time"
)

type Config struct {
	Enabled bool
	LogLevel string
	FirstTimeout time.Duration
	NumNodes int
	ThingTimeout time.Duration
}

type OtherConfig struct {
	Name string
	Timeout time.Duration
}

func Example() {
	t, err := time.ParseDuration("500ms")
	if err != nil {
		log.Fatalf("parsing duration: %v", err)
	}
	_ = http.Server{ReadTimeout: t * time.Second, IdleTimeout: t}
	cfg := Config{
		Enabled: true,
		LogLevel: "error",
		NumNodes: 4,
		ThingTimeout: t * time.Second,
	}
	_ = http.Server{WriteTimeout: cfg.ThingTimeout}
}
