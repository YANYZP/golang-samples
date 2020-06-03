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
	"log"
	"net/http"
	"os"
	"strings"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

func readFile(fileName string) map[string]map[string]string {

	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var itemPriceMap = make(map[string]map[string]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		priceInfo := scanner.Text()
		words := strings.Fields(priceInfo)
		fmt.Println(priceInfo)
		if len(words) != 3 {
			fmt.Println("Wrong format")
			continue
		}
		if itemPriceMap[words[1]] == nil {
			itemPriceMap[words[1]] = make(map[string]string)
		}
		itemPriceMap[words[1]][words[0]] = words[2]
	}

	// for k, v := range itemPriceMap {
	// 	for _, vv := range v {
	// 		fmt.Println(k, v, vv)
	// 	}
	// }

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return itemPriceMap
}

func getPriceInfo(itemVendorStr string) string {
	itemPriceMap := readFile("price.txt")
	fmt.Println("service c: itemVendorStr = " + itemVendorStr)
	infoArray := strings.Split(itemVendorStr, "/")

	if len(infoArray) < 2 {
		return "Service C fails to find enough info\n"
	}
	itemName := infoArray[0]

	vendorPriceMap, ok := itemPriceMap[itemName]

	if !ok {
		return "Service C: Not finding vendors for this item\n"
	}

	vendorPriceStrBuilder := strings.Builder{}

	for i := 1; i < len(infoArray); i++ {
		vendorName := infoArray[i]
		price, okok := vendorPriceMap[vendorName]

		if !okok {
			fmt.Println("service c: fail to find price of" + itemName + " in " + vendorName)
		} else {
			vendorPriceStrBuilder.WriteString(price + " dollar at " + vendorName + "\n")
		}

	}
	return vendorPriceStrBuilder.String()
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

	// client := &http.Client{
	// 	Transport: &ochttp.Transport{
	// 		// Use Google Cloud propagation format.
	// 		Propagation: &propagation.HTTPFormat{},
	// 	},
	// }

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		itemVendorStr := r.URL.Path[1:]

		req, _ := http.NewRequest("GET", "https://www.google.com", nil)

		// The trace ID from the incoming request will be
		// propagated to the outgoing request.
		req = req.WithContext(r.Context())

		//TODO: w writeString
		_, _ = io.WriteString(w, getPriceInfo(itemVendorStr))

		// The outgoing request will be traced with r's trace ID.
		// resp, err := client.Do(req)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// log.Println("service A:", resp.Body)

		// // Because we don't read the resp.Body, need to manually call Close().
		// resp.Body.Close()
	})
	http.Handle("/", handler)

	port := os.Getenv("PORT")
	port = "7777"
	if port == "" {
	}
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
