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

func main() {
	cmd.Execute()
}
