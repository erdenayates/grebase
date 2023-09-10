package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func executeCommand(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}

func isGitRepository() bool {
	_, err := os.Stat(".git")
	return !os.IsNotExist(err)
}

func main() {
	featureBranch := flag.String("feature-branch", "", "Name of the feature branch")
	targetBranch := flag.String("target-branch", "master", "Name of the target branch")
	commitMsg := flag.String("commit", "", "Commit message")
	addFiles := flag.String("add-file", ".", "Specify files or directories to add to git separated by spaces")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Println("This tool is designed to automate the git rebase strategy for updating git repositories.")
		fmt.Println("Example: grebase --feature-branch=my-feature --commit=\"My commit message\" --add-file=\"jenkins/plugins jenkins/testfile\"")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NFlag() == 0 {
		flag.Usage()
		return
	}

	if !isGitRepository() {
		log.Fatal("Current directory is not a git repository.")
	}

	if *featureBranch == "" || *commitMsg == "" {
		log.Fatal("feature-branch and commit are required flags.")
	}

	filesToAdd := strings.Split(*addFiles, " ")
	cmdArgs := append([]string{"add"}, filesToAdd...)
	cmd := exec.Command("git", cmdArgs...)
	executeCommand(cmd)

	cmd = exec.Command("git", "commit", "-m", *commitMsg)
	executeCommand(cmd)

	cmd = exec.Command("git", "push", "--set-upstream", "origin", *featureBranch)
	executeCommand(cmd)

	cmd = exec.Command("git", "checkout", *targetBranch)
	executeCommand(cmd)

	cmd = exec.Command("git", "rebase", *featureBranch)
	executeCommand(cmd)

	cmd = exec.Command("git", "push", "origin", *targetBranch)
	executeCommand(cmd)

	fmt.Println("Process completed!")
}

