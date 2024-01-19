package cmd

import (
	"github.com/spf13/cobra"
)

// adminCmd represents the admin command
var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Manage the Admin users",
	Long:  `Manage the Admin users.`,
}

func init() {
	rootCmd.AddCommand(adminCmd)
}
