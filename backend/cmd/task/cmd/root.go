package cmd

import (
	"github.com/samber/do"
	"github.com/spf13/cobra"
)

func NewRootCmd(injector *do.Injector) *cobra.Command {
	c := &cobra.Command{
		Use:   "task",
		Short: "stock-tool task",
		Run: func(c *cobra.Command, args []string) {
			c.Help()
		},
	}

	c.AddCommand(newFetchDataCmd(injector))

	return c
}
