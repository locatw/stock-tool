package cmd

import (
	"fmt"
	"time"

	"github.com/samber/do"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"stock-tool/internal/api/jquants"
	"stock-tool/internal/infra/repository"
	"stock-tool/internal/infra/storage"
	usecase "stock-tool/internal/usecase/task"
)

func newExtractCmd(injector *do.Injector) *cobra.Command {
	c := &cobra.Command{
		Use:   "extract",
		Short: "extract data from a source",
		Run: func(c *cobra.Command, args []string) {
			c.Help()
		},
	}

	c.AddCommand(newExtractJQuantsCmd(injector))

	return c
}

func newExtractJQuantsCmd(injector *do.Injector) *cobra.Command {
	c := &cobra.Command{
		Use:   "jquants",
		Short: "extract data from J-Quants source",
		RunE: func(c *cobra.Command, args []string) error {
			return newExtractCommand(c, injector).Execute()
		},
	}

	c.Flags().String("type", "", "type of data to extract from the source")
	c.Flags().String("code", "", "code of the listed issue to extract (optional)")
	c.Flags().String("start-date", "", "start date for extracting data (optional)")
	c.Flags().String("end-date", "", "end date for extracting data (optional)")
	c.MarkFlagRequired("type")

	return c
}

type extractJQuantsCommand struct {
	cmd      *cobra.Command
	injector *do.Injector
}

func newExtractCommand(cmd *cobra.Command, injector *do.Injector) *extractJQuantsCommand {
	return &extractJQuantsCommand{cmd: cmd, injector: injector}
}

func (c *extractJQuantsCommand) Execute() error {
	dataType, err := c.cmd.Flags().GetString("type")
	if err != nil {
		return err
	}

	code, err := c.getOptionStringFlag("code")
	if err != nil {
		return err
	}

	startDate, err := c.getOptionDateFlag("start-date")
	if err != nil {
		return err
	}

	endDate, err := c.getOptionDateFlag("end-date")
	if err != nil {
		return err
	}

	brandFetcher := do.MustInvoke[*jquants.BrandFetcher](c.injector)
	objectWriter := do.MustInvoke[*storage.S3Client](c.injector)
	extractTaskRepo := do.MustInvoke[*repository.ExtractTaskRepository](c.injector)

	req := &usecase.ExtractTaskRequest{
		Source:    "jquants",
		DataType:  dataType,
		Timing:    "daily",
		Code:      code,
		StartDate: startDate,
		EndDate:   endDate,
	}

	uc := usecase.NewExtractTaskUseCase(brandFetcher, objectWriter, extractTaskRepo)
	resp, err := uc.Extract(c.cmd.Context(), req)
	if err != nil {
		return err
	}

	fmt.Printf("Extract completed: status=%s, s3Key=%s\n", resp.Status, resp.S3Key)
	return nil
}

func (c *extractJQuantsCommand) getOptionStringFlag(flag string) (*string, error) {
	if !c.cmd.Flags().Changed(flag) {
		return nil, nil
	}

	s, err := c.cmd.Flags().GetString(flag)
	if err != nil {
		return nil, err
	} else if s == "" {
		return nil, nil
	}

	return lo.ToPtr(s), nil
}

func (c *extractJQuantsCommand) getOptionDateFlag(flag string) (*time.Time, error) {
	if !c.cmd.Flags().Changed(flag) {
		return nil, nil
	}

	dateStr, err := c.cmd.Flags().GetString(flag)
	if err != nil {
		return nil, err
	} else if dateStr == "" {
		return nil, nil
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %v", err)
	}

	return lo.ToPtr(date), nil
}
