package http_test

import (
	"fmt"
	"net/http"

	qhttp "gopkg.qsoa.cloud/service/http"
)

func ExampleRun() {
	qhttp.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "Hello world")
	}))

	qhttp.Run()

	// Output: test
}
