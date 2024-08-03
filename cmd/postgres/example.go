package main

import (
	"fmt"
	"log"
	"os"
	"tasks/pkg/storage"
	"tasks/pkg/storage/postgres"
	"time"
)

var (
	db  storage.Interface
	err error
)

func main() {
	pwd := os.Getenv("postgrespass")
	conString := "postgres://postgres:" + pwd + "@localhost:5432/tasks?sslmode=disable"
	db, err = postgres.New(conString)
	if err != nil {
		log.Fatal(err)
	}

	newTasks := []postgres.Task{
		{
			Title:    "Новая задача",
			Content:  "Описание новой задачи",
			LabelIDs: []int{1},
		},
		{
			AuthorID:   1,
			AssignedID: 0,
			Title:      "Новая задача",
			Content:    "Описание новой задачи",
			LabelIDs:   []int{1, 2},
		},
		{
			AuthorID:   0,
			AssignedID: 1,
			Title:      "Новая задача",
			Content:    "Описание новой задачи",
			LabelIDs:   []int{2},
		},
	}
	for _, task := range newTasks {
		taskID, err := db.NewTask(task)
		if err != nil {
			log.Println(err)
		} else {
			fmt.Println(taskID)
		}
	}

	tasks, err := db.Tasks()
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(tasks)
	}
	tasks, err = db.TasksByAuthor(0)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(tasks)
	}
	tasks, err = db.TasksByAuthor(1)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(tasks)
	}
	tasks, err = db.TasksByLabel(2)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(tasks)
	}

	err = db.DeleteTask(2)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println("Task #2 was deleted")
	}

	task, err := db.TaskByID(1)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(tasks)
	}
	task.Content += "_new"
	task.Title += "_new"
	task.Closed = time.Now().Unix()
	task.AssignedID = 0
	task.AuthorID = 0
	task.LabelIDs = []int{1, 2, 3}
	err = db.UpdateTask(task)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println("Task #1 was updated")
	}
	task, err = db.TaskByID(1)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(tasks)
	}
}
