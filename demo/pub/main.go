package main

import (
	"flag"
	"fmt"
	"os"

	prompt "github.com/c-bata/go-prompt"
	"github.com/chumvan/sipRestServer/src/pub"
)

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "register", Description: "Register to a SIP Server"},
		{Text: "create topic", Description: "Create a conference at the SIP Server to serve as a topic"},
		{Text: "get topic", Description: "Get the topic's name"},
		{Text: "publish data", Description: "Start sending data to a topic"},
		{Text: "stop publish data", Description: "Stop the data publishing process"},
		{Text: "leave topic", Description: "Remove itself from the topic's publisher list"},
		{Text: "terminate topic", Description: "Destroy the topic by itself"},
		{Text: "exit", Description: "Exit"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func usage() {
	fmt.Fprintf(os.Stderr, `go Publisher
Usage: publisher [-nc]

Options:
`)
	flag.PrintDefaults()
}

func consoleLoop(p *pub.Publisher) {
	fmt.Println("please select a command")
	for {
		t := prompt.Input("CLI> ", completer,
			prompt.OptionTitle("GO Publisher 1.0.0"),
			prompt.OptionPrefixTextColor(prompt.Yellow),
			prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
			prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
			prompt.OptionSuggestionBGColor(prompt.DarkGray))

		switch t {
		case "register":
			// Add REGISTER logic to this command line
		case "exit":
			fmt.Println("exit now")
			p.SIP.UA.Shutdown()
			return
		}
	}
}

func main() {
	p := pub.NewPublisher
}
