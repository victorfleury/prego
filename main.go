package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
)

var (
	destination_branch string
	//reviewers          []string
)

var reviewers = []string{"bramoul", "jdubuisson", "bob", "alice"}

var pr_description string = `#### Purpose of the PR

#### Overview of the changes

#### Type of feedback wanted

#### Where should the reviewer start looking at?

#### Potential risks of this change

#### Relationship with other PRs
`

func main() {

	// Branches
	cmd := exec.Command("git", "branch", "--all")
	branches, err := cmd.Output()
	if err != nil {
		log.Fatal("Not in a git repo\n", err)
	}

	branch_names := strings.Split(string(branches), "\n")
	var branch_names_cleaned []string
	for _, b := range branch_names {
		if b != "" {
			branch_names_cleaned = append(branch_names_cleaned, b)
		}
	}

	branches_option := make([]huh.Option[string], len(branch_names_cleaned))
	for i, branch := range branch_names {
		branch = strings.Trim(branch, "* ")
		if branch != "" {
			branches_option[i] = huh.NewOption(branch, branch)
		}
	}

	// Reviewers
	reviewers_option := make([]huh.Option[string], len(reviewers))
	for i, reviewer := range reviewers {
		reviewers_option[i] = huh.NewOption(reviewer, reviewer)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose the destination branch :").
				Options(
					branches_option...,
				).
				Value(&destination_branch),
			huh.NewMultiSelect[string]().
				Value(&reviewers).
				Title("Select reviewers").
				Options(reviewers_option...),
			huh.NewText().
				Value(&pr_description).
				Title("PR Description").
				Lines(10).
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
	time.Sleep(2 * time.Second)
	fmt.Println("Published PR successfully !")
}
