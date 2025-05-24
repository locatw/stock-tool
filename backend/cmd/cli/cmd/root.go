package cmd

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "cli",
		Short: "stock-tool command line tool",
		Run: func(c *cobra.Command, args []string) {
			c.Help()
		},
	}

	c.AddCommand(newInitDBCmd())
	c.AddCommand(newMigrateCmd())

	return c
}
