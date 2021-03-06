// Copyright (c) 2021 Micah Rosales. MIT License.

// Workflow to look up emojis by keyword.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	aw "github.com/deanishe/awgo"
	"github.com/deanishe/awgo/update"
	"github.com/mrosales/emoji-go"
)

const (
	// Name of the background job that checks for updates
	updateJobName     = "checkForUpdate"
	skinToneVarName   = "skin_tone"
	updateRepoVarName = "update_repository"
	imageDirectory    = "images"
)

var (
	// Command-line arguments
	doCheck bool

	// Icon to show if an update is available
	iconAvailable = &aw.Icon{Value: "update-available.png"}
)

var (
	autoUpdateRepository = os.Getenv(updateRepoVarName)
)

func main() {
	flag.BoolVar(&doCheck, "check", false, "check for a new version")

	var options []aw.Option
	if len(autoUpdateRepository) > 0 {
		options = append(options, update.GitHub(autoUpdateRepository))
	}

	wf := aw.New(options...)
	wf.Run(func() {
		// make sure alfred can intercept magic commands
		wf.Args()
		flag.Parse()
		runWorkflow(wf, flag.Arg(0))
	})
}

func runWorkflow(wf *aw.Workflow, query string) {
	// Alternate action: Get available releases from remote.
	if doCheck {
		wf.Configure(aw.TextErrors(true))
		log.Println("Checking for updates...")
		if err := wf.CheckForUpdate(); err != nil {
			wf.FatalError(err)
		}
		return
	}

	// Call self with "check" command if an update is due and a check
	// job isn't already running.
	if wf.UpdateCheckDue() && !wf.IsRunning(updateJobName) {
		log.Println("Running update check in background...")

		cmd := exec.Command(os.Args[0], "-check")
		if err := wf.RunInBackground(updateJobName, cmd); err != nil {
			log.Printf("Error starting update check: %s", err)
		}
	}

	// Only show update status if query is empty.
	if query == "" && wf.UpdateAvailable() {
		// Turn off UIDs to force this item to the top.
		// If UIDs are enabled, Alfred will apply its "knowledge"
		// to order the results based on your past usage.
		wf.Configure(aw.SuppressUIDs(true))

		// Notify user of update. As this item is invalid (Valid(false)),
		// actioning it expands the query to the Autocomplete value.
		// "workflow:update" triggers the updater Magic Action that
		// is automatically registered when you configure Workflow with
		// an Updater.
		//
		// If executed, the Magic Action downloads the latest version
		// of the workflow and asks Alfred to install it.
		wf.NewItem("Update available!").
			Subtitle("↩ to install").
			Autocomplete("workflow:update").
			Valid(false).
			Icon(iconAvailable)
	}

	skinToneStr := os.Getenv(skinToneVarName)
	skinTone, err := emoji.NewModifier(skinToneStr)
	if err != nil {
		wf.FatalError(fmt.Errorf("invalid skin_tone \"%s\"", skinToneStr))
		return
	}

	searcher := emoji.NewSearchIndex(
		emoji.WithLimit(0),
		emoji.WithMaxDistance(10),
	)

	results := searcher.Search(query)
	for _, info := range results {
		image := info.ImageForModifier(skinTone)
		icon := filepath.Join(imageDirectory, fmt.Sprintf("%s.png", image.Unified))
		item := wf.NewItem(info.Name).
			Arg(image.Character).
			Subtitle(fmt.Sprintf("Paste symbol \"%s\" in frontmost app", image.Character)).
			UID(info.Name).
			Valid(true).
			Icon(&aw.Icon{Value: icon, Type: aw.IconTypeImage}).
			Var("action", "paste")
		item.Cmd().
			Arg(image.Character).
			Subtitle(fmt.Sprintf("Copy symbol \"%s\" to the clipboard", image.Character)).
			Var("action", "copy")
		item.Alt().
			Arg(image.Character).
			Subtitle(fmt.Sprintf("Copy code \":%s:\" to the clipboard", info.Name)).
			Var("action", "copy")
	}

	wf.WarnEmpty("No matching items", "Try a different query?")
	wf.SendFeedback()
}
