package main

import (
	"bytes"
	"context"
	alert "elkalert/src/alert"
	"elkalert/src/config"

	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/mileusna/crontab"
)

var (
	confPath string
	conf     = config.NewConfig()
	es7, err = elasticsearch7.NewDefaultClient()
	wg       sync.WaitGroup
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
	Interval Interval
	Alert    alert.Alert
}

//Interval ...
type Interval struct {
	Minutes string `json:"minutes"`
	Hours   string `json:"hours"`
}

func init() {
	flag.StringVar(&confPath, "config-path", "configs/config.toml", "path to config file")
}

func readRulesFile(path string) Rules {
	jsonFile, err := os.Open(path)
	if err != nil {
		log.Println(err)
	}
	log.Println("Successfully Opened rules file")
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var rules Rules

	json.Unmarshal(byteValue, &rules)

	return rules
}

func notify() {
	fmt.Println("done")
}

func search(wg *sync.WaitGroup, rule Rule) {
	defer wg.Done()
	var (
		r        map[string]interface{}
		res, err = es7.Info()
	)

	var buf bytes.Buffer

	log.Println("Rule Run: " + rule.Name)

	query := map[string]interface{}{
		"query": rule.Query,
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}
	res, err = es7.Search(
		es7.Search.WithContext(context.Background()),
		es7.Search.WithIndex(rule.Index),
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
		alert.SendAlert(rule.Alert, hit.(map[string]interface{}))
	}
	log.Println(strings.Repeat("=", 37))
}

func searchJobStartWrapper(rule Rule) {
	wg.Add(1)
	search(&wg, rule)
}

func main() {
	c := make(chan struct{})
	flag.Parse()
	_, err := toml.DecodeFile(confPath, conf)
	ctab := crontab.New()
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}
	rules := readRulesFile(conf.RulesPath)
	for i := 0; i < len(rules.Rules); i++ {
		ctab.MustAddJob("*/"+string(rules.Rules[i].Interval.Minutes)+" * * * *", searchJobStartWrapper, rules.Rules[i])
	}
	wg.Wait()
	<-c
}
