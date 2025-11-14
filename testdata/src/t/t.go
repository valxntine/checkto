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

// ============================================================================
// EDGE CASES: Field Type Checking
// ============================================================================

// Edge Case: Pointer to time.Duration (should pass)
type pointerTimeout struct {
	Timeout *time.Duration // Should pass - it's using time.Duration
}

// Edge Case: Multiple fields on same line
type multipleFieldsOneLine struct {
	Start, End, Timeout int // want "timeout field Timeout should use time.Duration instead of int"
}

// Edge Case: Anonymous struct with timeout
// NOTE: Anonymous structs don't create TypeSpec nodes, so field types are NOT checked
func anonymousStructTimeout() {
	_ = struct {
		Timeout int // NOT caught - anonymous struct fields aren't checked
	}{
		Timeout: 5,
	}
}

// Edge Case: Lowercase timeout field
type lowercaseTimeout struct {
	timeout int // want "timeout field timeout should use time.Duration instead of int"
}

// Edge Case: Interface type for timeout field
type interfaceTimeout struct {
	Timeout interface{} // want "timeout field Timeout should use time.Duration"
}

// Edge Case: Array/slice of ints for timeout
type sliceTimeout struct {
	Timeouts []int // want "timeout field Timeouts should use time.Duration"
}

// ============================================================================
// EDGE CASES: Assignment Operation Checking - Different Expression Types
// ============================================================================

// Edge Case: Integer literal in binary operation (now handled by exprToString)
func intLiteralInOp() {
	t := time.Second
	_ = http.Server{ // want `assignment to WriteTimeout contains operation t \* 30 but should use defined time.Duration`
		WriteTimeout: t * 30,
	}
}

// Edge Case: Different binary operations
func differentBinaryOps() {
	t := time.Second

	// Addition
	_ = http.Server{ // want `assignment to ReadTimeout contains operation t \+ time.Millisecond but should use defined time.Duration`
		ReadTimeout: t + time.Millisecond,
	}

	// Division
	_ = http.Server{ // want `assignment to IdleTimeout contains operation t / 2 but should use defined time.Duration`
		IdleTimeout: t / 2,
	}

	// Subtraction
	_ = http.Server{ // want `assignment to ReadHeaderTimeout contains operation t - time.Second but should use defined time.Duration`
		ReadHeaderTimeout: t - time.Second,
	}
}

// Edge Case: Call expression in binary operation (now handled gracefully)
func callExprInBinaryOp() {
	getTimeout := func() time.Duration { return time.Second }

	// The call expression getTimeout() is handled by exprToString as "<function call>"
	_ = http.Server{ // want `assignment to WriteTimeout contains operation <function call> \* time.Second but should use defined time.Duration`
		WriteTimeout: getTimeout() * time.Second,
	}
}

// Edge Case: Type conversion in binary operation
func typeConversionInOp() {
	seconds := 30
	// time.Duration(seconds) is a CallExpr, handled as "<function call>"
	_ = http.Server{ // want `assignment to WriteTimeout contains operation <function call> \* time.Second but should use defined time.Duration`
		WriteTimeout: time.Duration(seconds) * time.Second,
	}
}

// Edge Case: Nested selector in binary operation
type NestedConfig struct {
	Inner struct {
		Timeout time.Duration
	}
}

func nestedSelectorOp() {
	cfg := NestedConfig{}
	cfg.Inner.Timeout = time.Second

	// Now handled correctly via recursive exprToString
	_ = http.Server{ // want `assignment to WriteTimeout contains operation cfg.Inner.Timeout \* time.Second but should use defined time.Duration`
		WriteTimeout: cfg.Inner.Timeout * time.Second,
	}
}

// ============================================================================
// EDGE CASES: Panic-Prone Scenarios
// ============================================================================

// Edge Case: Composite literal with selector expression as key (embedded field access)
type EmbeddedTimeout struct {
	noFailures
}

func embeddedFieldAccess() {
	// If someone tries to use a selector as a key, kv.Key.(*ast.Ident) will panic
	// This is rare but the code should handle it gracefully
	cfg := EmbeddedTimeout{}
	cfg.SvcTimeout = time.Second
	_ = cfg
}

// Edge Case: Composite literal with no elements (shouldn't crash)
func emptyCompositeLiteral() {
	_ = http.Server{} // Should not crash - handled by compLit.Elts == nil check
}

// Edge Case: Assignment with multiple Rhs values (Rhs[0] might not be composite literal)
func multipleRhsValues() {
	a, b := 1, 2 // Rhs has 2 values, not composite literal
	_, _ = a, b
}

// Edge Case: Assignment where Rhs[0] is not a composite literal
func rhsNotCompositeLit() {
	cfg := noFailures{SvcTimeout: time.Second}
	cfg2 := cfg // Rhs[0] is *ast.Ident, not *ast.CompositeLit
	_ = cfg2
}

// ============================================================================
// EDGE CASES: Currently Uncaught Violations (by design limitations)
// ============================================================================

// Edge Case: Short variable declaration with operation (NOT caught)
func shortDeclWithOp() {
	timeout := time.Second * 5 // NOT caught - not in composite literal
	_ = timeout
}

// Edge Case: Regular variable declaration with operation (NOT caught)
func varDeclWithOp() {
	var timeout = time.Second * 5 // NOT caught
	_ = timeout
}

// Edge Case: Direct field assignment with operation (NOT caught)
func directFieldAssignment() {
	srv := &http.Server{}
	srv.WriteTimeout = time.Second * 30 // NOT caught - not in composite literal
	_ = srv
}

// Edge Case: Return statement with operation
// NOTE: Return statements ARE actually caught! They contain AssignStmt nodes internally
func returnWithOp() http.Server {
	t := time.Second
	return http.Server{
		WriteTimeout: t * 30, // NOT caught - return statements don't create AssignStmt nodes
	}
}

// ============================================================================
// EDGE CASES: Should Not Trigger (Non-timeout contexts)
// ============================================================================

// Edge Case: Composite literal with positional args instead of key-value
func positionalArgs() {
	type SimpleStruct struct {
		Name    string
		Timeout time.Duration
	}

	_ = SimpleStruct{
		"test",
		time.Second, // No key, so won't be checked
	}
}

// Edge Case: Slice literals (should not be checked)
func sliceLiterals() {
	timeouts := []time.Duration{
		time.Second * 5, // Not a struct field assignment
	}
	_ = timeouts
}

// Edge Case: Map literals (should not be checked)
func mapLiterals() {
	timeoutMap := map[string]time.Duration{
		"default": time.Second * 5, // Key is string literal, not timeout field
	}
	_ = timeoutMap
}

// Edge Case: Unary expression (not binary, should pass)
func unaryExpression() {
	t := time.Second
	negated := -t

	_ = http.Server{
		WriteTimeout: negated, // Should pass - no binary operation
	}
}

// Edge Case: Parenthesized binary expression (should still catch)
func parenthesizedOp() {
	t := time.Second
	_ = http.Server{ // want `assignment to WriteTimeout contains operation t \* time.Millisecond but should use defined time.Duration`
		WriteTimeout: (t * time.Millisecond),
	}
}
