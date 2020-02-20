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

func QueryEndpoint(endpoint string) (metrics []byte, err error) {
	client := new(http.Client)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "curl/7.64.1")
	req.Header.Set("Accept-Encoding", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}


	metrics, err = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		log.Fatal(err)
	}
	return metrics, nil
}

func ParseMetrics(metrics []byte) (total int, errors int, err error) {
	// https://stackoverflow.com/questions/40429296/converting-string-to-json-or-struct
	m := []byte(metrics)
	JQ := &JQResponse{}
	if err := json.Unmarshal(m, JQ); err != nil {
		return 0, 0, err
	}
	return JQ.JQResponse.Totals, JQ.JQResponse.Errors, nil
}

func SendMetrics(total int, errors int, prefix string, env string) (err error) {
	conn, err := net.Dial("udp", "localhost:8125")
	if err != nil {
		return err
	}
	total_txt := []string{prefix, env, ".total:", strconv.Itoa(total), "|c"}
	errors_txt := []string{prefix, env, ".errors:", strconv.Itoa(errors), "|c"}
	defer conn.Close()
	fmt.Fprintf(conn, (strings.Join(total_txt, "") + "\n"))
	fmt.Fprintf(conn, (strings.Join(errors_txt, "") + "\n"))
	return err
}

func main() {
	conf, err := LoadConfig("jqm.json")
	if err != nil {
		log.Fatal(err)
	}

	completeUrl := (conf.Jqm_config.Url + conf.Jqm_config.Path + conf.Jqm_config.ApiKey)

	returnedMetrics, err := QueryEndpoint(completeUrl)
	if err != nil {
		log.Fatal(err)
	}

	totals, errors, err := ParseMetrics(returnedMetrics)
	if err != nil {
		log.Fatal(err)
	}

	err = SendMetrics(totals, errors, conf.Jqm_config.Prefix, conf.Jqm_config.Env)
	if err != nil {
		log.Fatal(err)
	}
}
