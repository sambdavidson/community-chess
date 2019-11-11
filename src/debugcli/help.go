package main

import (
	"fmt"
	"sort"
)

func init() {
	commands["help"] = command{
		helpText: "Display set of all commands.",
		action:   help,
	}
}

func help(args actionArgs) {
	fmt.Println("Commands: ")
	var keys []string
	for k := range commands {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		cmd := commands[k]
		fmt.Printf("- '%s': %s\n", k, cmd.helpText)
	}
}
