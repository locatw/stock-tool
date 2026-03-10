package usecase

import (
	"context"
	"fmt"
	"time"

	"stock-tool/internal/domain/extract"
)

// BrandDataFetcher fetches raw brand data from an external API.
type BrandDataFetcher interface {
	FetchBrands(ctx context.Context, code *string, date *time.Time) (rawBody []byte, err error)
}

// ObjectWriter writes data to object storage.
type ObjectWriter interface {
	PutObject(ctx context.Context, key string, data []byte) error
}

// ExtractTaskRepository provides persistence for extract task entities and their executions.
type ExtractTaskRepository interface {
	// Create persists a new ExtractTask.
	Create(ctx context.Context, task *extract.ExtractTask) error

	// FindBySourceAndDataType returns the task matching the given key,
	// or (nil, nil) if not found.
	FindBySourceAndDataType(
		ctx context.Context,
		source string,
		dataType string,
		timing string,
	) (*extract.ExtractTask, error)

	// CreateExecution persists a new execution under the given task.
	// Returns the created execution with server-assigned fields.
	CreateExecution(
		ctx context.Context,
		taskID int,
		exec *extract.ExtractTaskExecution,
	) (*extract.ExtractTaskExecution, error)

	// UpdateExecution persists status changes to an existing execution.
	UpdateExecution(ctx context.Context, exec *extract.ExtractTaskExecution) error

	// CreateExtractedDataS3 persists an S3 file record under the given execution.
	CreateExtractedDataS3(
		ctx context.Context,
		executionID int,
		s3File *extract.ExtractedDataS3,
	) (*extract.ExtractedDataS3, error)
}

type ExtractTaskUseCase struct {
	brandFetcher BrandDataFetcher
	objectWriter ObjectWriter
	repo         ExtractTaskRepository
}

func NewExtractTaskUseCase(
	brandFetcher BrandDataFetcher,
	objectWriter ObjectWriter,
	repo ExtractTaskRepository,
) *ExtractTaskUseCase {
	return &ExtractTaskUseCase{
		brandFetcher: brandFetcher,
		objectWriter: objectWriter,
		repo:         repo,
	}
}

// Extract fetches raw data from a source API and stores it in S3.
//
// Processing flow:
//  1. Find or create ExtractTask for (source, dataType, timing)
//  2. Create a running ExtractTaskExecution
//  3. Fetch raw data from the source API
//  4. Upload raw data to S3
//  5. Record S3 key in ExtractedDataS3
//  6. Mark execution as succeeded
//
// On failure at steps 3-5, the execution is marked as failed before
// returning the error.
//
// See doc/spec/data-ingestion/usecase/ingest-data.md for requirements.
func (uc *ExtractTaskUseCase) Extract(ctx context.Context, req *ExtractTaskRequest) (*ExtractTaskResponse, error) {
	// 1. Find-or-create the ExtractTask
	task, err := uc.findOrCreateTask(ctx, req.Source, req.DataType, req.Timing)
	if err != nil {
		return nil, err
	}

	// 2. Create a running execution
	now := time.Now()
	execution := extract.NewRunningExecution(now)
	execution, err = uc.repo.CreateExecution(ctx, task.ID(), execution)
	if err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	// 3. Fetch raw data from API
	rawBody, err := uc.fetchRawData(ctx, req)
	if err != nil {
		execution.Fail(err.Error())
		if updateErr := uc.repo.UpdateExecution(ctx, execution); updateErr != nil {
			return nil, fmt.Errorf(
				"failed to update execution status after error: %w (original: %w)",
				updateErr, err,
			)
		}
		return nil, err
	}

	// 4. Upload to S3
	s3Key := extract.GenerateS3Key(req.Source, req.DataType, now, "json")
	if err := uc.objectWriter.PutObject(ctx, s3Key, rawBody); err != nil {
		err = fmt.Errorf("failed to upload to S3: %w", err)
		execution.Fail(err.Error())
		if updateErr := uc.repo.UpdateExecution(ctx, execution); updateErr != nil {
			return nil, fmt.Errorf(
				"failed to update execution status after error: %w (original: %w)",
				updateErr, err,
			)
		}
		return nil, err
	}

	// 5. Record S3 file in DB
	s3File := extract.NewExtractedDataS3(s3Key)
	if _, err := uc.repo.CreateExtractedDataS3(ctx, execution.ID(), s3File); err != nil {
		execution.Fail(fmt.Sprintf("failed to record S3 file: %s", err.Error()))
		_ = uc.repo.UpdateExecution(ctx, execution)
		return nil, fmt.Errorf("failed to record S3 file: %w", err)
	}

	// 6. Mark execution as succeeded
	execution.Succeed()
	if err := uc.repo.UpdateExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to update execution status: %w", err)
	}

	return &ExtractTaskResponse{
		S3Key:  s3Key,
		Status: extract.ExecutionStatusSucceeded,
	}, nil
}

func (uc *ExtractTaskUseCase) findOrCreateTask(
	ctx context.Context,
	source string,
	dataType string,
	timing string,
) (*extract.ExtractTask, error) {
	task, err := uc.repo.FindBySourceAndDataType(ctx, source, dataType, timing)
	if err != nil {
		return nil, fmt.Errorf("failed to find extract task: %w", err)
	}
	if task != nil {
		return task, nil
	}

	newTask := extract.NewExtractTask(source, dataType, timing)
	if err := uc.repo.Create(ctx, newTask); err != nil {
		return nil, fmt.Errorf("failed to create extract task: %w", err)
	}

	// Re-fetch to get the assigned ID
	task, err = uc.repo.FindBySourceAndDataType(ctx, source, dataType, timing)
	if err != nil {
		return nil, fmt.Errorf("failed to find created extract task: %w", err)
	}
	return task, nil
}

func (uc *ExtractTaskUseCase) fetchRawData(ctx context.Context, req *ExtractTaskRequest) ([]byte, error) {
	switch req.Source {
	case "jquants":
		switch req.DataType {
		case "brand":
			return uc.brandFetcher.FetchBrands(ctx, req.Code, req.StartDate)
		default:
			return nil, fmt.Errorf("unsupported data type: %s.%s", req.Source, req.DataType)
		}
	default:
		return nil, fmt.Errorf("unsupported source: %s", req.Source)
	}
}
