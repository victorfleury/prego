/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>

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

import "github.com/victorfleury/prego/cmd"

//func main() {

//config := parse_config()
//repo, err := get_repo()
//if err != nil {
//log.Fatal("Prego needs to be run in a Git repository !")
//}

//// Branches
//current_branch, _ := repo.Head()
//branches, err := repo.Branches()
//if err != nil {
//log.Fatal("No branches found. Are you in a properly initialized repository?")
//}

//var branch_names = []string{"dev", "master"}
//branches.ForEach(func(b *plumbing.Reference) error {
//short_name := strings.Split(b.String(), "refs/heads/")[1]
//if short_name != current_branch.Name().Short() && !check(branch_names, short_name) {
//branch_names = append(branch_names, short_name)
//}
//return nil
//})

//var branch_names_cleaned []string
//for _, b := range branch_names {
//if b != "" {
//branch_names_cleaned = append(branch_names_cleaned, b)
//}
//}

//branch_options := make([]huh.Option[string], len(branch_names_cleaned))
//for i, branch := range branch_names {
//branch = strings.Trim(branch, "* ")
//if branch != "" {
//branch_options[i] = huh.NewOption(branch, branch)
//}
//}

//// Reviewers
//reviewers_option := make([]huh.Option[string], len(config.All_reviewers))
//for i, reviewer := range config.All_reviewers {
//selected := reviewer_in_prefs(config, reviewer)
//reviewers_option[i] = huh.NewOption(reviewer["user"]["name"], reviewer["user"]["name"]).Selected(selected)
//}

//var confirm bool

//form := huh.NewForm(
//huh.NewGroup(
//huh.NewSelect[string]().
//Title("Choose the destination branch :").
//Description("The branch you want to merge your changes to.").
//Options(
//branch_options...,
//).
//Value(&destination_branch),
//),
//huh.NewGroup(
//huh.NewMultiSelect[string]().
//Title("Select reviewers").
//Description("Pick which team members should review your PR").
//Options(reviewers_option...).
//Value(&reviewers),
//),
//huh.NewGroup(
//huh.NewText().
//Value(&PR_TEMPLATE).
//Title("PR Description").
//Editor(config.Editor).
//Lines(15).
//CharLimit(5000).
//Description("Content of the PR"),
//huh.NewConfirm().Title("Publish PR").Affirmative("Yes !").Negative("Cancel").Value(&confirm),
//),
//)

//err = form.Run()

//if err != nil {
//fmt.Println("Uh oh:", err)
//os.Exit(1)
//}

//if !confirm {
//log.Println("Publish PR aborted...")
//os.Exit(0)
//}

// _ = spinner.New().Title("Publishing PR...").Accessible(false).Action(publish_pr).Run()
// }
func main() {
	cmd.Execute()
}
