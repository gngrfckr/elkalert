package main

import (
	"bytes"
	"context"
	"elkalert/src/config"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
)

var (
	confPath string
	conf     = config.NewConfig()
)

//Rules an array of rules
type Rules struct {
	Rules []Rule `json:"rules"`
}

//Rule ...
type Rule struct {
	Name     string `json:"name"`
	Index    string `json:"index"`
	Query    interface{}
	Interval string `json:"interval"`
}

func init() {
	flag.StringVar(&confPath, "config-path", "configs/config.toml", "path to config file")
}

func constructQuery(q string, size int) *strings.Reader {
	var query = `{"query": {`
	query = query + q
	query = query + `}, "size": ` + strconv.Itoa(size) + `}`
	fmt.Println("\nquery:", query)

	isValid := json.Valid([]byte(query)) // returns bool

	if isValid == false {
		fmt.Println("constructQuery() ERROR: query string not valid:", query)
		fmt.Println("Using default match_all query")
		query = "{}"
	} else {
		fmt.Println("constructQuery() valid JSON:", isValid)
	}
	var b strings.Builder
	b.WriteString(query)

	read := strings.NewReader(b.String())

	return read
}

func readRulesFile(path string) Rules {
	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened rules file")
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var rules Rules

	json.Unmarshal(byteValue, &rules)

	return rules
}

func main() {
	var (
		r map[string]interface{}
		// wg sync.WaitGroup
	)
	flag.Parse()
	_, err := toml.DecodeFile(confPath, conf)

	es7, err := elasticsearch7.NewDefaultClient()
	res, err := es7.Info()
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	var results = readRulesFile("rules.json")

	for i := 0; i < len(results.Rules); i++ {
		var buf bytes.Buffer

		fmt.Println("User Name: " + results.Rules[i].Name)
		fmt.Println(results.Rules[i].Query)

		query := map[string]interface{}{
			"query": results.Rules[i].Query,
		}

		if err := json.NewEncoder(&buf).Encode(query); err != nil {
			log.Fatalf("Error encoding query: %s", err)
		}
		res, err = es7.Search(
			es7.Search.WithContext(context.Background()),
			es7.Search.WithIndex(results.Rules[i].Index),
			es7.Search.WithBody(&buf),
			es7.Search.WithTrackTotalHits(true),
			es7.Search.WithPretty(),
		)
		if err != nil {
			log.Fatalf("Error getting response: %s", err)
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
		log.Printf(
			"[%s] %d hits; took: %dms",
			res.Status(),
			int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
			int(r["took"].(float64)),
		)
		// Print the ID and document source for each hit.
		for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
			log.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
		}

		log.Println(strings.Repeat("=", 37))
	}

}
