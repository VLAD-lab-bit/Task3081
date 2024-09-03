package storage

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sqlx.DB
}

func NewStorage(connStr string) (*Storage, error) {
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

type Task struct {
	ID         int    `db:"id"`
	Opened     int64  `db:"opened"`
	Closed     int64  `db:"closed"`
	AuthorID   int    `db:"author_id"`
	AssignedID int    `db:"assigned_id"`
	Title      string `db:"title"`
	Content    string `db:"content"`
}

func (s *Storage) CreateTask(task *Task) (int, error) {
	var id int
	err := s.db.QueryRow(`
		INSERT INTO tasks (opened, closed, author_id, assigned_id, title, content)
		VALUES (EXTRACT(EPOCH FROM NOW()), 0, $1, $2, $3, $4)
		RETURNING id
	`, task.AuthorID, task.AssignedID, task.Title, task.Content).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Storage) GetAllTasks() ([]Task, error) {
	var tasks []Task
	query := `
        SELECT * FROM tasks
    `
	err := s.db.Select(&tasks, query)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *Storage) GetTasksByAuthor(authorID int) ([]Task, error) {
	var tasks []Task
	query := "SELECT * FROM tasks WHERE author_id = $1"
	err := s.db.Select(&tasks, query, authorID)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *Storage) GetTasksByLabel(labelID int) ([]Task, error) {
	var tasks []Task
	query := `
        SELECT t.* FROM tasks t
        JOIN tasks_labels tl ON t.id = tl.task_id
        WHERE tl.label_id = $1
    `
	err := s.db.Select(&tasks, query, labelID)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *Storage) UpdateTask(task *Task) error {
	query := `
        UPDATE tasks
        SET closed = $1, author_id = $2, assigned_id = $3, title = $4, content = $5
        WHERE id = $6
    `
	_, err := s.db.Exec(query, task.Closed, task.AuthorID, task.AssignedID, task.Title, task.Content, task.ID)
	return err
}

func (s *Storage) DeleteTask(taskID int) error {
	query := "DELETE FROM tasks WHERE id = $1"
	_, err := s.db.Exec(query, taskID)
	return err
}
