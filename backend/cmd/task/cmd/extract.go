package cmd

import (
	"fmt"
	"time"

	"github.com/samber/do"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"stock-tool/internal/api/jquants"
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
	c.Flags().String("dest-url", "", "destination url to save the extracted data (e.g. file://path/to/file.json)")
	c.Flags().String("code", "", "code of the listed issue to extract (optional, used for specific types)")
	c.Flags().String("start-date", "", "start date for extracting data (optional)")
	c.Flags().String("end-date", "", "end date for extracting data (optional)")
	c.MarkFlagRequired("type")
	c.MarkFlagRequired("dest-url")

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

	destURL, err := c.cmd.Flags().GetString("dest-url")
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

	client := do.MustInvoke[*jquants.Client](c.injector)

	req := &usecase.ExtractRequest{
		Source:    "jquants",
		DataType:  dataType,
		Code:      code,
		DestURL:   destURL,
		StartDate: startDate,
		EndDate:   endDate,
	}

	uc := usecase.NewExtractTaskUseCase(client)
	_, err = uc.Extract(c.cmd.Context(), req)
	if err != nil {
		return err
	}

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
