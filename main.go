/*
* TODO :
- Fetch all branches from the repo. -> OK
  - Add dev and main first, then the rest of the branches

- Default reviewers :
  - Use a map for reviewers -> OK
  - Read from a JSON config file -> TODO
  - Add cli arg to remove them altogether -> TODO

- Use Go Git package -> OK
- Read Token from file -> OK
- Build payload for POST request -> Wip need to add the HEADERs
- Execute payload successfully
*
*/
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

const URL_TEMPLATE string = "https://bitbucket.rodeofx.com/rest/api/1.0/projects/%s/repos/%s/pull-requests"

const USER_URL_TEMPLATE = "https://bitbucket.rodeofx.com/rest/api/1.0/users/vfleury/repos/hom_repo/pull_requests"

// fmt.Println(json_payload)
type Reviewer struct {
	Name string
}

var (
	destination_branch string
	title              string
	url                string
	reviewers          []string
)
var DEFAULT_REVIEWERS = []map[string]map[string]string{
	{"user": {"name": "bramoul"}},
	{"user": {"name": "agjolly"}},
	{"user": {"name": "jdubuisson"}},
	{"user": {"name": "alima"}},
	{"user": {"name": "lchikar"}},
	{"user": {"name": "ldepoix"}},
	{"user": {"name": "gnahmias"}},
	{"user": {"name": "opeloquin"}},
	{"user": {"name": "rpresset"}},
}

var PR_TEMPLATE string = `#### Purpose of the PR

#### Overview of the changes

#### Type of feedback wanted

#### Where should the reviewer start looking at?

#### Potential risks of this change

#### Relationship with other PRs
`

func main() {

	repo, err := get_repo()
	if err != nil {
		log.Fatal("Prego needs to be run in a Git repository !")
	}

	// Branches
	//var destination_branch string

	branches, err := repo.Branches()
	if err != nil {
		log.Fatal("No branches found. Are you in a properly initialized repository?")
	}

	var branch_names []string
	branches.ForEach(func(b *plumbing.Reference) error {
		short_name := strings.Split(b.String(), "refs/heads/")[1]
		branch_names = append(branch_names, short_name)
		return nil
	})
	//branch_names := strings.Split(string(branches), "\n")
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
	reviewers_option := make([]huh.Option[string], len(DEFAULT_REVIEWERS))
	for i, reviewer := range DEFAULT_REVIEWERS {
		reviewers_option[i] = huh.NewOption(reviewer["user"]["name"], reviewer["user"]["name"]).Selected(true)
	}

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
				Lines(15).
				Description("Content of the PR"),
			huh.NewConfirm().Title("Publish PR").Affirmative("Yes !").Negative("Cancel"),
		),
	)

	err = form.Run()

	if err != nil {
		fmt.Println("Uh oh:", err)
		os.Exit(1)
	}

	_ = spinner.New().Title("Publishing PR...").Accessible(false).Action(publish_pr).Run()
}

func publish_pr() {
	time.Sleep(1 * time.Second)

	log.Println("Fetching token")
	//token := get_token()
	repo_url := get_repo_url()

	reviewers_payload_data := build_reviewers_payload_data(reviewers)

	repo, _ := get_repo()
	head_ref, _ := repo.Head()
	title, _ := repo.CommitObject(head_ref.Hash())

	json_payload := build_payload_request(
		PR_TEMPLATE,
		destination_branch,
		title.Message,
		reviewers_payload_data,
	)
	result := publish_pr_request(repo_url, json_payload)
	if result {
		fmt.Println("Success !")
	} else {
		log.Fatal("Could not publish PR ...")
	}

}

func build_reviewers_payload_data(reviewers []string) []map[string]map[string]string {
	var selected_reviewers []map[string]map[string]string

	for _, reviewer := range reviewers {
		payload_reviewer := map[string]map[string]string{"user": {"name": reviewer}}
		selected_reviewers = append(selected_reviewers, payload_reviewer)
	}

	return selected_reviewers
}

func get_token() string {

	token_path := os.Getenv("HOME") + "/token.tk"
	token, err := os.ReadFile(token_path)
	if err != nil {
		log.Fatal("Panic ! No token found")
	}
	return string(token)
}

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

func get_repo_url() string {
	repo, _ := get_repo()
	remotes, err := repo.Remotes()
	if err != nil || len(remotes) == 0 {
		log.Fatal("Repository has no remote...")
	}
	//fmt.Println("Remotes", remotes)
	remote := remotes[0]
	config_url := remote.Config().URLs[0]

	parts := strings.Split(config_url, "/")
	project, slug_name := parts[len(parts)-2], parts[len(parts)-1]

	formatted_url := fmt.Sprintf(URL_TEMPLATE, project, strings.Split(slug_name, ".git")[0])

	log.Println(formatted_url)

	return formatted_url
}

func build_payload_request(description, destination_branch, title string, reviewers []map[string]map[string]string) []byte {

	data := map[string]interface{}{
		"description": description,
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

func publish_pr_request(url string, json_payload []byte) bool {

	log.Println("Publishing to ", url)
	json_data, err := json.Marshal(json_payload)
	if err != nil {
		log.Fatal("Could not marshal json_payload")
	}

	client := http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewReader(json_data))

	req.Header = http.Header{
		"content-type":  {"application/json"},
		"authorization": {fmt.Sprintf("Bearer %s", get_token())},
	}
	fmt.Println("HEADER :", req.Header)
	res, err := client.Do(req)
	if err != nil {
		log.Fatal("Could not publish PR ...", err)
	}

	defer res.Body.Close()
	fmt.Println(res.Status)

	return true
}
