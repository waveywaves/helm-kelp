package cmd

import (
	"fmt"

	"sigs.k8s.io/kustomize/pkg/commands/build"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "kust",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
 examples and usage of using your application. For example:
 
 Cobra is a CLI library for Go that empowers applications.
 This application is a tool to generate the needed files
 to quickly create a Cobra application.`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello Kust !")
		build.NewOptions("", "")
	},
}
