package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"slices"
	"strings"
)

type ConfigPayload struct {
	Editor        string
	All_reviewers []map[string]map[string]string
	My_reviewers  []map[string]map[string]string
}

var Default_config string = `
{
	"all_reviewers": [
		{"user": {"name": "agjolly"}},
		{"user": {"name": "bramoul"}},
		{"user": {"name": "jpepingagne"}},
		{"user": {"name": "jdubuisson"}},
		{"user": {"name": "mapaquin"}},
		{"user": {"name": "opeloquin"}},
		{"user": {"name": "pviolanti"}},
		{"user": {"name": "rbousquet"}},
		{"user": {"name": "tcarpentier"}},
		{"user": {"name": "sribet"}},
		{"user": {"name": "nsingh"}}
	]
}`

func Default_config_payload() ConfigPayload {
	var default_config_json ConfigPayload
	err := json.Unmarshal([]byte(Default_config), &default_config_json)
	if err != nil {
		log.Fatal("Something is wrong with the configuration. It could not be parsed fully.")
	}
	return default_config_json
}

// Check if a given reviewer is in the config for default reviewers
func Reviewer_in_prefs(config ConfigPayload, reviewer map[string]map[string]string) bool {
	for _, r := range config.My_reviewers {
		if reflect.DeepEqual(r, reviewer) {
			return true
		}
	}
	return false
}

// Check if a given branch name is in a list of branch names
func IsNameInNames(branches_name []string, name string) bool {
	if slices.Contains(branches_name, name) {
		return true
	}
	return false
	//for _, b := range branches_name {
	//if name == b {
	//return true
	//}
	//}
	//return false
}

// Read token information from the source file.
func Get_token() string {

	token_path := os.Getenv("HOME") + "/token.tk"
	token, err := os.ReadFile(token_path)
	if err != nil {
		log.Fatal("Panic ! No token found : ", err)
	}
	return strings.Trim(string(token), "\n")
}

// Build the Payload for the request to the Bitbucket REST API
func Build_payload_request(description, source_branch, destination_branch, title string, reviewers []map[string]map[string]string) []byte {

	data := map[string]any{
		"description": description,
		"fromRef": map[string]any{
			"id": source_branch,
		},
		"toRef": map[string]any{
			"id": fmt.Sprintf("refs/heads/%s", destination_branch),
		},
		"state":     "OPEN",
		"title":     title,
		"reviewers": reviewers,
	}

	json_payload, err := json.Marshal(data)
	if err != nil {
		log.Fatal("Could not build JSON payload for creating PR.")
	}
	//log.Printf("Payload: %s\n", json_payload)
	return json_payload
}

// Build the reviewers bit of the payload. Huh gives us back some string for reviewers, we need to make a slice of maps out of it.
func Build_reviewers_payload_data(reviewers []string) []map[string]map[string]string {
	var selected_reviewers []map[string]map[string]string

	for _, reviewer := range reviewers {
		payload_reviewer := map[string]map[string]string{"user": {"name": reviewer}}
		selected_reviewers = append(selected_reviewers, payload_reviewer)
	}

	return selected_reviewers
}
