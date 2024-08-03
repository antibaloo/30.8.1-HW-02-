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
	err := db.NewData()
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
			fmt.Println("Создана задача:", taskID)
		}
	}

	tasks, err := db.Tasks()
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println("Список всех задач: ", tasks)
	}
	tasks, err = db.TasksByAuthor(0)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println("Задачи автора 0:", tasks)
	}
	tasks, err = db.TasksByAuthor(1)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println("Задачи автора 1:", tasks)
	}
	tasks, err = db.TasksByLabel(2)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println("Задачи зачачи с меткой 2:", tasks)
	}

	err = db.DeleteTask(2)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println("Задача #2 удалена")
	}
	tasks, err = db.Tasks()
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println("Список всех задач: ", tasks)
	}
	task, err := db.TaskByID(1)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println("Задача с номером 1: ", task)
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
		fmt.Println("Задача #1 изменена")
	}
	task, err = db.TaskByID(1)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println("Задача с номером 1: ", task)
	}
}
