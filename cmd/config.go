package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

type ConfigPayload struct {
	Editor        string
	All_reviewers []map[string]map[string]string
	My_reviewers  []map[string]map[string]string
}

var default_config string = `
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
		config = []byte(default_config)
	}

	var config_payload ConfigPayload

	err = json.Unmarshal(config, &config_payload)
	if err != nil {
		log.Fatal("Something is wrong with the configuration. It could not be parsed fully.")
	}

	return config_payload
}

var config_command = &cobra.Command{
	Use:   "config",
	Short: "Wizard to setup your config",
	Run: func(cmd *cobra.Command, args []string) {
		config_wizard()
	},
}

func init() {
	rootCmd.AddCommand(config_command)
}

func config_wizard() {
	var editor string
	// Reviewers
	var default_config_payload ConfigPayload

	err := json.Unmarshal([]byte(default_config), &default_config_payload)
	reviewers_option := make([]huh.Option[string], len(default_config_payload.All_reviewers))
	if err != nil {
		log.Fatal("Could not read default config from code...")
	}
	for i, reviewer := range default_config_payload.All_reviewers {
		selected := reviewer_in_prefs(default_config_payload, reviewer)
		reviewers_option[i] = huh.NewOption(reviewer["user"]["name"], reviewer["user"]["name"]).Selected(selected)
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select your preferred editor : ").
				Options(
					huh.NewOption("vim", "vim").Selected(true),
					huh.NewOption("nano", "nano"),
				).
				Value(&editor),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select your preferred reviewers: ").
				Options(reviewers_option...).
				Value(&reviewers),
		),
	)

	err = form.Run()
	if err != nil {
		fmt.Println("Uh oh:", err)
		os.Exit(1)
	}

	fmt.Println("Selected editor :", editor)

	// Writing out the json file
	var custom_config_payload ConfigPayload
	custom_config_payload.Editor = editor
	var reviewers_payload []map[string]map[string]string
	for _, r := range reviewers {
		payload := map[string]map[string]string{"user": {"name": r}}
		reviewers_payload = append(reviewers_payload, payload)
	}
	custom_config_payload.My_reviewers = reviewers_payload

	root_path_to_config := os.Getenv("XDG_CONFIG_HOME")
	if root_path_to_config == "" {
		root_path_to_config = os.Getenv("HOME") + "/.config"
	}

	fmt.Println(custom_config_payload)

	path_to_config := root_path_to_config + "/prego/prego.json"
	log.Println("Path to config", path_to_config)

	data, err := json.Marshal(custom_config_payload)
	err = os.WriteFile(path_to_config, data, 644)
	if err != nil {
		log.Fatal("Could not encore custom config", err)
	}

}