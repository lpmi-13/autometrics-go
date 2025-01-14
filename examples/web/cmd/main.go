package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	amImpl "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// This should be `//go:generate autometrics` in practice. Those are hacks to get the example working, see
// README
//go:generate go run ../../../cmd/autometrics/main.go

func main() {
	rand.Seed(time.Now().UnixNano())

	amImpl.Init(nil, amImpl.DefBuckets)

	http.HandleFunc("/", errorable(indexHandler))
	http.HandleFunc("/random-error", errorable(randomErrorHandler))
	http.Handle("/metrics", promhttp.Handler())

	log.Println("binding on http://localhost:62086")
	log.Fatal(http.ListenAndServe(":62086", nil))
}

// indexHandler handles the / route.
//
// It always succeeds and says hello.
//
//	autometrics:doc-start Generated documentation by Autometrics.
//
// # Autometrics
//
// # Prometheus
//
// View the live metrics for the `indexHandler` function:
//   - [Request Rate]
//   - [Error Ratio]
//   - [Latency (95th and 99th percentiles)]
//   - [Concurrent Calls]
//
// Or, dig into the metrics of *functions called by* `indexHandler`
//   - [Request Rate Callee]
//   - [Error Ratio Callee]
//
//	autometrics:doc-end Generated documentation by Autometrics.
//
// [Request Rate]: http://localhost:9090/graph?g0.expr=%23+Rate+of+calls+to+the+%60indexHandler%60+function+per+second%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bfunction%3D%22indexHandler%22%7D%5B5m%5D%29%29&g0.tab=0
// [Error Ratio]: http://localhost:9090/graph?g0.expr=%23+Percentage+of+calls+to+the+%60indexHandler%60+function+that+return+errors%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bfunction%3D%22indexHandler%22%2Cresult%3D%22error%22%7D%5B5m%5D%29%29&g0.tab=0
// [Latency (95th and 99th percentiles)]: http://localhost:9090/graph?g0.expr=%23+95th+and+99th+percentile+latencies+%28in+seconds%29+for+the+%60indexHandler%60+function%0A%0Ahistogram_quantile%280.99%2C+sum+by+%28le%2C+function%2C+module%29+%28rate%28function_calls_duration_bucket%7Bfunction%3D%22indexHandler%22%7D%5B5m%5D%29%29%29+or+histogram_quantile%280.95%2C+sum+by+%28le%2C+function%2C+module%29+%28rate%28function_calls_duration_bucket%7Bfunction%3D%22indexHandler%22%7D%5B5m%5D%29%29%29&g0.tab=0
// [Concurrent Calls]: http://localhost:9090/graph?g0.expr=%23+Concurrent+calls+to+the+%60indexHandler%60+function%0A%0Asum+by+%28function%2C+module%29+function_calls_concurrent%7Bfunction%3D%22indexHandler%22%7D&g0.tab=0
// [Request Rate Callee]: http://localhost:9090/graph?g0.expr=%23+Rate+of+function+calls+emanating+from+%60indexHandler%60+function+per+second%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bcaller%3D%22main.indexHandler%22%7D%5B5m%5D%29%29&g0.tab=0
// [Error Ratio Callee]: http://localhost:9090/graph?g0.expr=%23+Percentage+of+function+emanating+from+%60indexHandler%60+function+that+return+errors%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bcaller%3D%22main.indexHandler%22%2Cresult%3D%22error%22%7D%5B5m%5D%29%29&g0.tab=0
//
//autometrics:doc --slo "API" --latency-target 99 --latency-ms 250
func indexHandler(w http.ResponseWriter, _ *http.Request) error {
	defer amImpl.Instrument(amImpl.PreInstrument(amImpl.NewContext(
		amImpl.WithConcurrentCalls(true),
		amImpl.WithCallerName(true),
		amImpl.WithSloName("API"),
		amImpl.WithAlertLatency(250000000*time.Nanosecond, 99),
	)), nil) //autometrics:defer

	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

	_, err := fmt.Fprintf(w, "Hello, World!\n")
	return err
}

var handlerError = errors.New("failed to handle request")

// randomErrorHandler handles the /random-error route.
//
// It returns an error around 50% of the time.
//
//	autometrics:doc-start Generated documentation by Autometrics.
//
// # Autometrics
//
// # Prometheus
//
// View the live metrics for the `randomErrorHandler` function:
//   - [Request Rate]
//   - [Error Ratio]
//   - [Latency (95th and 99th percentiles)]
//   - [Concurrent Calls]
//
// Or, dig into the metrics of *functions called by* `randomErrorHandler`
//   - [Request Rate Callee]
//   - [Error Ratio Callee]
//
//	autometrics:doc-end Generated documentation by Autometrics.
//
// [Request Rate]: http://localhost:9090/graph?g0.expr=%23+Rate+of+calls+to+the+%60randomErrorHandler%60+function+per+second%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bfunction%3D%22randomErrorHandler%22%7D%5B5m%5D%29%29&g0.tab=0
// [Error Ratio]: http://localhost:9090/graph?g0.expr=%23+Percentage+of+calls+to+the+%60randomErrorHandler%60+function+that+return+errors%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bfunction%3D%22randomErrorHandler%22%2Cresult%3D%22error%22%7D%5B5m%5D%29%29&g0.tab=0
// [Latency (95th and 99th percentiles)]: http://localhost:9090/graph?g0.expr=%23+95th+and+99th+percentile+latencies+%28in+seconds%29+for+the+%60randomErrorHandler%60+function%0A%0Ahistogram_quantile%280.99%2C+sum+by+%28le%2C+function%2C+module%29+%28rate%28function_calls_duration_bucket%7Bfunction%3D%22randomErrorHandler%22%7D%5B5m%5D%29%29%29+or+histogram_quantile%280.95%2C+sum+by+%28le%2C+function%2C+module%29+%28rate%28function_calls_duration_bucket%7Bfunction%3D%22randomErrorHandler%22%7D%5B5m%5D%29%29%29&g0.tab=0
// [Concurrent Calls]: http://localhost:9090/graph?g0.expr=%23+Concurrent+calls+to+the+%60randomErrorHandler%60+function%0A%0Asum+by+%28function%2C+module%29+function_calls_concurrent%7Bfunction%3D%22randomErrorHandler%22%7D&g0.tab=0
// [Request Rate Callee]: http://localhost:9090/graph?g0.expr=%23+Rate+of+function+calls+emanating+from+%60randomErrorHandler%60+function+per+second%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bcaller%3D%22main.randomErrorHandler%22%7D%5B5m%5D%29%29&g0.tab=0
// [Error Ratio Callee]: http://localhost:9090/graph?g0.expr=%23+Percentage+of+function+emanating+from+%60randomErrorHandler%60+function+that+return+errors%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_count%7Bcaller%3D%22main.randomErrorHandler%22%2Cresult%3D%22error%22%7D%5B5m%5D%29%29&g0.tab=0
//
//autometrics:doc --slo "API" --success-target 90
func randomErrorHandler(w http.ResponseWriter, _ *http.Request) (err error) {
	defer amImpl.Instrument(amImpl.PreInstrument(amImpl.NewContext(
		amImpl.WithConcurrentCalls(true),
		amImpl.WithCallerName(true),
		amImpl.WithSloName("API"),
		amImpl.WithAlertSuccess(90),
	)), &err) //autometrics:defer

	isErr := rand.Intn(2) == 0

	if isErr {
		err = handlerError
	} else {
		w.WriteHeader(http.StatusOK)
	}

	return
}

// errorable is a wrapper to allow using functions that return `error` in route handlers.
func errorable(handler func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
