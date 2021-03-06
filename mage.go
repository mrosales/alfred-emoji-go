//+build mage

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/magefile/mage/sh"
)

const (
	binaryName = "alfred-emoji"
)

// Build builds the go executable for the workflow.
func Build() error {
	return buildPackage(binaryName, ".")
}

// DownloadImages uses the emojigen utility to populate
// the images directory from the CDN spritesheet.
func DownloadImages() error {
	env := map[string]string{
		"GOBIN": getwd(),
	}
	log.Printf("installing emojigen executable")
	if err := sh.RunWithV(
		env,
		"go",
		"install",
		"github.com/mrosales/emoji-go/cmd/emojigen",
	); err != nil {
		return err
	}

	defer func() {
		log.Printf("removing emojigen executable")
		_ = sh.RunV("rm", "-f", "emojigen")
	}()

	log.Printf("running emojigen to download image dataset")
	if err := sh.RunV("./emojigen", "-images", "images"); err != nil {
		return err
	}
	return nil
}

func getwd() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return wd
}

func buildPackage(output, srcPackage string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	return sh.RunV(
		"go", "build",
		fmt.Sprintf("-gcflags=-trimpath=%s", pwd),
		fmt.Sprintf("-asmflags=-trimpath=%s", pwd),
		"-o", output,
		srcPackage,
	)
}
