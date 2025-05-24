package cmd

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "task",
		Short: "stock-tool task",
		Run: func(c *cobra.Command, args []string) {
			c.Help()
		},
	}

	return c
}
