package git

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"log"
	"os"
	"strings"
)

const URL_TEMPLATE string = "https://bitbucket.rodeofx.com/rest/api/1.0/projects/%s/repos/%s/pull-requests"

// Get the git repository from the current working directory
func Get_repo() (*git.Repository, error) {
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

// Get the repository URL
func Get_repo_url() string {
	repo, _ := Get_repo()
	remotes, err := repo.Remotes()
	if err != nil || len(remotes) == 0 {
		log.Fatal("Repository has no remote...")
	}
	remote := remotes[0]
	config_url := remote.Config().URLs[0]

	parts := strings.Split(config_url, "/")
	project, slug_name := parts[len(parts)-2], parts[len(parts)-1]

	formatted_url := fmt.Sprintf(URL_TEMPLATE, project, strings.Split(slug_name, ".git")[0])

	log.Println(formatted_url)

	return formatted_url
}

// Get the last commit message to use as input for the description value of the PR template
func Get_last_commit_message() string {
	repo, _ := Get_repo()

	head, _ := repo.Head()

	last_commit, err := repo.Log(&git.LogOptions{From: head.Hash()})
	if err != nil {
		log.Fatal("Could not retrieve last commit message...")
	}

	last_commit_object, err := last_commit.Next()
	if err != nil {
		log.Fatal("bu")
	}
	return strings.Trim(last_commit_object.Message, "\n")
}
