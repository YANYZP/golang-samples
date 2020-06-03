// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// [START trace_setup_go_quickstart]

// Sample trace_quickstart traces incoming and outgoing requests.
package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

func readFromFile(fileName string) map[string][]string {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var vendorOfItemMap = make(map[string][]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		itemVendorInfo := scanner.Text()
		words := strings.Fields(itemVendorInfo)
		fmt.Println(itemVendorInfo)
		for i := 1; i < len(words); i++ {
			fmt.Println("vendor", words[i])
			vendorOfItemMap[words[0]] = append(vendorOfItemMap[words[0]], words[i])
		}
		fmt.Println("vendor list", vendorOfItemMap[words[0]])
	}

	// for k, v := range vendorOfItemMap {
	// 	for _, vv := range v {
	// 		fmt.Println(k, vv)
	// 	}
	// }

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return vendorOfItemMap
}

func generateURL(ingredientName string) string {
	vendorOfItemMap := readFromFile("vendor.txt")

	vendorNameList, ok := vendorOfItemMap[ingredientName]

	if !ok {
		return "Service B: No vendor info about " + ingredientName + "\n"
	}

	URLStrBuilder := strings.Builder{}
	URLStrBuilder.WriteString(ingredientName)
	URLStrBuilder.WriteString("/")

	for _, vendorName := range vendorNameList {
		URLStrBuilder.WriteString(vendorName)
		URLStrBuilder.WriteString("/")
	}

	URLStr := URLStrBuilder.String()
	URLStr = URLStr[:len(URLStr)-1]

	fmt.Println("vendorNameList", vendorNameList)

	fmt.Println("sending url", URLStr, "to service c")

	return URLStr
}

func main() {
	// Create and register a OpenCensus Stackdriver Trace exporter.
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: os.Getenv("PROJECT_ID"),
	})
	if err != nil {
		log.Fatal(err)
	}
	trace.RegisterExporter(exporter)

	// By default, traces will be sampled relatively rarely. To change the
	// sampling frequency for your entire program, call ApplyConfig. Use a
	// ProbabilitySampler to sample a subset of traces, or use AlwaysSample to
	// collect a trace on every run.
	//
	// Be careful about using trace.AlwaysSample in a production application
	// with significant traffic: a new trace will be started and exported for
	// every request.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	client := &http.Client{
		Transport: &ochttp.Transport{
			// Use Google Cloud propagation format.
			Propagation: &propagation.HTTPFormat{},
		},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ingredientName := r.URL.Path[1:]

		queryStr := generateURL(ingredientName)

		req, _ := http.NewRequest("GET", "http://34.67.111.154:7777/"+queryStr, nil)

		// The trace ID from the incoming request will be
		// propagated to the outgoing request.
		req = req.WithContext(r.Context())

		// The outgoing request will be traced with r's trace ID.
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("service B:", resp.Body)

		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			panic(err)
		}

		_, _ = io.WriteString(w, string(body))

		resp.Body.Close()
	})
	http.Handle("/", handler)

	port := os.Getenv("PORT")
	port = "7777"
	// if port == "" {
	// 	port = "8080"
	// }
	log.Printf("Listening on port %s", port)

	// Use an ochttp.Handler in order to instrument OpenCensus for incoming
	// requests.
	httpHandler := &ochttp.Handler{
		// Use the Google Cloud propagation format.
		Propagation: &propagation.HTTPFormat{},
	}
	if err := http.ListenAndServe(":"+port, httpHandler); err != nil {
		log.Fatal(err)
	}
}

// [END trace_setup_go_quickstart]
