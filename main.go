package main

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "os"
)

type Config struct {
  Jqm_config struct {
    Url string `json:"url"`
    ApiKey string `json:"api_key"`
    Path string `json:"path"`
  } `json:"config"`
}

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

func SendMetrics(total int, errors int) {
  // TODO publish to localhost:8125 for statsd

}

func main() {
  config, _ := LoadConfig("jqm.json")
  completeUrl := (config.Jqm_config.Url + config.Jqm_config.Path + config.Jqm_config.ApiKey)
  returnedMetrics := QueryEndpoint(completeUrl)
  totals, errors := ParseMetrics(returnedMetrics)

  fmt.Printf("%d %d\n", totals, errors)
}
