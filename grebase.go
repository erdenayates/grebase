package main

import (
	"bufio"
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
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	output, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(output)) == "true"
}

func promptForInput(prompt string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

func backupBranch(branchName string) {
	counter := 1
	backupName := fmt.Sprintf("%s-backup-%d", branchName, counter)

	// Check if the backup branch name already exists
	for branchExists(backupName) {
		counter++
		backupName = fmt.Sprintf("%s-backup-%d", branchName, counter)
	}
	cmd := exec.Command("git", "branch", backupName)
	executeCommand(cmd)
}

func branchExists(branchName string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", branchName)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func gitAdd(files []string) {
	cmdArgs := append([]string{"add"}, files...)
	cmd := exec.Command("git", cmdArgs...)
	executeCommand(cmd)
}

func gitCommit(commitMsg string) {
	cmd := exec.Command("git", "commit", "-m", commitMsg)
	executeCommand(cmd)
}

func gitPush(branch string) {
	cmd := exec.Command("git", "push", "--set-upstream", "origin", branch)
	executeCommand(cmd)
}

func gitCheckout(branch string) {
	cmd := exec.Command("git", "checkout", branch)
	executeCommand(cmd)
}

func gitRebase(branch string) {
	cmd := exec.Command("git", "rebase", branch)
	executeCommand(cmd)
}

func main() {
	featureBranch := flag.String("feature-branch", "", "Name of the feature branch")
	targetBranch := flag.String("target-branch", "master", "Name of the target branch")
	commitMsg := flag.String("commit", "", "Commit message")
	addFiles := flag.String("add-file", ".", "Specify files or directories to add to git separated by spaces")
	backupFeatureBranch := flag.Bool("backup-feature-branch", false, "Create a backup of the feature branch before rebasing")
	backupTargetBranch := flag.Bool("backup-target-branch", false, "Create a backup of the target branch before rebasing")
	interactive := flag.Bool("interactive", false, "Run the tool in interactive mode")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Println("This tool is designed to automate the git rebase strategy for updating git repositories.")
		fmt.Println("Example: grebase --feature-branch=my-feature --commit=\"My commit message\" --add-file=\"jenkins/plugins jenkins/testfile\"")
		fmt.Println("Interactive Mode: grebase --interactive")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NFlag() == 0 && !*interactive {
		flag.Usage()
		return
	}

	if *interactive {
		var err error
		*featureBranch, err = promptForInput("Enter the name of the feature branch: ")
		if err != nil {
			log.Fatalf("Error reading input: %v", err)
		}

		*targetBranch, err = promptForInput("Enter the name of the target branch (default is master): ")
		if err != nil {
			log.Fatalf("Error reading input: %v", err)
		}
		if *targetBranch == "" {
			*targetBranch = "master"
		}
		
		*commitMsg, err = promptForInput("Enter the commit message: ")
		if err != nil {
			log.Fatalf("Error reading input: %v", err)
		}

		files, err := promptForInput("Specify files or directories to add to git (separated by spaces, default is '.'): ")
		if err != nil {
			log.Fatalf("Error reading input: %v", err)
		}
		if files == "" {
			files = "."
		}
		*addFiles = files
		
		backupFeature, err := promptForInput("Do you want to create a backup of the feature branch before rebasing? (yes/no): ")
		if err != nil {
			log.Fatalf("Error reading input: %v", err)
		}
		if backupFeature == "yes" {
			*backupFeatureBranch = true
		}

		backupTarget, err := promptForInput("Do you want to create a backup of the target branch before rebasing? (yes/no): ")
		if err != nil {
			log.Fatalf("Error reading input: %v", err)
		}
		if backupTarget == "yes" {
			*backupTargetBranch = true
		}
	}

	if !isGitRepository() {
		log.Fatal("Current directory is not a git repository.")
	}

	if *backupFeatureBranch {
		backupBranch(*featureBranch)
	}

	if *backupTargetBranch {
		backupBranch(*targetBranch)
	}

	filesToAdd := strings.Split(*addFiles, " ")
	gitAdd(filesToAdd)
	gitCommit(*commitMsg)
	gitPush(*featureBranch)
	gitCheckout(*targetBranch)
	gitRebase(*featureBranch)
	gitPush(*targetBranch)

	fmt.Println("Process completed!")
}

