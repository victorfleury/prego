package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "prego",
	Short: "Create your PR from the CLI",
	Long: `A tool to quickly create a PR from the CLI to Bitbucket. All it needs is a token from
Bitbucket to ensure authentication to the REST API.
You can create a config file at ${XDG_CONFIG_HOME:HOME}/prego/prego.json`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		root_prego()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// TODO: How do you pass the value of the flag to the CMD ?
	rootCmd.Flags().BoolP("empty-reviewers", "e", false, "Unselect all reviewers. Useful to create an empty PR")
}
