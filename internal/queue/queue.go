package queue

import (
	"context"
	"path/filepath"
	"sync"
	"time"
)

type Status int

const (
	StatusPending    Status = iota
	StatusProcessing
	StatusPaused
	StatusDone
	StatusError
)

func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusProcessing:
		return "processing"
	case StatusPaused:
		return "paused"
	case StatusDone:
		return "done"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

type TaskInfo struct {
	ID       string  `json:"id"`
	FilePath string  `json:"file_path"`
	FileName string  `json:"file_name"`
	Status   Status  `json:"status"`
	Progress float64 `json:"progress"`
	ErrorMsg string  `json:"error_msg,omitempty"`
}

type Queue struct {
	mu       sync.Mutex
	tasks    []*TaskInfo
	ctx      context.Context
	cancel   context.CancelFunc
	onUpdate func()
}

func New() *Queue {
	ctx, cancel := context.WithCancel(context.Background())
	return &Queue{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (q *Queue) AddFile(path string) *TaskInfo {
	q.mu.Lock()
	defer q.mu.Unlock()

	t := &TaskInfo{
		ID:       time.Now().Format("20060102150405.000000"),
		FilePath: path,
		FileName: filepath.Base(path),
		Status:   StatusPending,
	}
	q.tasks = append(q.tasks, t)
	if q.onUpdate != nil {
		q.onUpdate()
	}
	return t
}

func (q *Queue) AddFiles(paths []string) []*TaskInfo {
	var added []*TaskInfo
	for _, p := range paths {
		added = append(added, q.AddFile(p))
	}
	return added
}

func (q *Queue) Remove(id string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, t := range q.tasks {
		if t.ID == id {
			q.tasks = append(q.tasks[:i], q.tasks[i+1:]...)
			if q.onUpdate != nil {
				q.onUpdate()
			}
			return true
		}
	}
	return false
}

func (q *Queue) Tasks() []TaskInfo {
	q.mu.Lock()
	defer q.mu.Unlock()

	out := make([]TaskInfo, len(q.tasks))
	for i, t := range q.tasks {
		out[i] = *t
	}
	return out
}

func (q *Queue) UpdateTask(id string, status Status, progress float64, errMsg string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, t := range q.tasks {
		if t.ID == id {
			t.Status = status
			t.Progress = progress
			t.ErrorMsg = errMsg
			break
		}
	}
	if q.onUpdate != nil {
		q.onUpdate()
	}
}

func (q *Queue) SetUpdateCallback(fn func()) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.onUpdate = fn
}

func (q *Queue) Context() context.Context {
	return q.ctx
}

func (q *Queue) Cancel() {
	q.cancel()
}

func (q *Queue) ResetContext() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.cancel()
	q.ctx, q.cancel = context.WithCancel(context.Background())
}
