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
		log.Printf(
			"[%s] %d hits; took: %dms",
			res.Status(),
			int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
			int(r["took"].(float64)),
		)
		for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
			log.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
		}

		log.Println(strings.Repeat("=", 37))
	}

}
