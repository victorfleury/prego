package cmd

import (
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

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

func root_prego() {
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
		log.Println("Uh oh:", err)
		os.Exit(1)
	}

	if !confirm {
		log.Println("Publish PR aborted...")
		os.Exit(0)
	}

	_ = spinner.New().Title("Publishing PR...").Accessible(false).Action(publish_pr).Run()
}

func publish_pr() {
	log.Println("PUBLISHING PR !!!")
}
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

// Check if a given reviewer is in the config for default reviewers
func reviewer_in_prefs(config ConfigPayload, reviewer map[string]map[string]string) bool {
	for _, r := range config.My_reviewers {
		if reflect.DeepEqual(r, reviewer) {
			return true
		}
	}
	return false
}

func check(branches_name []string, name string) bool {
	for _, b := range branches_name {
		if name == b {
			return true
		}
	}
	return false
}
