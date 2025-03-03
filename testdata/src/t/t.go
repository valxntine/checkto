package t

import (
	"net/http"
	"time"
)

type noFailures struct {
	SvcTimeout  time.Duration
	Name        string
	IdleTimeout time.Duration
}

type timeoutWithInt struct {
	Name       string
	Port       string
	SvcTimeout int // want "timeout field SvcTimeout should use time.Duration instead of int"
}

type timeoutWithTime struct {
	Name       string
	SvcTimeout time.Time // want "timeout field SvcTimeout should use time.Duration"
	Port       string
}

type timeoutWithString struct {
	Name       string
	SvcTimeout string // want "timeout field SvcTimeout should use time.Duration instead of string"
	SomeField  bool
}

func serverWithOp() {
	t, _ := time.ParseDuration("500ms")
	_ = http.Server{WriteTimeout: t * time.Second} // want `assignment to WriteTimeout contains operation t \* time.Second but should use defined time.Duration`
}

func serverWithCfgOp() {
	t, _ := time.ParseDuration("500ms")
	cfg := noFailures{SvcTimeout: t}
	_ = http.Server{IdleTimeout: cfg.SvcTimeout * time.Second} // want `assignment to IdleTimeout contains operation cfg.SvcTimeout \* time.Second but should use defined time.Duration`
}

func serverWithDurOp() {
	t, _ := time.ParseDuration("500ms")
	_ = http.Server{IdleTimeout: t * time.Second} // want `assignment to IdleTimeout contains operation t \* time.Second but should use defined time.Duration`
}

func serverWithNoIssues() {
	t, _ := time.ParseDuration("500ms")
	cfg := noFailures{SvcTimeout: t}
	_ = http.Server{IdleTimeout: cfg.SvcTimeout}
}
