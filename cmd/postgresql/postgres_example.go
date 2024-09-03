package main

import (
	"fmt"
	"log"

	"Task3081/pkg/storage"
)

func main() {
	// Define the connection string for PostgreSQL.
	connStr := "user=postgres password=vlad5043 dbname=mydatabase sslmode=disable"

	// Initialize the Storage.
	store, err := storage.NewStorage(connStr)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Fatalf("Failed to close the database connection: %v", err)
		}
	}()

	// Example 1: Create a new task
	task := &storage.Task{
		AuthorID:   1,
		AssignedID: 2,
		Title:      "New Feature Request",
		Content:    "Add a new user profile page",
	}
	taskID, err := store.CreateTask(task)
	if err != nil {
		log.Fatalf("Failed to create task: %v", err)
	}
	fmt.Printf("Task created with ID: %d\n", taskID)

	// Example 2: Get all tasks
	tasks, err := store.GetAllTasks()
	if err != nil {
		log.Fatalf("Failed to retrieve tasks: %v", err)
	}
	fmt.Println("All Tasks:")
	for _, t := range tasks {
		fmt.Printf("ID: %d, Title: %s, Content: %s\n", t.ID, t.Title, t.Content)
	}

	// Example 3: Get tasks by author
	authorTasks, err := store.GetTasksByAuthor(1)
	if err != nil {
		log.Fatalf("Failed to retrieve tasks by author: %v", err)
	}
	fmt.Println("Tasks by Author ID 1:")
	for _, t := range authorTasks {
		fmt.Printf("ID: %d, Title: %s, Content: %s\n", t.ID, t.Title, t.Content)
	}

	// Example 4: Get tasks by label
	labelTasks, err := store.GetTasksByLabel(1)
	if err != nil {
		log.Fatalf("Failed to retrieve tasks by label: %v", err)
	}
	fmt.Println("Tasks with Label ID 1:")
	for _, t := range labelTasks {
		fmt.Printf("ID: %d, Title: %s, Content: %s\n", t.ID, t.Title, t.Content)
	}

	// Example 5: Update a task
	taskToUpdate := &storage.Task{
		ID:         taskID,
		AuthorID:   1,
		AssignedID: 2,
		Title:      "Updated Feature Request",
		Content:    "Update the user profile page with new fields",
		Closed:     0,
	}
	if err := store.UpdateTask(taskToUpdate); err != nil {
		log.Fatalf("Failed to update task: %v", err)
	}
	fmt.Printf("Task with ID %d has been updated\n", taskID)
	fmt.Println("All Tasks after update:")
	tasksAfterUpdate, err := store.GetAllTasks()
	if err != nil {
		log.Fatalf("Failed to retrieve tasks after update: %v", err)
	}
	for _, t := range tasksAfterUpdate {
		fmt.Printf("ID: %d, Title: %s, Content: %s\n", t.ID, t.Title, t.Content)
	}
	// Example 6: Delete a task
	if err := store.DeleteTask(taskID); err != nil {
		log.Fatalf("Failed to delete task: %v", err)
	}
	fmt.Printf("Task with ID %d has been deleted\n", taskID)

	// Example 7: Get tasks by label after deletion (to confirm deletion)
	tasksAfterDeletion, err := store.GetAllTasks()
	if err != nil {
		log.Fatalf("Failed to retrieve tasks after deletion: %v", err)
	}
	fmt.Println("All Tasks after deletion:")
	for _, t := range tasksAfterDeletion {
		fmt.Printf("ID: %d, Title: %s, Content: %s\n", t.ID, t.Title, t.Content)
	}
}
