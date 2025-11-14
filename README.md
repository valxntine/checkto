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

## golangci-lint Integration

`checkto` can be integrated with golangci-lint as a custom linter plugin.

### Setup

1. Build the plugin:

```bash
cd plugin
go build -buildmode=plugin -o plugin.so plugin.go
```

2. Add the plugin configuration to your `.golangci.yml`:

#### golangci-lint v2 (current)

```yaml
version: "2"

linters:
  enable:
    - checkto
  settings:
    custom:
      checkto:
        path: ./plugin/plugin.so
        description: Checks timeout fields use time.Duration and assignments don't use operations
        original-url: github.com/valxntine/checkto
```

#### golangci-lint v1 (legacy)

```yaml
linters-settings:
  custom:
    checkto:
      path: ./plugin/plugin.so
      description: Checks timeout fields use time.Duration and assignments don't use operations
      original-url: github.com/valxntine/checkto

linters:
  enable:
    - checkto
```

3. Run golangci-lint:

```bash
golangci-lint run
```

### Standalone Usage

You can also use `checkto` as a standalone analyzer without golangci-lint by importing it directly in your analysis tool:

```go
import "github.com/valxntine/checkto"

// Use checkto.DurationAnalyzer in your analysis driver
```
