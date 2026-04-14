package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type RepeatType string

const (
	RepeatNone    RepeatType = "none"
	RepeatDaily   RepeatType = "daily"
	RepeatEveryN  RepeatType = "every_n_days"
	RepeatMonthly RepeatType = "monthly"
	RepeatEven    RepeatType = "even"
	RepeatOdd     RepeatType = "odd"
)

type EndType string

const (
	EndNever   EndType = "never"
	EndByDate  EndType = "by_date"
	EndByCount EndType = "by_count"
)

type RepeatConfig struct {
	Type      RepeatType `json:"type"`
	Interval  int        `json:"interval,omitempty"`
	Days      []int      `json:"days,omitempty"`
	EndType   EndType    `json:"end_type"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	MaxOccurs *int       `json:"max_occurs,omitempty"`
}

type Task struct {
	ID          int64     `json:"id"`
	ParentID    *int64    `json:"parent_id,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	DueDate     time.Time `json:"due_date"`

	RepeatConfig   *RepeatConfig `json:"repeat,omitempty"`
	GeneratedUntil *time.Time    `json:"generated_until,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}
