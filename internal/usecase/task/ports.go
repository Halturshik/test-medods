package task

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	CreateIfNotExists(ctx context.Context, task *taskdomain.Task) error
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
	ListTemplates(ctx context.Context) ([]taskdomain.Task, error)
	CountChildren(ctx context.Context, parentID int64) (int, error)
	SetGeneratedUntil(ctx context.Context, templateID int64, t time.Time) error
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
	GenerateMissingInstances(ctx context.Context, horizonDays int) error
}

type CreateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	DueDate     time.Time
	Repeat      *taskdomain.RepeatConfig
}

type UpdateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	DueDate     time.Time
	Repeat      *taskdomain.RepeatConfig
}
