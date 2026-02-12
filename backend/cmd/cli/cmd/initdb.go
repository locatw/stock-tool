package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"stock-tool/database"
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
	defer func() {
		if err := db.Shutdown(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to shutdown database: %v\n", err)
		}
	}()

	if err := db.Init(); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "database initialized successfully.")

	return nil
}
