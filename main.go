package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

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
}

type configuration map[string]string

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