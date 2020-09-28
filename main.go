package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func IndexIndefinitely() {
	es, _ := elasticsearch.NewDefaultClient()
	i := 0
	title := ""

	for {
		title = fmt.Sprintf("Test %d - %d", i, rand.Int())
		i = i + 1

		var b strings.Builder
		b.WriteString(`{"title":"`)
		b.WriteString(title)
		b.WriteString(`"}`)

		req := esapi.IndexRequest{
			Index:      "test",
			DocumentID: strconv.Itoa(i + 1),
			Body:       strings.NewReader(b.String()),
			Refresh:    "true",
		}

		res, err := req.Do(context.Background(), es)
		if err != nil {
			log.Printf("Error getting index response: %s", err)
		}
		defer res.Body.Close()

		if res.IsError() {
			log.Printf("[%s] Error indexing document ID=%d", res.Status(), i+1)
		} else {
			// Deserialize the response into a map.
			var r map[string]interface{}
			if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
				log.Printf("Error parsing the response body: %s", err)
			} else {
				if i%100 == 0 {
					// Print the response status and indexed document version.
					log.Printf("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
				}
			}
		}
	}
}

func SearchIndefinitely() {
	var (
		buf bytes.Buffer
		r   map[string]interface{}
	)

	i := 0
	es, _ := elasticsearch.NewDefaultClient()
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"title": fmt.Sprintf("test %d", rand.Intn(100)),
			},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Printf("Error encoding query: %s", err)
	}

	for {
		i = i + 1
		res, err := es.Search(
			es.Search.WithContext(context.Background()),
			es.Search.WithIndex("test"),
			es.Search.WithBody(&buf),
			es.Search.WithTrackTotalHits(true),
			es.Search.WithPretty(),
		)
		if err != nil {
			log.Fatalf("Error getting search response: %s", err)
		}
		defer res.Body.Close()

		if res.IsError() {
			var e map[string]interface{}
			if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
				log.Fatalf("Error parsing the response body: %s", err)
			} else {
				// Print the response status and error information.
				log.Fatalf("[%s] %s: %s",
					res.Status(),
					e["error"].(map[string]interface{})["type"],
					e["error"].(map[string]interface{})["reason"],
				)
			}
		}

		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		}

		// Print the response status, number of results, and request duration.
		if i%500 == 0 {
			log.Printf(
				"[%s] %d hits; took: %dms",
				res.Status(),
				int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
				int(r["took"].(float64)),
			)
		}
	}
}

func main() {
	go IndexIndefinitely()
	go SearchIndefinitely()

	for {
	}
}
