package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/victorfleury/prego/internal/utils"
)

//type ConfigPayload struct {
//Editor        string
//All_reviewers []map[string]map[string]string
//My_reviewers  []map[string]map[string]string
//}

// Parse the JSON config for prego
func parse_config() utils.ConfigPayload {

	root_path_to_config := os.Getenv("XDG_CONFIG_HOME")
	if root_path_to_config == "" {
		root_path_to_config = os.Getenv("HOME") + "/.config"
	}

	path_to_config := root_path_to_config + "/prego/prego.json"
	log.Println("Path to config", path_to_config)

	config, err := os.ReadFile(path_to_config)
	if err != nil {
		log.Println("Could not read the config file at ", path_to_config, "Using default")
		config = []byte(utils.Default_config)
	}

	var config_payload utils.ConfigPayload

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
	// Get the existing configuration if any
	path_to_config := get_config_path()
	existing_config := get_existing_config(path_to_config)

	// Editor
	var editor string
	// Reviewers
	var default_config_payload utils.ConfigPayload

	err := json.Unmarshal([]byte(utils.Default_config), &default_config_payload)
	reviewers_option := make([]huh.Option[string], len(default_config_payload.All_reviewers))
	if err != nil {
		log.Fatal("Could not read default config from code...")
	}
	for i, reviewer := range default_config_payload.All_reviewers {
		selected := utils.Reviewer_in_prefs(default_config_payload, reviewer) || utils.Reviewer_in_prefs(existing_config, reviewer)
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

	// Writing out the json file
	var custom_config_payload utils.ConfigPayload
	custom_config_payload.Editor = editor
	var reviewers_payload []map[string]map[string]string
	for _, r := range reviewers {
		payload := map[string]map[string]string{"user": {"name": r}}
		reviewers_payload = append(reviewers_payload, payload)
	}
	custom_config_payload.My_reviewers = reviewers_payload

	data, err := json.Marshal(custom_config_payload)
	err = os.WriteFile(path_to_config, data, 0644)
	if err != nil {
		log.Fatal("Could not encore custom config", err)
	}
	log.Println("Configuration successfully saved at ", path_to_config)

}

// Read the existing config if it exists
func get_existing_config(path string) utils.ConfigPayload {

	var existing_config utils.ConfigPayload
	data, err := os.ReadFile(path)
	if err != nil {
		log.Println("Could not read the config file from ", path)
		return existing_config
	}

	err = json.Unmarshal(data, &existing_config)
	if err != nil {
		log.Fatal("Could not read existing config.... Is it properly formatted?")
	}
	return existing_config
}

// Get the config path.
func get_config_path() string {
	root_path_to_config := os.Getenv("XDG_CONFIG_HOME")
	if root_path_to_config == "" {
		root_path_to_config = os.Getenv("HOME") + "/.config"
	}

	// makedirs
	os.Mkdir(root_path_to_config+"/prego", 0755)

	path_to_config := root_path_to_config + "/prego/prego.json"
	//log.Println("Path to config", path_to_config)
	return path_to_config
}
