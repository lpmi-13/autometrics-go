package generate

import (
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/stretchr/testify/assert"

	"github.com/autometrics-dev/autometrics-go/internal/doc"
)

// TestCommentDirective calls GenerateDocumentationAndInstrumentation on a
// decorated function, making sure that the autometrics
// directive adds a new comment section about autometrics.
func TestCommentDirective(t *testing.T) {
	sourceCode := `// This is the package comment.
package main

// This comment is associated with the main function.
//
//autometrics:doc
func main() {
	fmt.Println(hello) // line comment 3
}
`

	want := "// This is the package comment.\n" +
		"package main\n" +
		"\n" +
		"// This comment is associated with the main function.\n" +
		"//\n" +
		"//\n" +
		"//   autometrics:doc-start DO NOT EDIT HERE AND LINE ABOVE\n" +
		"//\n" +
		"// # Autometrics\n" +
		"//\n" +
		"// ## Prometheus\n" +
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
		"// [Request Rate]: http://localhost:9090/graph?g0.expr=%23+Rate+of+calls+to+the+%60main%60+function+per+second%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bfunction%3D%22main%22%7D%5B5m%5D%29%29&g0.tab=0\n" +
		"// [Error Ratio]: http://localhost:9090/graph?g0.expr=%23+Percentage+of+calls+to+the+%60main%60+function+that+return+errors%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bfunction%3D%22main%22%2Cresult%3D%22error%22%7D%5B5m%5D%29%29&g0.tab=0\n" +
		"// [Latency (95th and 99th percentiles)]: http://localhost:9090/graph?g0.expr=%23+95th+and+99th+percentile+latencies+%28in+seconds%29+for+the+%60main%60+function%0A%0Ahistogram_quantile%280.99%2C+sum+by+%28le%2C+function%2C+module%29+%28rate%28function_calls_duration_bucket%7Bfunction%3D%22main%22%7D%5B5m%5D%29%29%29+or+histogram_quantile%280.95%2C+sum+by+%28le%2C+function%2C+module%29+%28rate%28function_calls_duration_bucket%7Bfunction%3D%22main%22%7D%5B5m%5D%29%29%29&g0.tab=0\n" +
		"// [Concurrent Calls]: http://localhost:9090/graph?g0.expr=%23+Concurrent+calls+to+the+%60main%60+function%0A%0Asum+by+%28function%2C+module%29+function_calls_concurrent%7Bfunction%3D%22main%22%7D&g0.tab=0\n" +
		"// [Request Rate Callee]: http://localhost:9090/graph?g0.expr=%23+Rate+of+function+calls+emanating+from+%60main%60+function+per+second%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bcaller%3D%22main.main%22%7D%5B5m%5D%29%29&g0.tab=0\n" +
		"// [Error Ratio Callee]: http://localhost:9090/graph?g0.expr=%23+Percentage+of+function+emanating+from+%60main%60+function+that+return+errors%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bcaller%3D%22main.main%22%2Cresult%3D%22error%22%7D%5B5m%5D%29%29&g0.tab=0\n" +
		"//\n" +
		"//\n" +
		"//   autometrics:doc-end DO NOT EDIT HERE AND LINE BELOW\n" +
		"//\n" +
		"//autometrics:doc\n" +
		"func main() {\n" +
		"\tdefer autometrics.Instrument(autometrics.PreInstrument(), nil) //autometrics:defer\n" +
		"\n" +
		"	fmt.Println(hello) // line comment 3\n" +
		"}\n"

	actual, err := GenerateDocumentationAndInstrumentation(sourceCode, "main", doc.NewPrometheusDoc(doc.DefaultPrometheusInstanceUrl))
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

// This comment is associated with the main function.
//
//   autometrics:doc-start
//
// Obviously not a good comment
//
//   autometrics:doc-end DO NOT EDIT
//
//autometrics:doc
func main() {
	fmt.Println(hello) // line comment 3
}
`

	want := "// This is the package comment.\n" +
		"package main\n" +
		"\n" +
		"// This comment is associated with the main function.\n" +
		"//\n" +
		"//   autometrics:doc-start DO NOT EDIT HERE AND LINE ABOVE\n" +
		"//\n" +
		"// # Autometrics\n" +
		"//\n" +
		"// ## Prometheus\n" +
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
		"// [Request Rate]: http://localhost:9090/graph?g0.expr=%23+Rate+of+calls+to+the+%60main%60+function+per+second%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bfunction%3D%22main%22%7D%5B5m%5D%29%29&g0.tab=0\n" +
		"// [Error Ratio]: http://localhost:9090/graph?g0.expr=%23+Percentage+of+calls+to+the+%60main%60+function+that+return+errors%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bfunction%3D%22main%22%2Cresult%3D%22error%22%7D%5B5m%5D%29%29&g0.tab=0\n" +
		"// [Latency (95th and 99th percentiles)]: http://localhost:9090/graph?g0.expr=%23+95th+and+99th+percentile+latencies+%28in+seconds%29+for+the+%60main%60+function%0A%0Ahistogram_quantile%280.99%2C+sum+by+%28le%2C+function%2C+module%29+%28rate%28function_calls_duration_bucket%7Bfunction%3D%22main%22%7D%5B5m%5D%29%29%29+or+histogram_quantile%280.95%2C+sum+by+%28le%2C+function%2C+module%29+%28rate%28function_calls_duration_bucket%7Bfunction%3D%22main%22%7D%5B5m%5D%29%29%29&g0.tab=0\n" +
		"// [Concurrent Calls]: http://localhost:9090/graph?g0.expr=%23+Concurrent+calls+to+the+%60main%60+function%0A%0Asum+by+%28function%2C+module%29+function_calls_concurrent%7Bfunction%3D%22main%22%7D&g0.tab=0\n" +
		"// [Request Rate Callee]: http://localhost:9090/graph?g0.expr=%23+Rate+of+function+calls+emanating+from+%60main%60+function+per+second%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bcaller%3D%22main.main%22%7D%5B5m%5D%29%29&g0.tab=0\n" +
		"// [Error Ratio Callee]: http://localhost:9090/graph?g0.expr=%23+Percentage+of+function+emanating+from+%60main%60+function+that+return+errors%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bcaller%3D%22main.main%22%2Cresult%3D%22error%22%7D%5B5m%5D%29%29&g0.tab=0\n" +
		"//\n" +
		"//\n" +
		"//   autometrics:doc-end DO NOT EDIT HERE AND LINE BELOW\n" +
		"//\n" +
		"//autometrics:doc\n" +
		"func main() {\n" +
		"\tdefer autometrics.Instrument(autometrics.PreInstrument(), nil) //autometrics:defer\n" +
		"\n" +
		"	fmt.Println(hello) // line comment 3\n" +
		"}\n"

	actual, err := GenerateDocumentationAndInstrumentation(sourceCode, "main", doc.NewPrometheusDoc(doc.DefaultPrometheusInstanceUrl))
	if err != nil {
		t.Fatalf("error generating the documentation: %s", err)
	}

	assert.Equal(t, want, actual, "The generated source code is not as expected.")
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