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
	"sort"
	"strconv"
	"strings"
	"time"
)

func main() {
	host := os.Getenv("REGISTRY_HOST") // use https://myregistry.foo/v2/ with https://github.com/villevaltonen/registry-k8s
	insecure, err := strconv.ParseBool(os.Getenv("INSECURE_REGISTRY"))

	// HTTP-client basic config
	var client = &http.Client{
		Timeout: 10 * time.Second,
	}

	// configure insecure client, if needed
	if err == nil && insecure {
		log.Println("Using insecure HTTP-client")
		tr := &http.Transport{
        	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    	}
		client = &http.Client{Transport: tr}
	}

	// parse retention rules
	props, err := parseRules("config.properties")
	if err != nil {
		fmt.Println(err.Error())
	}

	// delete images
	deleteImages(host, props, client)
}

type configuration map[string]string

type tags struct {
	Name string `json:"name"`
	Tags []string `json:"tags"`
}

type digest struct {
	Config struct {
		MediaType string `json:"mediaType"`
		Size int `json:"size"`
		Digest string `json:"digest"`
	}
}

func deleteImages(host string,  props configuration, client *http.Client) {
	for image, count := range props {
		log.Println("Deleting all " + image + " images except the last " + count + " images")
		// get image tags
		tags := getTags(host, image, client)

		// convert tags to int and sort them
		t := make([]int, len(tags))
		for i, tag := range tags {
			ti, err := strconv.Atoi(tag)
			if err != nil {
				log.Println(err.Error())
			} else {
				t[i] = ti
			}
		}
		sort.Ints(t)

		// get a subset of tags to be deleted based on saved count
		c, err :=  strconv.Atoi(count)
		if err != nil {
			log.Println(err.Error())
		} else {
			// if more or as many as current amount of tags are meant to be left into the registry, skip rest of the logic
			if c < len(tags) {
				save := len(tags) - c 
				result := t[:save]

				// delete images
				for _, tag := range result {
					digest := getDigest(host, image, tag, client)
					deleteImage(host, image, digest, client)
				}	
			}
		}
	}
	log.Println("Clean up finished!")
}

// Deletes an image with digest
func deleteImage(host, image, digest string, client *http.Client) {
	url := host + image + "/manifests/" + digest
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Println(err.Error())
	}

	r, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
	}
	if(r.StatusCode >= 200 && r.StatusCode < 300) {
		log.Println("Digest " + digest + " deleted!")
	} else {
		log.Println("Manifest " + digest + " is waiting for garbage collection or it cannot be deleted")
	}
}

// Gets the digest for tag
func getDigest(host, image string, tag int, client *http.Client) string {
	ts := strconv.Itoa(tag)
	url := host + image + "/manifests/" + ts
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err.Error())
	}

	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	r, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
	}

	// parse tags from response
	var digest = digest{}
	if(r.StatusCode >= 200 && r.StatusCode < 300) {
		// parse response body
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Println(err.Error())
		}

		if err := json.Unmarshal([]byte(b), &digest); err != nil {
				log.Fatal(err)
		}
	} else {
		log.Println("Couldn't find any tags for image " + image)
	}

	return digest.Config.Digest
}

// Returns a list of tags from registry API
func getTags(host, image string, client *http.Client) []string {
	// get tags from registry API
	r, err := client.Get(host + image + "/tags/list") // maybe an env variable for host?
	if err != nil {
			log.Fatal(err.Error())
	}

	// parse tags from response
	var tags = tags{}
	if(r.StatusCode == 200) {
		// parse response body
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Println(err.Error())
		}

		if err := json.Unmarshal([]byte(b), &tags); err != nil {
				log.Fatal(err)
		}
	} else {
		log.Println("Couldn't find any tags for image " + image)
	}

	return tags.Tags
}

// Parse config file to a map
func parseRules(filename string) (configuration, error) {
	log.Println("Parsing configuration")
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