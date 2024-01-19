package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// createFgaModelCmd represents the createFgaModel command
var createFgaModelCmd = &cobra.Command{
	Use:   "create-fga-model",
	Short: "Create the openfga model.",
	Long:  `Create the openfga model.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO")
	},
}

func init() {
	rootCmd.AddCommand(createFgaModelCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createFgaModelCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createFgaModelCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
