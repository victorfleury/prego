/*
* TODO :
- Fetch all branches from the repo. -> OK
  - Add dev and main first, then the rest of the branches

- Default reviewers :
  - Use a map for reviewers -> OK
  - Read from a JSON config file
  - Add cli arg to remove them altogether

- Use Go Git package -> OK
- Read Token from file -> OK
- Build payload for POST request
- Execute payload successfully
*
*/
package main

import (
	"fmt"
	"log"
	"os"
	//"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

const URL_TEMPLATE string = "https://bitbucket.rodeofx.com/rest/api/1.0/projects/%s/repos/%s/pull-requests"

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
	var destination_branch string
	//cmd := exec.Command("git", "branch", "--all")
	//branches, err := cmd.Output()

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
	var reviewers []string
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
				Value(&reviewers).
				Title("Select reviewers").
				Description("Pick which team members should review your PR").
				Options(reviewers_option...),
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

	_ = spinner.New().Title("Publishing PR...").Accessible(true).Action(publish_pr).Run()
}

func publish_pr() {
	time.Sleep(1 * time.Second)

	log.Println("Fetching token")
	token := get_token()
	repo_url, err := get_repo_url()
	if err != nil {
		log.Fatal("Could not retrieve the repo URL")
	}
	log.Println("Repo URL is ", repo_url)
	log.Println("Token is ", token)
	log.Println("Published PR successfully !")

}

func get_token() string {

	token_path := os.Getenv("HOME") + "/token.tk"
	fmt.Println("Token path is ", token_path)
	token, err := os.ReadFile(token_path)
	if err != nil {
		log.Fatal("Panic ! No token found")
	}
	return string(token)
}

func get_repo() (*git.Repository, error) {
	current_directory, err := os.Getwd()
	if err != nil {
		fmt.Println("Current directory could not be found?")
	}
	fmt.Println("Current directory is", current_directory)

	//repo, err := git.PlainOpen(current_directory)
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
	if err != nil {
		log.Fatal("Repository has no remote...")
	}
	remote := remotes[0]
	config_url := remote.Config().URLs[0]

	fmt.Println("---> Config url", config_url)
	parts := strings.Split(config_url, "/")
	project, slug_name := parts[len(parts)-2], parts[len(parts)-1]

	formatted_url := fmt.Sprintf(URL_TEMPLATE, project, slug_name)

	log.Println(formatted_url)

	return formatted_url
}
