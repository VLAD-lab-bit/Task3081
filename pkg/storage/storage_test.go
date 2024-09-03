package storage

import (
	"database/sql"
	"reflect"
	"testing"

	_ "github.com/lib/pq"
)

func TestNewStorage(t *testing.T) {
	type args struct {
		connStr string
	}
	tests := []struct {
		name    string
		args    args
		want    *Storage
		wantErr bool
	}{
		{
			name:    "Valid Connection String",
			args:    args{connStr: "user=postgres password=vlad5043 dbname=mydatabase sslmode=disable"},
			want:    &Storage{},
			wantErr: false,
		},
		{
			name:    "Invalid Connection String",
			args:    args{connStr: "invalid-connection-string"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewStorage(tt.args.connStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewStorage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Errorf("NewStorage() = nil, want non-nil")
				return
			}

			if !tt.wantErr && reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("NewStorage() = %v, want %v", reflect.TypeOf(got), reflect.TypeOf(tt.want))
			}

			if got != nil {
				err := got.Close()
				if err != nil {
					t.Errorf("Failed to close the database connection: %v", err)
				}
			}
		})
	}
}

func setupTestDB(t *testing.T) *Storage {
	connStr := "user=postgres password=vlad5043 dbname=mydatabase sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	_, err = db.Exec(`
	DROP TABLE IF EXISTS tasks_labels, tasks, labels, users;

	CREATE TABLE users (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL
	);

	CREATE TABLE labels (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL
	);

	CREATE TABLE tasks (
		id SERIAL PRIMARY KEY,
		opened BIGINT NOT NULL DEFAULT extract(epoch from now()),
		closed BIGINT DEFAULT 0,
		author_id INTEGER REFERENCES users(id),
		assigned_id INTEGER REFERENCES users(id),
		title TEXT,
		content TEXT
	);

	CREATE TABLE tasks_labels (
		task_id INTEGER REFERENCES tasks(id),
		label_id INTEGER REFERENCES labels(id)
	);

	-- Заполняем таблицу пользователей
	INSERT INTO users (name) VALUES ('Alice'), ('Bob');

	-- Заполняем таблицу меток
	INSERT INTO labels (name) VALUES ('Bug'), ('Feature');

	-- Заполняем таблицу задач
	INSERT INTO tasks (opened, closed, author_id, assigned_id, title, content)
	VALUES (extract(epoch from now()), 0, 1, 2, 'Fix login issue', 'The login form throws an error on submit'),
	       (extract(epoch from now()), 0, 2, 1, 'Add search feature', 'Implement search functionality for the dashboard');

	-- Заполняем таблицу tasks_labels
	INSERT INTO tasks_labels (task_id, label_id) VALUES (1, 1), (2, 2);
	`)
	if err != nil {
		t.Fatalf("failed to set up database schema and data: %v", err)
	}

	store, err := NewStorage(connStr)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	return store
}

func TestStorage_Close(t *testing.T) {
	tests := []struct {
		name    string
		s       *Storage
		wantErr bool
	}{
		{
			name:    "Successful Close",
			s:       setupTestDB(t),
			wantErr: false,
		},
		{
			name: "Close Already Closed",
			s: func() *Storage {
				store := setupTestDB(t)
				store.Close()
				return store
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Storage.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStorage_CreateTask(t *testing.T) {
	store := setupTestDB(t)

	task := &Task{
		AuthorID:   1, // Используем существующего пользователя
		AssignedID: 2, // Используем существующего пользователя
		Title:      "Test Task",
		Content:    "This is a test task",
	}

	taskID, err := store.CreateTask(task)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	var createdTask Task
	err = store.db.Get(&createdTask, "SELECT id, opened, closed, author_id, assigned_id, title, content FROM tasks WHERE id = $1", taskID)
	if err != nil {
		t.Fatalf("failed to get created task: %v", err)
	}

	if createdTask.AuthorID != task.AuthorID ||
		createdTask.AssignedID != task.AssignedID ||
		createdTask.Title != task.Title ||
		createdTask.Content != task.Content {
		t.Fatalf("created task data does not match input data: got %+v, want %+v", createdTask, *task)
	}
}

func TestStorage_GetAllTasks(t *testing.T) {
	store := setupTestDB(t)
	defer func() {
		if err := store.Close(); err != nil {
			t.Errorf("Failed to close the database connection: %v", err)
		}
	}()

	tests := []struct {
		name    string
		s       *Storage
		want    []Task
		wantErr bool
	}{
		{
			name: "Get all tasks",
			s:    store,
			want: []Task{
				{ID: 1, AuthorID: 1, AssignedID: 2, Title: "Fix login issue", Content: "The login form throws an error on submit"},
				{ID: 2, AuthorID: 2, AssignedID: 1, Title: "Add search feature", Content: "Implement search functionality for the dashboard"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.GetAllTasks()
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.GetAllTasks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for i := range got {
				got[i].Opened = 0
				got[i].Closed = 0
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Storage.GetAllTasks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_DeleteTask(t *testing.T) {
	store := setupTestDB(t)
	defer store.Close()

	_, err := store.db.Exec(`
		INSERT INTO tasks (opened, closed, author_id, assigned_id, title, content)
		VALUES (extract(epoch from now()), 0, 1, 2, 'Task to delete', 'This task will be deleted');
	`)
	if err != nil {
		t.Fatalf("failed to insert task for deletion test: %v", err)
	}

	var taskID int
	err = store.db.Get(&taskID, "SELECT id FROM tasks WHERE title = 'Task to delete'")
	if err != nil {
		t.Fatalf("failed to get task ID for deletion test: %v", err)
	}

	tests := []struct {
		name    string
		taskID  int
		wantErr bool
	}{
		{
			name:    "Successful Deletion",
			taskID:  taskID,
			wantErr: false,
		},
		{
			name:    "Delete Non-Existing Task",
			taskID:  999999,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.DeleteTask(tt.taskID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.DeleteTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				var count int
				err := store.db.Get(&count, "SELECT COUNT(*) FROM tasks WHERE id = $1", tt.taskID)
				if err != nil {
					t.Errorf("failed to verify task deletion: %v", err)
					return
				}

				if count != 0 {
					t.Errorf("task with ID %d was not deleted, count = %d", tt.taskID, count)
				}
			}
		})
	}
}

func TestStorage_UpdateTask(t *testing.T) {
	store := setupTestDB(t)
	defer store.Close()

	// Вставляем начальную задачу, которую будем обновлять
	_, err := store.db.Exec(`
		INSERT INTO tasks (opened, closed, author_id, assigned_id, title, content)
		VALUES (extract(epoch from now()), 0, 1, 2, 'Original Title', 'Original Content');
	`)
	if err != nil {
		t.Fatalf("failed to insert initial task for update test: %v", err)
	}

	// Получаем ID добавленной задачи
	var taskID int
	err = store.db.Get(&taskID, "SELECT id FROM tasks WHERE title = 'Original Title'")
	if err != nil {
		t.Fatalf("failed to get task ID for update test: %v", err)
	}

	tests := []struct {
		name    string
		task    *Task
		wantErr bool
	}{
		{
			name: "Successful Update",
			task: &Task{
				ID:         taskID,
				Closed:     1234567890, // Новое значение для закрытого времени
				AuthorID:   2,          // Измененное значение автора
				AssignedID: 1,          // Измененное значение назначенного
				Title:      "Updated Title",
				Content:    "Updated Content",
			},
			wantErr: false,
		},
		{
			name: "Update Non-Existing Task",
			task: &Task{
				ID:         999999, // Предполагаемый несуществующий ID
				Closed:     1234567890,
				AuthorID:   2,
				AssignedID: 1,
				Title:      "Non-Existent Title",
				Content:    "Non-Existent Content",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.UpdateTask(tt.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.UpdateTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.task.ID != 999999 { // Для существующей задачи
					// Проверяем, что задача была обновлена правильно
					var updatedTask Task
					err := store.db.Get(&updatedTask, "SELECT id, opened, closed, author_id, assigned_id, title, content FROM tasks WHERE id = $1", tt.task.ID)
					if err != nil {
						t.Errorf("failed to get updated task: %v", err)
						return
					}

					// Печатаем отладочные данные
					t.Logf("Expected task: %+v", *tt.task)
					t.Logf("Updated task: %+v", updatedTask)

					if updatedTask.Closed != tt.task.Closed ||
						updatedTask.AuthorID != tt.task.AuthorID ||
						updatedTask.AssignedID != tt.task.AssignedID ||
						updatedTask.Title != tt.task.Title ||
						updatedTask.Content != tt.task.Content {
						t.Errorf("updated task data does not match input data: got %+v, want %+v", updatedTask, *tt.task)
					}
				} else {
					// Проверяем, что несуществующая задача не была обновлена
					var count int
					err := store.db.Get(&count, "SELECT COUNT(*) FROM tasks WHERE id = $1", tt.task.ID)
					if err != nil {
						t.Errorf("failed to verify task existence: %v", err)
						return
					}

					if count != 0 {
						t.Errorf("task with ID %d should not exist, count = %d", tt.task.ID, count)
					}
				}
			}
		})
	}
}

func TestStorage_GetTasksByAuthor(t *testing.T) {
	store := setupTestDB(t)
	defer func() {
		if err := store.Close(); err != nil {
			t.Errorf("Failed to close the database connection: %v", err)
		}
	}()

	type args struct {
		authorID int
	}
	tests := []struct {
		name    string
		s       *Storage
		args    args
		want    []Task
		wantErr bool
	}{
		{
			name: "Author with tasks",
			s:    store,
			args: args{authorID: 1},
			want: []Task{
				{ID: 1, AuthorID: 1, AssignedID: 2, Title: "Fix login issue", Content: "The login form throws an error on submit"},
			},
			wantErr: false,
		},
		{
			name:    "Author without tasks",
			s:       store,
			args:    args{authorID: 3},
			want:    []Task{},
			wantErr: false,
		},
		{
			name:    "Invalid author ID",
			s:       store,
			args:    args{authorID: -1},
			want:    []Task{},
			wantErr: false,
		},
		{
			name: "Author with multiple tasks",
			s:    store,
			args: args{authorID: 2},
			want: []Task{
				{ID: 2, AuthorID: 2, AssignedID: 1, Title: "Add search feature", Content: "Implement search functionality for the dashboard"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.GetTasksByAuthor(tt.args.authorID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.GetTasksByAuthor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for i := range got {
				got[i].Opened = 0
				got[i].Closed = 0
			}

			if len(got) == 0 && len(tt.want) == 0 {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Storage.GetTasksByAuthor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_GetTasksByLabel(t *testing.T) {
	store := setupTestDB(t)
	defer func() {
		if err := store.Close(); err != nil {
			t.Errorf("Failed to close the database connection: %v", err)
		}
	}()

	type args struct {
		labelID int
	}
	tests := []struct {
		name    string
		s       *Storage
		args    args
		want    []Task
		wantErr bool
	}{
		{
			name: "Tasks with a valid label",
			s:    store,
			args: args{labelID: 1},
			want: []Task{
				{ID: 1, AuthorID: 1, AssignedID: 2, Title: "Fix login issue", Content: "The login form throws an error on submit"},
			},
			wantErr: false,
		},
		{
			name:    "No tasks for a valid label",
			s:       store,
			args:    args{labelID: 3},
			want:    []Task{},
			wantErr: false,
		},
		{
			name:    "Invalid label ID",
			s:       store,
			args:    args{labelID: -1},
			want:    []Task{},
			wantErr: false,
		},
		{
			name: "Tasks with another valid label",
			s:    store,
			args: args{labelID: 2},
			want: []Task{
				{ID: 2, AuthorID: 2, AssignedID: 1, Title: "Add search feature", Content: "Implement search functionality for the dashboard"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.GetTasksByLabel(tt.args.labelID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.GetTasksByLabel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for i := range got {
				got[i].Opened = 0
				got[i].Closed = 0
			}

			if len(got) == 0 && len(tt.want) == 0 {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Storage.GetTasksByLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}
