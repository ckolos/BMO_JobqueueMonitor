package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Config file format
// all fields are required
type Config struct {
	Jqm_config struct {
		Url    string `json:"url"`
		ApiKey string `json:"api_key"`
		Path   string `json:"path"`
		Prefix string `json:"prefix"`
		Env    string `json:"env"`
	} `json:"config"`
}

// BMO's jobqueue server responds with a JSON object
// containing only total and errors
// {"total": <int>, "errors": <int>}
type JQResponse struct {
	JQResponse struct {
		Totals int `json:"total"`
		Errors int `json:"errors"`
	} `json:"response"`
}

func LoadConfig(filename string) (config Config, err error) {
	configFile, err := os.Open(filename)
	defer configFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	return config, err
}

func QueryEndpoint(endpoint string) (metrics []byte) {
	client := new(http.Client)
	req, _ := http.NewRequest("GET", endpoint, nil)
	req.Header.Set("User-Agent", "curl/7.64.1")
	req.Header.Set("Accept-Encoding", "application/json")
	resp, err := client.Do(req)

	metrics, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return
}

func ParseMetrics(metrics []byte) (total int, errors int) {
	// https://stackoverflow.com/questions/40429296/converting-string-to-json-or-struct
	m := []byte(metrics)
	JQ := &JQResponse{}
	if err := json.Unmarshal(m, JQ); err != nil {
		log.Fatal(err)
	}
	return JQ.JQResponse.Totals, JQ.JQResponse.Errors
}

func SendMetrics(total int, errors int, prefix string, env string) {
	conn, err := net.Dial("udp", "localhost:8125")
	if err != nil {
		log.Fatal(err)
	}
	total_txt := []string{prefix, env, ".total:", strconv.Itoa(total), "|c"}
	errors_txt := []string{prefix, env, ".errors:", strconv.Itoa(errors), "|c"}
	fmt.Fprintf(conn, (strings.Join(total_txt, "") + "\n"))
	fmt.Fprintf(conn, (strings.Join(errors_txt, "") + "\n"))
	_ = conn.Close()
}

func main() {
	conf, _ := LoadConfig("jqm.json")
	completeUrl := (conf.Jqm_config.Url + conf.Jqm_config.Path + conf.Jqm_config.ApiKey)
	returnedMetrics := QueryEndpoint(completeUrl)
	totals, errors := ParseMetrics(returnedMetrics)
	SendMetrics(totals, errors, conf.Jqm_config.Prefix, conf.Jqm_config.Env)
}
