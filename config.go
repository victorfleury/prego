package main

import (
	"encoding/json"
	"log"
	"os"
)

type ConfigPayload struct {
	Editor        string
	All_reviewers []map[string]map[string]string
	My_reviewers  []map[string]map[string]string
}

var default_config_payload string = `
{
	"all_reviewers": [
		{"user": {"name": "agjolly"}},
		{"user": {"name": "akasimov"}},
		{"user": {"name": "alima"}},
		{"user": {"name": "bramoul"}},
		{"user": {"name": "csarrazin"}},
		{"user": {"name": "cslimani"}},
		{"user": {"name": "dguillemette"}},
		{"user": {"name": "gnahmias"}},
		{"user": {"name": "jpepingagne"}},
		{"user": {"name": "jdubuisson"}},
		{"user": {"name": "ldepoix"}},
		{"user": {"name": "lchikar"}},
		{"user": {"name": "mapaquin"}},
		{"user": {"name": "opeloquin"}},
		{"user": {"name": "pviolant"}},
		{"user": {"name": "rpresset"}},
		{"user": {"name": "tcarpentier"}},
		{"user": {"name": "vfleury"}}
	]
}`

// Parse the JSON config for prego
func parse_config() ConfigPayload {

	root_path_to_config := os.Getenv("XDG_CONFIG_HOME")
	if root_path_to_config == "" {
		root_path_to_config = os.Getenv("HOME") + "/.config"
	}

	path_to_config := root_path_to_config + "/prego/prego.json"
	log.Println("Path to config", path_to_config)

	config, err := os.ReadFile(path_to_config)
	if err != nil {
		log.Println("Could not read the config file at ", path_to_config, "Using default")
		config = []byte(default_config_payload)
	}

	var config_payload ConfigPayload

	err = json.Unmarshal(config, &config_payload)
	if err != nil {
		log.Fatal("Something is wrong with the configuration. It could not be parsed fully.")
	}

	return config_payload

}
