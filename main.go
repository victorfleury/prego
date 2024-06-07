/*
* TODO :
- Fetch all branches from the repo. -> OK
  - Add dev and main first, then the rest of the branches

- Default reviewers :
  - Use a map for reviewers -> OK
  - Read from a JSON config file -> WIP
  - Add cli arg to remove them altogether -> TODO

- Use Go Git package -> OK
- Read Token from file -> OK
- Build payload for POST request -> OK
- Execute payload successfully -> OK
- Add CLI command to generate the config
*
*/
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"

	"net/http"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

const URL_TEMPLATE string = "https://bitbucket.rodeofx.com/rest/api/1.0/projects/%s/repos/%s/pull-requests"

type ConfigPayload struct {
	Editor        string
	All_reviewers []map[string]map[string]string
	My_reviewers  []map[string]map[string]string
}

var (
	destination_branch string
	title              string
	url                string
	reviewers          []string
)

var PR_TEMPLATE string = `#### Purpose of the PR

#### Type of feedback wanted

#### Potential risks of this change

#### Relationship with other PRs
`

func check(branches_name []string, name string) bool {
	for _, b := range branches_name {
		if name == b {
			fmt.Println("Found")
			return true
		}
	}
	return false
}

func main() {

	config := parse_config()
	repo, err := get_repo()
	if err != nil {
		log.Fatal("Prego needs to be run in a Git repository !")
	}

	// Branches
	current_branch, _ := repo.Head()
	branches, err := repo.Branches()
	if err != nil {
		log.Fatal("No branches found. Are you in a properly initialized repository?")
	}

	var branch_names = []string{"dev", "master"}
	branches.ForEach(func(b *plumbing.Reference) error {
		short_name := strings.Split(b.String(), "refs/heads/")[1]
		if short_name != current_branch.Name().Short() && !check(branch_names, short_name) {
			branch_names = append(branch_names, short_name)
		}
		return nil
	})

	var branch_names_cleaned []string
	for _, b := range branch_names {
		if b != "" {
			branch_names_cleaned = append(branch_names_cleaned, b)
		}
	}

	branch_options := make([]huh.Option[string], len(branch_names_cleaned))
	for i, branch := range branch_names {
		branch = strings.Trim(branch, "* ")
		if branch != "" {
			branch_options[i] = huh.NewOption(branch, branch)
		}
	}

	// Reviewers
	reviewers_option := make([]huh.Option[string], len(config.All_reviewers))
	for i, reviewer := range config.All_reviewers {
		selected := reviewer_in_prefs(config, reviewer)
		reviewers_option[i] = huh.NewOption(reviewer["user"]["name"], reviewer["user"]["name"]).Selected(selected)
	}

	var confirm bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose the destination branch :").
				Description("The branch you want to merge your changes to.").
				Options(
					branch_options...,
				).
				Value(&destination_branch),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select reviewers").
				Description("Pick which team members should review your PR").
				Options(reviewers_option...).
				Value(&reviewers),
		),
		huh.NewGroup(
			huh.NewText().
				Value(&PR_TEMPLATE).
				Title("PR Description").
				Editor(config.Editor).
				Lines(15).
				CharLimit(5000).
				Description("Content of the PR"),
			huh.NewConfirm().Title("Publish PR").Affirmative("Yes !").Negative("Cancel").Value(&confirm),
		),
	)

	err = form.Run()

	if err != nil {
		fmt.Println("Uh oh:", err)
		os.Exit(1)
	}

	if !confirm {
		log.Println("Publish PR aborted...")
		os.Exit(0)
	}

	_ = spinner.New().Title("Publishing PR...").Accessible(false).Action(publish_pr).Run()
}

// Entrypoint for huh to publish the PR
func publish_pr() {

	log.Println("Fetching token")
	repo_url := get_repo_url()

	reviewers_payload_data := build_reviewers_payload_data(reviewers)

	repo, _ := get_repo()
	head_ref, _ := repo.Head()
	commit_message, _ := repo.CommitObject(head_ref.Hash())
	title := strings.Split(commit_message.Message, "\n")[0]

	json_payload := build_payload_request(
		PR_TEMPLATE,
		string(head_ref.Name()),
		destination_branch,
		title,
		reviewers_payload_data,
	)
	result := publish_pr_request(repo_url, json_payload)
	if result {
		log.Println("Success !")
	} else {
		log.Fatal("Could not publish PR ...")
	}

}

// Build the reviewers bit of the payload. Huh gives us back some string for reviewers, we need to make a slice of maps out of it.
func build_reviewers_payload_data(reviewers []string) []map[string]map[string]string {
	var selected_reviewers []map[string]map[string]string

	for _, reviewer := range reviewers {
		payload_reviewer := map[string]map[string]string{"user": {"name": reviewer}}
		selected_reviewers = append(selected_reviewers, payload_reviewer)
	}

	return selected_reviewers
}

// Get the token from the token file
// TODO: Add it in the config and fetch it from there.
func get_token() string {

	token_path := os.Getenv("HOME") + "/token.tk"
	token, err := os.ReadFile(token_path)
	if err != nil {
		log.Fatal("Panic ! No token found")
	}
	return strings.Trim(string(token), "\n")
}

// Get the git repository from the current working directory
func get_repo() (*git.Repository, error) {
	current_directory, err := os.Getwd()
	if err != nil {
		log.Fatal("Current directory could not be found?")
	}

	options := git.PlainOpenOptions{DetectDotGit: true}
	repo, err := git.PlainOpenWithOptions(current_directory, &options)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

// Get the remote url for the repo to extract the project name and slug name so we can build the REST API request url
func get_repo_url() string {
	repo, _ := get_repo()
	remotes, err := repo.Remotes()
	if err != nil || len(remotes) == 0 {
		log.Fatal("Repository has no remote...")
	}
	remote := remotes[0]
	config_url := remote.Config().URLs[0]

	parts := strings.Split(config_url, "/")
	project, slug_name := parts[len(parts)-2], parts[len(parts)-1]

	formatted_url := fmt.Sprintf(URL_TEMPLATE, project, strings.Split(slug_name, ".git")[0])

	log.Println("Formattted url :", formatted_url)

	return formatted_url
}

// Build the Payload for the request to the Bitbucket REST API
func build_payload_request(description, source_branch, destination_branch, title string, reviewers []map[string]map[string]string) []byte {

	data := map[string]interface{}{
		"description": description,
		"fromRef": map[string]interface{}{
			"id": source_branch,
		},
		"toRef": map[string]interface{}{
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
	log.Printf("Payload: %s\n", json_payload)
	return json_payload
}

// Perform the HTTP Request to the Bitbucket REST API
func publish_pr_request(url string, json_payload []byte) bool {

	log.Println("Publishing to ", url)

	client := http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewReader(json_payload))

	req.Header = http.Header{
		"content-type":  {"application/json"},
		"authorization": {fmt.Sprintf("Bearer %s", get_token())},
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal("Could not publish PR ...", err)
	}
	defer res.Body.Close()

	log.Println("Request :", res.Request)
	log.Println("Status code of the request :", res.StatusCode)

	if res.StatusCode == 201 {
		return true
	} else {
		return false
	}
}

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
		log.Println("Could not read the config file at ", path_to_config)
	}

	var config_payload ConfigPayload

	err = json.Unmarshal(config, &config_payload)
	if err != nil {
		log.Fatal("Something is wrong with the configuration. It could not be parsed fully.")
	}

	return config_payload

}

// Check if a given reviewer is in the config for default reviewers
func reviewer_in_prefs(config ConfigPayload, reviewer map[string]map[string]string) bool {
	for _, r := range config.My_reviewers {
		if reflect.DeepEqual(r, reviewer) {
			return true
		}
	}
	return false
}
