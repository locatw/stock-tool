package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"stock-tool/database"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
)

const (
	migrationDir = "migrations"
)

func newMigrateCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "migrate",
		Short: "database migration commands",
		Run: func(c *cobra.Command, args []string) {
			c.Help()
		},
	}

	c.AddCommand(newMigrateCmdCreate())
	c.AddCommand(newMigrateCmdList())
	c.AddCommand(newMigrateCmdUp())
	c.AddCommand(newMigrateCmdDown())
	c.AddCommand(newMigrateCmdGoto())
	c.AddCommand(newMigrateCmdVersion())

	return c
}

func newMigrateCmdCreate() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "create a new migration file",
		RunE: func(c *cobra.Command, args []string) error {
			return newMigrateCommand(migrationDir).Create(c, args)
		},
	}
}

func newMigrateCmdList() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list migration versions",
		RunE: func(c *cobra.Command, args []string) error {
			return newMigrateCommand(migrationDir).List(c)
		},
	}
}

func newMigrateCmdUp() *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "migrate up to latest version",
		RunE: func(c *cobra.Command, _ []string) error {
			return newMigrateCommand(migrationDir).Up(c)
		},
	}
}

func newMigrateCmdDown() *cobra.Command {
	return &cobra.Command{
		Use:   "down",
		Short: "migrate down all",
		RunE: func(c *cobra.Command, _ []string) error {
			return newMigrateCommand(migrationDir).Down(c)
		},
	}
}

func newMigrateCmdGoto() *cobra.Command {
	return &cobra.Command{
		Use:   "goto",
		Short: "migrate to a specific version",
		RunE: func(c *cobra.Command, args []string) error {
			return newMigrateCommand(migrationDir).Goto(c, args)
		},
	}
}

func newMigrateCmdVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "show current migration version",
		RunE: func(c *cobra.Command, _ []string) error {
			return newMigrateCommand(migrationDir).Version(c)
		},
	}
}

type migrateCommand struct {
	migrationDir string
}

func newMigrateCommand(migrationDir string) *migrateCommand {
	return &migrateCommand{migrationDir: migrationDir}
}

func (m *migrateCommand) Create(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("migration name is required")
	}

	command := exec.Command("go", "tool", "migrate", "create", "-ext", "sql", "-dir", migrationDir, args[0])
	command.Stdout = cmd.OutOrStdout()
	command.Stderr = cmd.OutOrStderr()

	return command.Run()
}

func (m *migrateCommand) List(cmd *cobra.Command) error {
	entries, err := os.ReadDir(m.migrationDir)
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		baseName := strings.TrimSuffix(name, ".up.sql")
		parts := strings.Split(baseName, "_")
		if len(parts) < 2 {
			continue
		}

		version := parts[0]
		desc := strings.Join(parts[1:], " ")

		fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", version, desc)
	}

	return nil
}

func (m *migrateCommand) Up(cmd *cobra.Command) error {
	ctx := cmd.Context()
	mig, err := m.makeMigrationInstance(ctx)
	if err != nil {
		return err
	}

	return mig.Up()
}

func (m *migrateCommand) Down(cmd *cobra.Command) error {
	ctx := cmd.Context()
	mig, err := m.makeMigrationInstance(ctx)
	if err != nil {
		return err
	}

	if !m.askConfirmation(cmd, "Are you sure you want to apply all down migrations?") {
		fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
		return nil
	}

	return mig.Down()
}

func (m *migrateCommand) Goto(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	mig, err := m.makeMigrationInstance(ctx)
	if err != nil {
		return err
	}

	version, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}

	cur, _, err := mig.Version()
	if err != nil {
		return err
	}

	if uint(version) < cur {
		if !m.askConfirmation(cmd, "Are you sure you want to apply down migrations?") {
			fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
			return nil
		}
	}

	if !m.askConfirmation(cmd, "Are you sure you want to apply all down migrations?") {
		fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
		return nil
	}

	return mig.Migrate(uint(version))
}

func (m *migrateCommand) Version(cmd *cobra.Command) error {
	ctx := cmd.Context()
	mig, err := m.makeMigrationInstance(ctx)
	if err != nil {
		return err
	}

	version, _, err := mig.Version()
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Current migration version: %d\n", version)

	return nil
}

func (m *migrateCommand) makeMigrationInstance(ctx context.Context) (*migrate.Migrate, error) {
	config := ctx.Value(database.CTXKeyDBConfig)
	if config == nil {
		return nil, errors.New("database config is nil")
	}

	db := database.NewRawDB(config.(database.Config))
	if err := db.Connect(); err != nil {
		return nil, err
	}
	defer db.Shutdown()

	driver, err := postgres.WithInstance(db.DB(), &postgres.Config{
		SchemaName: "stock",
	})
	if err != nil {
		return nil, err
	}

	return migrate.NewWithDatabaseInstance("file://"+m.migrationDir, "postgres", driver)
}

func (m *migrateCommand) askConfirmation(cmd *cobra.Command, q string) bool {
	fmt.Fprintf(cmd.OutOrStdout(), "%s (y/n): ", q)

	s := bufio.NewScanner(cmd.InOrStdin())
	s.Scan()
	res := strings.TrimSpace(strings.ToLower(s.Text()))

	return res == "y" || res == "yes"
}
