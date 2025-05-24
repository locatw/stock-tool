package cmd

import (
	"errors"
	"fmt"
	"stock-tool/database"

	"github.com/spf13/cobra"
)

func newInitDBCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init-db",
		Short: "initialize the database",
		RunE: func(c *cobra.Command, args []string) error {
			return newInitDBCommand().Execute(c, args)
		},
	}
}

type initDBCommand struct{}

func newInitDBCommand() *initDBCommand {
	return &initDBCommand{}
}

func (c *initDBCommand) Execute(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	config := ctx.Value(database.CTXKeyDBConfig)
	if config == nil {
		return errors.New("database config is nil")
	}

	db := database.NewRawDB(config.(database.Config))
	if err := db.Connect(); err != nil {
		return err
	}
	defer db.Shutdown()

	if err := db.Init(); err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "database initialized successfully.")

	return nil
}
