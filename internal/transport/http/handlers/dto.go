package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string                   `json:"title"`
	Description string                   `json:"description"`
	Status      taskdomain.Status        `json:"status"`
	DueDate     time.Time                `json:"due_date"`
	Repeat      *taskdomain.RepeatConfig `json:"repeat,omitempty"`
}

type taskDTO struct {
	ID          int64                    `json:"id"`
	ParentID    *int64                   `json:"parent_id,omitempty"`
	Title       string                   `json:"title"`
	Description string                   `json:"description"`
	Status      taskdomain.Status        `json:"status"`
	DueDate     time.Time                `json:"due_date"`
	Repeat      *taskdomain.RepeatConfig `json:"repeat,omitempty"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		ParentID:    task.ParentID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		DueDate:     task.DueDate,
		Repeat:      task.RepeatConfig,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}
