package task

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo   Repository
	now    func() time.Time
	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		now:    func() time.Time { return time.Now().UTC() },
		logger: logger,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}
	if err := validateRepeat(input.Repeat); err != nil {
		return nil, err
	}

	now := s.now()

	model := &taskdomain.Task{
		Title:        normalized.Title,
		Description:  normalized.Description,
		Status:       normalized.Status,
		DueDate:      input.DueDate,
		RepeatConfig: input.Repeat,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	if input.Repeat != nil && input.Repeat.Type != taskdomain.RepeatNone {
		instances := GenerateInstances(created, now, now, 60)

		for _, inst := range instances {
			if err := s.repo.CreateIfNotExists(ctx, inst); err != nil {
				s.logger.Warn("failed to create recurring instance",
					"parent_id", created.ID, "due_date", inst.DueDate, "error", err)
			}
		}

		if len(instances) > 0 {
			last := instances[len(instances)-1].DueDate
			if err := s.repo.SetGeneratedUntil(ctx, created.ID, truncateDay(last)); err != nil {
				s.logger.Error("failed to update generated_until after create", "task_id", created.ID, "error", err)
			}
		}
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}
	if err := validateRepeat(input.Repeat); err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:           id,
		Title:        normalized.Title,
		Description:  normalized.Description,
		Status:       normalized.Status,
		DueDate:      input.DueDate,
		RepeatConfig: input.Repeat,
		UpdatedAt:    s.now(),
	}

	return s.repo.Update(ctx, model)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

func (s *Service) GenerateMissingInstances(ctx context.Context, horizonDays int) error {
	templates, err := s.repo.ListTemplates(ctx)
	if err != nil {
		s.logger.Error("failed to list templates", "error", err)
		return err
	}

	now := s.now()

	for i := range templates {
		t := &templates[i]

		if t.RepeatConfig != nil && t.RepeatConfig.EndType == taskdomain.EndByCount && t.RepeatConfig.MaxOccurs != nil {
			count, err := s.repo.CountChildren(ctx, t.ID)
			if err != nil {
				s.logger.Error("failed to count children", "template_id", t.ID, "error", err)
				continue
			}

			remaining := *t.RepeatConfig.MaxOccurs - count
			if remaining <= 0 {
				continue
			}

			instances := GenerateInstances(t, now, now, horizonDays)
			if len(instances) > remaining {
				instances = instances[:remaining]
			}

			for _, inst := range instances {
				if err := s.repo.CreateIfNotExists(ctx, inst); err != nil {
					s.logger.Warn("failed to create instance", "parent_id", t.ID, "error", err)
				}
			}
		} else {
			instances := GenerateInstances(t, now, now, horizonDays)
			for _, inst := range instances {
				if err := s.repo.CreateIfNotExists(ctx, inst); err != nil {
					s.logger.Warn("failed to create instance", "parent_id", t.ID, "error", err)
				}
			}
		}

		if err := s.repo.SetGeneratedUntil(ctx, t.ID, now.AddDate(0, 0, horizonDays)); err != nil {
			s.logger.Error("failed to update generated_until", "template_id", t.ID, "error", err)
		}
	}

	return nil
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}
	if input.DueDate.IsZero() {
		return CreateInput{}, fmt.Errorf("%w: due_date is required", ErrInvalidInput)
	}
	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}
	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}
	if input.DueDate.IsZero() {
		return UpdateInput{}, fmt.Errorf("%w: due_date is required", ErrInvalidInput)
	}
	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}

func validateRepeat(rc *taskdomain.RepeatConfig) error {
	if rc == nil {
		return nil
	}

	switch rc.Type {
	case taskdomain.RepeatNone, taskdomain.RepeatDaily, taskdomain.RepeatEven, taskdomain.RepeatOdd:
	case taskdomain.RepeatEveryN:
		if rc.Interval <= 0 {
			return fmt.Errorf("%w: interval must be > 0", ErrInvalidInput)
		}
	case taskdomain.RepeatMonthly:
		if len(rc.Days) == 0 {
			return fmt.Errorf("%w: days required for monthly", ErrInvalidInput)
		}
	default:
		return fmt.Errorf("%w: invalid repeat type", ErrInvalidInput)
	}

	switch rc.EndType {
	case taskdomain.EndNever, taskdomain.EndByDate, taskdomain.EndByCount:
	default:
		return fmt.Errorf("%w: invalid end type", ErrInvalidInput)
	}

	return nil
}
