# Go Check Timeout

`checkto` is a linter for identifying when structs contain timeout fields that aren't `time.Duration`, and when a timeout is assigned, that's its done so with a value rather than an expression.

## Examples

The following struct would fail the linter because the timeout field isn't using `time.Duration`

```go
type SomeConfig struct {
	SomeHost string
	SomeTimeout int
	SomePort string
}
```

The following assignment would fail because it's using an operation, instead of the aforementioned time.Duration field of a config struct:

```go
func main() {
	t, _ := time.ParseDuration("500ms")
	cfg := SomeConfig{SomeTimeout: t}
	_ = http.Server{WriteTimeout: cfg.SomeTimeout * time.Second}
}
```

The following assignment would fail for the same reason as above:

```go
func main() {
	t, _ := time.ParseDuration("500ms")
	_ = http.Server{WriteTimeout: t * time.Second}
}
```
