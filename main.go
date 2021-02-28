package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type configuration map[string]string

type tags struct {
	Name string `json:"name"`
	Tags []int64 `json:"tags"`
}

func main() {
	props, err := parseConfiguration("config.properties")
	if err != nil {
		fmt.Println(err.Error())
	}

	log.Println("Starting the cleanup!")

	// Log rules
	for key, element := range props {
		log.Println("Deleting " + key + " except the last " + element + " images")
	}

	// Since minikube ca is not trusted, an insecure client is created
	tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
		client := &http.Client{Transport: tr}

	for key, element := range props {
		getTags(key, client)
		log.Print(element)
	}
}

// Returns a list of tags from registry API
func getTags(image string, client *http.Client) {
	// Get tags from registry API
	r, err := client.Get("https://myregistry.foo/v2/" + image + "/tags/list")
	if err != nil {
			log.Fatal(err.Error())
	}
	if(r.StatusCode == 200) {
		// Parse response body
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Println(err.Error())
		}
		
		obj := map[string]interface{}{}
		if err := json.Unmarshal([]byte(b), &obj); err != nil {
				log.Fatal(err)
		}

		fmt.Println(obj)
	}
}

// Parse config file to map
func parseConfiguration(filename string) (configuration, error) {
	config := configuration{}

	if len(filename) == 0 {
		return config, nil
	}

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// ignore comment lines (starting with #)
		if strings.HasPrefix(line, "#") {
			continue
		}
		// check if line has proper format and values
		if equal := strings.Index(line, "="); equal > 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}
				config[key] = value
			}
		}
	}

	return config, err
}