package cmd

import (
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/octoboy233/kubectl-ai/pkg/client"
	"github.com/octoboy233/kubectl-ai/pkg/helper"
	"github.com/octoboy233/kubectl-ai/pkg/template"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"time"
)

func GetDiagnoseCommand(opt *helper.Options) *cobra.Command {
	cf := genericclioptions.NewConfigFlags(true)
	cmd := &cobra.Command{
		Use:   "diagnose TYPE NAME",
		Short: "Diagnose a resource",
		Args:  cobra.ExactArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opt.Token == "" && (opt.AppID == "" || opt.APISecret == "" || opt.APIKey == "") {
				return fmt.Errorf("please specify the token for %s with ENV %s", opt.Typ, helper.EnvToken)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := helper.GetResourceYaml(cf, args[0], args[1])
			if err != nil {
				return err
			}

			data := template.NewData(string(b), opt.Lang)
			text, err := data.Parse(template.DiagnoseTpl)
			if err != nil {
				return err
			}

			s := spinner.New(spinner.CharSets[39], 100*time.Millisecond)
			s.Suffix = "Diagnosing..."

			s.Start()

			err = client.CreateCompletion(cmd.Context(), opt, string(text), cmd.OutOrStderr(), s)

			return err
		},
	}

	cmd.PersistentFlags().StringVar(cf.KubeConfig, "kubeconfig", "", "path to the kubeconfig file")
	cmd.PersistentFlags().StringVar(cf.Namespace, "namespace", "default", "namespace of the pod")

	return cmd
}
