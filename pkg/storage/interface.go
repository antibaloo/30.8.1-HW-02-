package storage

import "tasks/pkg/storage/postgres"

type Interface interface {
	NewTask(postgres.Task) (int, error)         //Новая задача
	Tasks() ([]postgres.Task, error)            //Возвращает список всех задач
	TaskByID(int) (postgres.Task, error)        //Возвращает задачу по идентификатору
	TasksByAuthor(int) ([]postgres.Task, error) //Возвращает список задач по идентификатору постановщика
	TasksByLabel(int) ([]postgres.Task, error)  //Возвращает список задач с заданной меткой
	UpdateTask(postgres.Task) error             //Обновить данные задачи по идентификатору
	DeleteTask(int) error                       //Удалить задачу по индентификатору
}
