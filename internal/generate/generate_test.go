package generate

import (
	"fmt"
	"go/token"
	"strings"
	"testing"
	"time"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/stretchr/testify/assert"

	internal "github.com/autometrics-dev/autometrics-go/internal/autometrics"
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
)

const DefaultPrometheusInstanceUrl = "http://localhost:9090/"

func TestCommentDirective(t *testing.T) {
	sourceCode := `// This is the package comment.
package main

import (
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	prom "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"
)

// This comment is associated with the main function.
//
//autometrics:doc --slo "Service Test" --success-target 99
func main() {
	fmt.Println(hello) // line comment 3
}
`

	want := "// This is the package comment.\n" +
		"package main\n" +
		"\n" +
		"import (\n" +
		"\t\"github.com/autometrics-dev/autometrics-go/pkg/autometrics\"\n" +
		"\tprom \"github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus\"\n" +
		")\n" +
		"\n" +
		"// This comment is associated with the main function.\n" +
		"//\n" +
		"//\tautometrics:doc-start Generated documentation by Autometrics.\n" +
		"//\n" +
		"// # Autometrics\n" +
		"//\n" +
		"// # Prometheus\n" +
		"//\n" +
		"// View the live metrics for the `main` function:\n" +
		"//   - [Request Rate]\n" +
		"//   - [Error Ratio]\n" +
		"//   - [Latency (95th and 99th percentiles)]\n" +
		"//   - [Concurrent Calls]\n" +
		"//\n" +
		"// Or, dig into the metrics of *functions called by* `main`\n" +
		"//   - [Request Rate Callee]\n" +
		"//   - [Error Ratio Callee]\n" +
		"//\n" +
		"//\tautometrics:doc-end Generated documentation by Autometrics.\n" +
		"//\n" +
		"// [Request Rate]: http://localhost:9090/graph?g0.expr=%23+Rate+of+calls+to+the+%60main%60+function+per+second%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bfunction%3D%22main%22%7D%5B5m%5D%29%29&g0.tab=0\n" +
		"// [Error Ratio]: http://localhost:9090/graph?g0.expr=%23+Percentage+of+calls+to+the+%60main%60+function+that+return+errors%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bfunction%3D%22main%22%2Cresult%3D%22error%22%7D%5B5m%5D%29%29&g0.tab=0\n" +
		"// [Latency (95th and 99th percentiles)]: http://localhost:9090/graph?g0.expr=%23+95th+and+99th+percentile+latencies+%28in+seconds%29+for+the+%60main%60+function%0A%0Ahistogram_quantile%280.99%2C+sum+by+%28le%2C+function%2C+module%29+%28rate%28function_calls_duration_bucket%7Bfunction%3D%22main%22%7D%5B5m%5D%29%29%29+or+histogram_quantile%280.95%2C+sum+by+%28le%2C+function%2C+module%29+%28rate%28function_calls_duration_bucket%7Bfunction%3D%22main%22%7D%5B5m%5D%29%29%29&g0.tab=0\n" +
		"// [Concurrent Calls]: http://localhost:9090/graph?g0.expr=%23+Concurrent+calls+to+the+%60main%60+function%0A%0Asum+by+%28function%2C+module%29+function_calls_concurrent%7Bfunction%3D%22main%22%7D&g0.tab=0\n" +
		"// [Request Rate Callee]: http://localhost:9090/graph?g0.expr=%23+Rate+of+function+calls+emanating+from+%60main%60+function+per+second%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bcaller%3D%22main.main%22%7D%5B5m%5D%29%29&g0.tab=0\n" +
		"// [Error Ratio Callee]: http://localhost:9090/graph?g0.expr=%23+Percentage+of+function+emanating+from+%60main%60+function+that+return+errors%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bcaller%3D%22main.main%22%2Cresult%3D%22error%22%7D%5B5m%5D%29%29&g0.tab=0\n" +
		"//\n" +
		"//autometrics:doc --slo \"Service Test\" --success-target 99\n" +
		"func main() {\n" +
		"\tdefer prom.Instrument(prom.PreInstrument(prom.NewContext(\n" +
		"\t\tprom.WithConcurrentCalls(true),\n" +
		"\t\tprom.WithCallerName(true),\n" +
		"\t\tprom.WithSloName(\"Service Test\"),\n" +
		"\t\tprom.WithAlertSuccess(99),\n" +
		"\t)), nil) //autometrics:defer\n" +
		"\n" +
		"	fmt.Println(hello) // line comment 3\n" +
		"}\n"

	ctx, err := internal.NewGeneratorContext(autometrics.PROMETHEUS, DefaultPrometheusInstanceUrl, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	actual, err := GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	if err != nil {
		t.Fatalf("error generating the documentation: %s", err)
	}

	assert.Equal(t, want, actual, "The generated source code is not as expected.")
}

// TestCommentRefresh calls GenerateDocumentationAndInstrumentation on a
// decorated function that already has a comment, making sure that the autometrics
// directive only updates the comment section about autometrics.
func TestCommentRefresh(t *testing.T) {
	sourceCode := `// This is the package comment.
package main

import (
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	prom "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"
)

// This comment is associated with the main function.
//
//   autometrics:doc-start
//
// Obviously not a good comment
//
//   autometrics:doc-end DO NOT EDIT
//
//autometrics:doc --slo "API" --latency-target 99.9 --latency-ms 500
func main() {
	fmt.Println(hello) // line comment 3
}
`

	want := "// This is the package comment.\n" +
		"package main\n" +
		"\n" +
		"import (\n" +
		"\t\"github.com/autometrics-dev/autometrics-go/pkg/autometrics\"\n" +
		"\tprom \"github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus\"\n" +
		")\n" +
		"\n" +
		"// This comment is associated with the main function.\n" +
		"//\n" +
		"//\tautometrics:doc-start Generated documentation by Autometrics.\n" +
		"//\n" +
		"// # Autometrics\n" +
		"//\n" +
		"// # Prometheus\n" +
		"//\n" +
		"// View the live metrics for the `main` function:\n" +
		"//   - [Request Rate]\n" +
		"//   - [Error Ratio]\n" +
		"//   - [Latency (95th and 99th percentiles)]\n" +
		"//   - [Concurrent Calls]\n" +
		"//\n" +
		"// Or, dig into the metrics of *functions called by* `main`\n" +
		"//   - [Request Rate Callee]\n" +
		"//   - [Error Ratio Callee]\n" +
		"//\n" +
		"//\tautometrics:doc-end Generated documentation by Autometrics.\n" +
		"//\n" +
		"// [Request Rate]: http://localhost:9090/graph?g0.expr=%23+Rate+of+calls+to+the+%60main%60+function+per+second%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bfunction%3D%22main%22%7D%5B5m%5D%29%29&g0.tab=0\n" +
		"// [Error Ratio]: http://localhost:9090/graph?g0.expr=%23+Percentage+of+calls+to+the+%60main%60+function+that+return+errors%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bfunction%3D%22main%22%2Cresult%3D%22error%22%7D%5B5m%5D%29%29&g0.tab=0\n" +
		"// [Latency (95th and 99th percentiles)]: http://localhost:9090/graph?g0.expr=%23+95th+and+99th+percentile+latencies+%28in+seconds%29+for+the+%60main%60+function%0A%0Ahistogram_quantile%280.99%2C+sum+by+%28le%2C+function%2C+module%29+%28rate%28function_calls_duration_bucket%7Bfunction%3D%22main%22%7D%5B5m%5D%29%29%29+or+histogram_quantile%280.95%2C+sum+by+%28le%2C+function%2C+module%29+%28rate%28function_calls_duration_bucket%7Bfunction%3D%22main%22%7D%5B5m%5D%29%29%29&g0.tab=0\n" +
		"// [Concurrent Calls]: http://localhost:9090/graph?g0.expr=%23+Concurrent+calls+to+the+%60main%60+function%0A%0Asum+by+%28function%2C+module%29+function_calls_concurrent%7Bfunction%3D%22main%22%7D&g0.tab=0\n" +
		"// [Request Rate Callee]: http://localhost:9090/graph?g0.expr=%23+Rate+of+function+calls+emanating+from+%60main%60+function+per+second%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bcaller%3D%22main.main%22%7D%5B5m%5D%29%29&g0.tab=0\n" +
		"// [Error Ratio Callee]: http://localhost:9090/graph?g0.expr=%23+Percentage+of+function+emanating+from+%60main%60+function+that+return+errors%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bcaller%3D%22main.main%22%2Cresult%3D%22error%22%7D%5B5m%5D%29%29&g0.tab=0\n" +
		"//\n" +
		"//autometrics:doc --slo \"API\" --latency-target 99.9 --latency-ms 500\n" +
		"func main() {\n" +
		"\tdefer prom.Instrument(prom.PreInstrument(prom.NewContext(\n" +
		"\t\tprom.WithConcurrentCalls(true),\n" +
		"\t\tprom.WithCallerName(true),\n" +
		"\t\tprom.WithSloName(\"API\"),\n" +
		"\t\tprom.WithAlertLatency(500000000*time.Nanosecond, 99.9),\n" +
		"\t)), nil) //autometrics:defer\n" +
		"\n" +
		"	fmt.Println(hello) // line comment 3\n" +
		"}\n"

	ctx, err := internal.NewGeneratorContext(autometrics.PROMETHEUS, DefaultPrometheusInstanceUrl, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	actual, err := GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	if err != nil {
		t.Fatalf("error generating the documentation: %s", err)
	}

	assert.Equal(t, want, actual, "The generated source code is not as expected.")
}

func TestCommentDirectiveErrors(t *testing.T) {
	sourceCode := `// This is the package comment.
package main

import (
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	prom "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"
)

// This comment is associated with the main function.
//
//autometrics:doc --slo "Service Test" --success-target 12394
func main() {
	fmt.Println(hello) // line comment 3
}
`
	ctx, err := internal.NewGeneratorContext(autometrics.PROMETHEUS, DefaultPrometheusInstanceUrl, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	_, err = GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	assert.Error(t, err, "Calling generation must fail if the target success rate is unrealistic.")

	sourceCode = `// This is the package comment.
package main

import (
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	prom "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"
)

// This comment is associated with the main function.
//
//autometrics:doc --slo "Service Test" --success-target -49
func main() {
	fmt.Println(hello) // line comment 3
}
`
	ctx, err = internal.NewGeneratorContext(autometrics.PROMETHEUS, DefaultPrometheusInstanceUrl, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	_, err = GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	assert.Error(t, err, "Calling generation must fail if the target success rate is unrealistic.")

	sourceCode = `// This is the package comment.
package main

import (
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	prom "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"
)

// This comment is associated with the main function.
//
//autometrics:doc --success-target 90
func main() {
	fmt.Println(hello) // line comment 3
}
`
	ctx, err = internal.NewGeneratorContext(autometrics.PROMETHEUS, DefaultPrometheusInstanceUrl, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	_, err = GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	assert.Error(t, err, "Calling generation must fail if no service name is given.")

	sourceCode = `// This is the package comment.
package main

import (
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	prom "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"
)

// This comment is associated with the main function.
//
//autometrics:doc --latency-ms 5000 --latency-target 99.9
func main() {
	fmt.Println(hello) // line comment 3
}
`
	ctx, err = internal.NewGeneratorContext(autometrics.PROMETHEUS, DefaultPrometheusInstanceUrl, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	_, err = GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	assert.Error(t, err, "Calling generation must fail if no service name is given.")

	sourceCode = `// This is the package comment.
package main

import (
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	prom "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"
)

// This comment is associated with the main function.
//
//autometrics:doc --slo "API" --latency-target 90
func main() {
	fmt.Println(hello) // line comment 3
}
`
	ctx, err = internal.NewGeneratorContext(autometrics.PROMETHEUS, DefaultPrometheusInstanceUrl, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	_, err = GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	assert.Error(t, err, "Calling generation must fail if latency-target is given without latency-ms.")

	sourceCode = `// This is the package comment.
package main

import (
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	prom "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"
)

// This comment is associated with the main function.
//
//autometrics:doc --slo "API" --latency-ms 1000
func main() {
	fmt.Println(hello) // line comment 3
}
`
	ctx, err = internal.NewGeneratorContext(autometrics.PROMETHEUS, DefaultPrometheusInstanceUrl, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	_, err = GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	assert.Error(t, err, "Calling generation must fail if latency-target is given without latency-ms.")

	sourceCode = `// This is the package comment.
package main

import (
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	prom "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"
)

// This comment is associated with the main function.
//
//autometrics:doc --slo "API" --latency-ms -5000 --latency-target 99.9
func main() {
	fmt.Println(hello) // line comment 3
}
`
	ctx, err = internal.NewGeneratorContext(autometrics.PROMETHEUS, DefaultPrometheusInstanceUrl, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	_, err = GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	assert.Error(t, err, "Calling generation must fail if latency expectations are unrealistic.")

	sourceCode = `// This is the package comment.
package main

import (
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	prom "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"
)

// This comment is associated with the main function.
//
//autometrics:doc --slo "API" --latency-ms 5000 --latency-target 49999
func main() {
	fmt.Println(hello) // line comment 3
}
`
	ctx, err = internal.NewGeneratorContext(autometrics.PROMETHEUS, DefaultPrometheusInstanceUrl, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	_, err = GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	assert.Error(t, err, "Calling generation must fail if latency expectations are unrealistic.")

	sourceCode = `// This is the package comment.
package main

import (
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	prom "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"
)

// This comment is associated with the main function.
//
//autometrics:doc --slo "API" --latency-ms 5000 --latency-target -123
func main() {
	fmt.Println(hello) // line comment 3
}
`
	ctx, err = internal.NewGeneratorContext(autometrics.PROMETHEUS, DefaultPrometheusInstanceUrl, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	_, err = GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	assert.Error(t, err, "Calling generation must fail if latency expectations are unrealistic.")
}

func TestNamedReturnDetectionNothing(t *testing.T) {
	// package statement is mandatory for decorator.Parse call
	sourceCode := `
package main

func main() {
	fmt.Println(hello) // line comment 3
}
`
	want := ""

	sourceAst, err := decorator.Parse(sourceCode)
	if err != nil {
		t.Fatalf("error parsing the source code: %s", err)
	}

	funcNode, ok := sourceAst.Decls[0].(*dst.FuncDecl)
	if !ok {
		t.Fatalf("First node of source code is not a function declaration")
	}

	actual, err := errorReturnValueName(funcNode)
	if err != nil {
		t.Fatalf("error getting the returned value name: %s", err)
	}

	assert.Equal(t, want, actual, "The return value doesn't match what's expected")
}

func TestNamedReturnDetectionNoError(t *testing.T) {
	// package statement is mandatory for decorator.Parse call
	sourceCode := `
package main

func main() int {
	fmt.Println(hello) // line comment 3
        return 0
}
`
	want := ""

	sourceAst, err := decorator.Parse(sourceCode)
	if err != nil {
		t.Fatalf("error parsing the source code: %s", err)
	}

	funcNode, ok := sourceAst.Decls[0].(*dst.FuncDecl)
	if !ok {
		t.Fatalf("First node of source code is not a function declaration")
	}

	actual, err := errorReturnValueName(funcNode)
	if err != nil {
		t.Fatalf("error getting the returned value name: %s", err)
	}

	assert.Equal(t, want, actual, "The return value doesn't match what's expected")
}

func TestNamedReturnDetectionUnnamedError(t *testing.T) {
	// package statement is mandatory for decorator.Parse call
	sourceCode := `
package main

func main() error {
	fmt.Println(hello) // line comment 3
        return nil
}
`
	want := ""

	sourceAst, err := decorator.Parse(sourceCode)
	if err != nil {
		t.Fatalf("error parsing the source code: %s", err)
	}

	funcNode, ok := sourceAst.Decls[0].(*dst.FuncDecl)
	if !ok {
		t.Fatalf("First node of source code is not a function declaration")
	}

	actual, err := errorReturnValueName(funcNode)
	if err != nil {
		t.Fatalf("error getting the returned value name: %s", err)
	}

	assert.Equal(t, want, actual, "The return value doesn't match what's expected")
}

func TestNamedReturnDetectionUnnamedPairError(t *testing.T) {
	// package statement is mandatory for decorator.Parse call
	sourceCode := `
package main

func main() (int, error) {
	fmt.Println(hello) // line comment 3
        return 0, nil
}
`
	want := ""

	sourceAst, err := decorator.Parse(sourceCode)
	if err != nil {
		t.Fatalf("error parsing the source code: %s", err)
	}

	funcNode, ok := sourceAst.Decls[0].(*dst.FuncDecl)
	if !ok {
		t.Fatalf("First node of source code is not a function declaration")
	}

	actual, err := errorReturnValueName(funcNode)
	if err != nil {
		t.Fatalf("error getting the returned value name: %s", err)
	}

	assert.Equal(t, want, actual, "The return value doesn't match what's expected")
}

func TestNamedReturnDetectionUnnamedPairNoError(t *testing.T) {
	// package statement is mandatory for decorator.Parse call
	sourceCode := `
package main

func main() (int, int) {
	fmt.Println(hello) // line comment 3
        return 0, 1
}
`
	want := ""

	sourceAst, err := decorator.Parse(sourceCode)
	if err != nil {
		t.Fatalf("error parsing the source code: %s", err)
	}

	funcNode, ok := sourceAst.Decls[0].(*dst.FuncDecl)
	if !ok {
		t.Fatalf("First node of source code is not a function declaration")
	}

	actual, err := errorReturnValueName(funcNode)
	if err != nil {
		t.Fatalf("error getting the returned value name: %s", err)
	}

	assert.Equal(t, want, actual, "The return value doesn't match what's expected")
}

func TestNamedReturnDetectionNamedError(t *testing.T) {
	// package statement is mandatory for decorator.Parse call
	sourceCode := `
package main

func main() (cannotGetLuckyCollision error) {
	fmt.Println(hello) // line comment 3
        return nil
}
`
	want := "cannotGetLuckyCollision"

	sourceAst, err := decorator.Parse(sourceCode)
	if err != nil {
		t.Fatalf("error parsing the source code: %s", err)
	}

	funcNode, ok := sourceAst.Decls[0].(*dst.FuncDecl)
	if !ok {
		t.Fatalf("First node of source code is not a function declaration")
	}

	actual, err := errorReturnValueName(funcNode)
	if err != nil {
		t.Fatalf("error getting the returned value name: %s", err)
	}

	assert.Equal(t, want, actual, "The return value doesn't match what's expected")
}

func TestNamedReturnDetectionNamedErrorInPair(t *testing.T) {
	// package statement is mandatory for decorator.Parse call
	sourceCode := `
package main

func main() (i int, cannotGetLuckyCollision error) {
	fmt.Println(hello) // line comment 3
        return 0, nil
}
`
	want := "cannotGetLuckyCollision"

	sourceAst, err := decorator.Parse(sourceCode)
	if err != nil {
		t.Fatalf("error parsing the source code: %s", err)
	}

	funcNode, ok := sourceAst.Decls[0].(*dst.FuncDecl)
	if !ok {
		t.Fatalf("First node of source code is not a function declaration")
	}

	actual, err := errorReturnValueName(funcNode)
	if err != nil {
		t.Fatalf("error getting the returned value name: %s", err)
	}

	assert.Equal(t, want, actual, "The return value doesn't match what's expected")
}

func TestNamedReturnDetectionErrorsOnMultipleNamedErrors(t *testing.T) {
	// package statement is mandatory for decorator.Parse call
	sourceCode := `
package main

func main() (cannotGetLuckyCollision, otherError error) {
	fmt.Println(hello) // line comment 3
        return nil, nil
}
`
	sourceAst, err := decorator.Parse(sourceCode)
	if err != nil {
		t.Fatalf("error parsing the source code: %s", err)
	}

	funcNode, ok := sourceAst.Decls[0].(*dst.FuncDecl)
	if !ok {
		t.Fatalf("First node of source code is not a function declaration")
	}

	_, err = errorReturnValueName(funcNode)
	assert.Error(t, err, "Calling the named return detection must fail if there are multiple error values.")
}

func implementContextCodeGenTest(t *testing.T, contextToSerialize autometrics.Context, expected string) {
	sourceContext := internal.GeneratorContext{
		RuntimeCtx: contextToSerialize,
		FuncCtx: internal.GeneratorFunctionContext{
			CommentIndex: -1,
			ImplImportName:   "autometrics",
		},
	}

	node, err := buildAutometricsContextNode(sourceContext)
	if err != nil {
		t.Fatalf("error building the context node: %s", err)
	}

	want := fmt.Sprintf(`package main

var dummy2 = %v
`, expected)

	// We're obliged to reparse, modify, and build a complete Go file in order
	// to test that the context has been correctly built
	dummyCodeHeader := "package main"
	dummyFile, err := decorator.Parse(dummyCodeHeader)
	if err != nil {
		t.Fatalf("error parsing the dummy code header for test purposes: %s", err)
	}

	dummyFile.Decls = append(dummyFile.Decls, &dst.GenDecl{
		Tok:    token.VAR,
		Lparen: false,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names:  []*dst.Ident{dst.NewIdent("dummy2")},
				Type:   nil,
				Values: []dst.Expr{node},
				Decs:   dst.ValueSpecDecorations{},
			},
		},
		Rparen: false,
		Decs:   dst.GenDeclDecorations{},
	})

	var actualBuilder strings.Builder
	err = decorator.Fprint(&actualBuilder, dummyFile)
	if err != nil {
		t.Fatalf("error writing the dummy code testing file to the string: %s", err)
	}

	actual := actualBuilder.String()
	assert.Equal(t, want, actual, "Differences between the compile time context and the runtime constant generated.")
}

func TestNewContextCodeGen(t *testing.T) {
	implementContextCodeGenTest(t,
		autometrics.NewContext(),
		`autometrics.NewContext(
	autometrics.WithConcurrentCalls(true),
	autometrics.WithCallerName(true),
)`,
	)
}

func TestNoTrackContextCodeGen(t *testing.T) {
	ctx := autometrics.NewContext()
	ctx.TrackCallerName = false
	ctx.TrackConcurrentCalls = false
	implementContextCodeGenTest(t,
		ctx,
		`autometrics.NewContext(
	autometrics.WithConcurrentCalls(false),
	autometrics.WithCallerName(false),
)`,
	)
}

func TestLatencyContextCodeGen(t *testing.T) {
	ctx := autometrics.NewContext()
	ctx.TrackCallerName = false
	ctx.AlertConf = &autometrics.AlertConfiguration{
		ServiceName: "api",
		Latency: &autometrics.LatencySlo{
			Target:    243 * time.Microsecond,
			Objective: 0.99,
		},
		Success: nil,
	}
	implementContextCodeGenTest(t,
		ctx,
		`autometrics.NewContext(
	autometrics.WithConcurrentCalls(true),
	autometrics.WithCallerName(false),
	autometrics.WithSloName("api"),
	autometrics.WithAlertLatency(243000*time.Nanosecond, 0.99),
)`,
	)
}

func TestSuccessContextCodeGen(t *testing.T) {
	ctx := autometrics.NewContext()
	ctx.TrackCallerName = false
	ctx.AlertConf = &autometrics.AlertConfiguration{
		ServiceName: "api",
		Latency:     nil,
		Success: &autometrics.SuccessSlo{
			Objective: 0.99999,
		},
	}
	implementContextCodeGenTest(t,
		ctx,
		`autometrics.NewContext(
	autometrics.WithConcurrentCalls(true),
	autometrics.WithCallerName(false),
	autometrics.WithSloName("api"),
	autometrics.WithAlertSuccess(0.99999),
)`,
	)
}
