package cmd

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/pflag"

	//"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	prego_git "github.com/victorfleury/prego/internal/git"
	"github.com/victorfleury/prego/internal/utils"
)

var (
	destination_branch string
	title              string
	url                string
	reviewers          []string
)

var PR_TEMPLATE string = `#### Purpose of the PR
{DESCRIPTION}
#### Type of feedback wanted

#### Potential risks of this change

#### Relationship with other PRs
`

func root_prego(empty_reviewers pflag.Value) {
	config := parse_config()
	repo, err := prego_git.Get_repo()
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
		if short_name != current_branch.Name().Short() && !utils.IsNameInNames(branch_names, short_name) {
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
	reviewers_option := make([]huh.Option[string], len(utils.Default_config_payload().All_reviewers))
	if config.All_reviewers == nil {
	}
	for i, reviewer := range utils.Default_config_payload().All_reviewers {
		//for i, reviewer := range config.All_reviewers {
		var selected bool
		if empty_reviewers.String() == "true" {
			selected = false
		} else {
			selected = utils.Reviewer_in_prefs(config, reviewer)
		}
		reviewers_option[i] = huh.NewOption(reviewer["user"]["name"], reviewer["user"]["name"]).Selected(selected)
	}

	// Update template with last message commit :
	replacement_string := []string{"{DESCRIPTION}", prego_git.Get_last_commit_message()}
	replacer := strings.NewReplacer(replacement_string...)
	UPDATED_PR_TEMPLATE := replacer.Replace(PR_TEMPLATE)

	var confirm bool
	editor := "vim"
	if os.Getenv("EDITOR") != "" {
		editor = os.Getenv("EDITOR")
	}
	if config.Editor != "" {
		editor = config.Editor
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
				Title("PR Description").
				Editor(editor).
				Lines(25).
				CharLimit(500000).
				Description("Content of the PR").
				Value(&UPDATED_PR_TEMPLATE),
			huh.NewConfirm().Title("Publish PR").Affirmative("Yes !").Negative("Cancel").Value(&confirm),
		),
	)

	err = form.Run()

	if err != nil {
		log.Println("Uh oh:", err)
		os.Exit(1)
	}

	if !confirm {
		log.Println("Publish PR aborted...")
		os.Exit(0)
	}

	accessible := true
	if os.Getenv("PREGO_ACCESSIBLE") != "" {
		accessible = false
	}
	_ = spinner.New().Title("Publishing PR...").Accessible(accessible).Action(publish_pr).Run()
}

func publish_pr() {
	log.Println("Publishing PR")
	repo_url := prego_git.Get_repo_url()

	reviewers_payload_data := utils.Build_reviewers_payload_data(reviewers)

	repo, _ := prego_git.Get_repo()
	head_ref, _ := repo.Head()
	commit_message, _ := repo.CommitObject(head_ref.Hash())
	title := strings.Split(commit_message.Message, "\n")[0]

	json_payload := utils.Build_payload_request(
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
		log.Fatal("Could not publish PR ...\n")
	}
}

// Perform the HTTP Request to the Bitbucket REST API
func publish_pr_request(url string, json_payload []byte) bool {

	log.Println("Publishing to ", url)

	client := http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewReader(json_payload))

	req.Header = http.Header{
		"content-type":  {"application/json"},
		"authorization": {fmt.Sprintf("Bearer %s", utils.Get_token())},
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
