package main

import (
	"fmt"
	"github.com/octoboy233/kubectl-ai/pkg/cmd"
	"github.com/octoboy233/kubectl-ai/pkg/helper"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func getCommandName() string {
	if strings.HasPrefix(filepath.Base(os.Args[0]), "kubectl-") {
		return "kubectl-pecker"
	} else {
		return "pecker"
	}
}

func main() {
	opt := helper.NewOptions()

	rootCmd := &cobra.Command{
		Use:  getCommandName(),
		Long: opt.GenerateHelpMessage(),
	}

	rootCmd.AddCommand(cmd.GetDiagnoseCommand(opt))

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
