package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/mitchellh/go-homedir"
)

const (
	ipCheckURL    = "https://ieserver.net/ipcheck.shtml"
	timeout       = time.Second * 10
	ddnsUpdateURL = "https://ieserver.net/cgi-bin/dip.cgi"
)

var httpClient = http.Client{Timeout: time.Duration(timeout)}

func getGlobalIPValue() string {
	resp, err := httpClient.Get(ipCheckURL)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalln(resp.Status)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	return string(b)
}

type config struct {
	Account string
	Domain  string
	Passwd  string
}

func readConfig(cfgPath string) config {
	configPath := ""
	if cfgPath == "" {
		dir, err := homedir.Dir()
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		configPath = path.Join(dir, ".ieserver.cfg")
	} else {
		configPath = cfgPath
	}

	if _, err := os.Stat(configPath); err != nil {
		log.Fatalf("%v\n", err)
	}

	buf, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	cfg := config{}
	if err = json.Unmarshal(buf, &cfg); err != nil {
		log.Fatalf("%v\n", err)
	}
	return cfg
}

func registerIP(ip string, cfg config) {
	val := url.Values{
		"username":   {cfg.Account},
		"domain":     {cfg.Domain},
		"password":   {cfg.Passwd},
		"updatehost": {"1"},
	}
	resp, err := httpClient.PostForm(ddnsUpdateURL, val)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatalln(resp.Status)
	}
	log.Printf("done")
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	cfgPath := flag.String("c", "", "path to config file")
	flag.Parse()
	cfg := readConfig(*cfgPath)

	ipString := getGlobalIPValue()
	if ipString == "" {
		log.Fatalf("can't fetch ip")
	}
	registerIP(ipString, cfg)
}
