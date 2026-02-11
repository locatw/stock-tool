package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
)

func NewRootCmd(db *sql.DB, port int) *cobra.Command {
	c := &cobra.Command{
		Use:   "api",
		Short: "stock-tool API server",
		RunE: func(c *cobra.Command, args []string) error {
			return NewAPICommand(db, port).Run(c.Context())
		},
	}

	return c
}

type APICommand struct {
	db   *sql.DB
	port int
}

func NewAPICommand(db *sql.DB, port int) *APICommand {
	return &APICommand{db: db, port: port}
}

func (c *APICommand) Run(ctx context.Context) error {
	e := c.setupRouter()

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := e.Start(fmt.Sprintf(":%d", c.port)); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal(err)
		}
	}()

	<-ctx.Done()

	if err := e.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return nil
}

func (c *APICommand) setupRouter() *echo.Echo {
	e := echo.New()
	e.GET("/health", c.healthCheck)
	return e
}

func (c *APICommand) healthCheck(ctx echo.Context) error {
	if err := c.db.PingContext(ctx.Request().Context()); err != nil {
		return ctx.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "unhealthy",
			"error":  err.Error(),
		})
	}
	return ctx.JSON(http.StatusOK, map[string]string{
		"status": "healthy",
	})
}
