package postgres

import (
	"context"
	"fmt"
	"strconv"
	"tasks/pkg/other"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Хранилище данных.
type Storage struct {
	db *pgxpool.Pool
}

// Конструктор, принимает строку подключения к БД.
func New(conStr string) (*Storage, error) {
	db, err := pgxpool.Connect(context.Background(), conStr)
	if err != nil {
		return nil, err
	}
	s := Storage{
		db: db,
	}
	return &s, nil
}

// Струткура задачи
type Task struct {
	ID         int
	Opened     int64
	Closed     int64
	AuthorID   int
	AssignedID int
	Title      string
	Content    string
	LabelIDs   []int
}

// NewTask создаёт новую задачу и возвращает её идентификатор.
// Используется транзакция, т.к. нам необходимо гаранированно добавить записи в разные таблицы
func (s *Storage) NewTask(t Task) (int, error) {
	var id int
	// Начинаем транзакцию
	tx, err := s.db.Begin(context.Background())
	// Обрабатываем ошибки
	if err != nil {
		return 0, err
	}

	// Создаем новую задачу
	err = tx.QueryRow(context.Background(), `
		INSERT INTO tasks (author_id, assigned_id, title, content)
		VALUES ($1, $2, $3, $4) RETURNING id;
		`,
		t.AuthorID,
		t.AssignedID,
		t.Title,
		t.Content,
	).Scan(&id)

	// Если ошибка, откатываем транзацию
	if err != nil {
		tx.Rollback(context.Background())
		return id, err
	}

	// Если переданы идентификаторы меток, довавляем их в таблицу tasks_labels
	for _, LabelID := range t.LabelIDs {
		_, err := tx.Exec(context.Background(), `
			INSERT INTO tasks_labels (task_id,label_id) VALUES ($1,$2)`,
			id,
			LabelID,
		)
		// Если ошибка, откатываем транзацию
		if err != nil {
			tx.Rollback(context.Background())
			return id, err
		}
	}
	// Выполняем транзакцию
	tx.Commit(context.Background())
	return id, nil
}

func (s *Storage) TaskByID(taskID int) (Task, error) {
	row := s.db.QueryRow(context.Background(), `
		SELECT
			id,
			opened,
			closed,
			author_id,
			assigned_id,
			title,
			content
		FROM tasks
		WHERE id = $1;
		`,
		taskID,
	)
	var task Task
	err := row.Scan(
		&task.ID,
		&task.Opened,
		&task.Closed,
		&task.AuthorID,
		&task.AssignedID,
		&task.Title,
		&task.Content,
	)
	if err != nil {
		return Task{}, err
	}
	// Проверяем есть ли метки для задачи
	labels, err := s.LabelsByTaskID(task.ID)
	if err != nil {
		fmt.Println("labels error")
		return Task{}, err
	}
	task.LabelIDs = labels
	return task, nil
}

// Tasks возвращает слайс, содержащий все задачи
func (s *Storage) Tasks() ([]Task, error) {
	rows, err := s.db.Query(context.Background(), `
		SELECT 
			id,
			opened,
			closed,
			author_id,
			assigned_id,
			title,
			content
		FROM tasks
		ORDER BY id;
		`,
	)
	if err != nil {
		return nil, err
	}
	var tasks []Task
	// перебираем результаты, сканируем значения в структуру
	for rows.Next() {
		var task Task
		err = rows.Scan(
			&task.ID,
			&task.Opened,
			&task.Closed,
			&task.AuthorID,
			&task.AssignedID,
			&task.Title,
			&task.Content,
		)
		if err != nil {
			return nil, err
		}

		// Проверяем есть ли метки для задачи
		labels, err := s.LabelsByTaskID(task.ID)
		if err != nil {
			return nil, err
		}
		task.LabelIDs = labels

		// добавляем запись в слайс
		tasks = append(tasks, task)
	}
	// возвращаем rows.Err(), если что-то пошло не так
	return tasks, rows.Err()
}

// LabelsByTaskID получает слайс меток по идентификатору задачи
func (s *Storage) LabelsByTaskID(taskID int) ([]int, error) {
	var labels []int
	// Получаем все метки задачи
	rows, err := s.db.Query(context.Background(), `
		SELECT label_id
		FROM tasks_labels
		WHERE task_id = $1`,
		taskID,
	)
	// Проверяем ошибки
	if err != nil {
		return labels, err
	}
	// перебираем результаты, сканируем значения
	for rows.Next() {
		var label int
		err = rows.Scan(&label)
		if err != nil {
			return labels, err
		}
		labels = append(labels, label)
	}
	// возвращаем rows.Err(), если что-то пошло не так
	return labels, rows.Err()
}

// TasksByAuthor возвращает список задач по автору, проверка существования пользователя с таким идентификатором не произволдится, т.к.
// подразумевается, что идентификатор пользователя выбирается из таблицы users
func (s *Storage) TasksByAuthor(authorID int) ([]Task, error) {
	rows, err := s.db.Query(context.Background(), `
		SELECT 
			id,
			opened,
			closed,
			author_id,
			assigned_id,
			title,
			content
		FROM tasks
		WHERE author_id = $1
		ORDER BY id;
		`,
		authorID,
	)
	if err != nil {
		return nil, err
	}
	var tasks []Task
	// перебираем результаты, сканируем значения в структуру
	for rows.Next() {
		var task Task
		err = rows.Scan(
			&task.ID,
			&task.Opened,
			&task.Closed,
			&task.AuthorID,
			&task.AssignedID,
			&task.Title,
			&task.Content,
		)
		if err != nil {
			return nil, err
		}

		// Проверяем есть ли метки для задачи
		labels, err := s.LabelsByTaskID(task.ID)
		if err != nil {
			return nil, err
		}
		task.LabelIDs = labels

		// добавляем запись в слайс
		tasks = append(tasks, task)
	}
	// возвращаем rows.Err(), если что-то пошло не так
	return tasks, rows.Err()
}

// TasksByLabel возвращает список задач по метке, проверка существования метки с таким идентификатором не производится, т.к. подразумевается,
// что идентификатор метки выбирается из таблицы labels
func (s *Storage) TasksByLabel(labelID int) ([]Task, error) {
	var tasks []Task
	var taskIDs []int
	// Получем список идентификаторов задач с указанной меткой
	rows, err := s.db.Query(context.Background(), `
			SELECT task_id
			FROM tasks_labels
			WHERE label_id = $1
		`,
		labelID,
	)
	// Проверяем ошибки
	if err != nil {
		return nil, err
	}

	// Перебирает результаты, формируем слайс с идентификаторами задач
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			return tasks, err
		}
		taskIDs = append(taskIDs, id)
	}

	// Техническая переменная
	var taskIDsString string
	// Формируем в технической переменной список идентификаторов
	for i, taskID := range taskIDs {
		if i > 0 {
			taskIDsString += ","
		}
		taskIDsString += strconv.Itoa(taskID)
	}
	taskIDsString = "{" + taskIDsString + "}"
	// Получаем список задачь с множеством идентификаторов
	rows, err = s.db.Query(context.Background(), `
			SELECT 
				id,
				opened,
				closed,
				author_id,
				assigned_id,
				title,
				content
			FROM tasks
			WHERE id = ANY($1::int[]);
		`,
		taskIDsString,
	)
	if err != nil {
		return tasks, err
	}
	for rows.Next() {
		var t Task
		err = rows.Scan(
			&t.ID,
			&t.Opened,
			&t.Closed,
			&t.AuthorID,
			&t.AssignedID,
			&t.Title,
			&t.Content,
		)
		if err != nil {
			return nil, err
		}

		// Проверяем есть ли метки для задачи
		labels, err := s.LabelsByTaskID(t.ID)
		if err != nil {
			return nil, err
		}
		t.LabelIDs = labels

		// добавляем запись в слайс
		tasks = append(tasks, t)
	}
	// возвращаем rows.Err(), если что-то пошло не так
	return tasks, rows.Err()
}

// UpdateTask обновляет задачу по идентификатору (заголовок, описание, автор, ответственный, метки, дата закрытия)
// Подразумевается, что структура, передаваемая в метод, получена из метода Task(int) и содержит все поля, которые были заполнены в
// исходной задаче, никакой дополнительной проверки на заполнение полей структуры нет
func (s *Storage) UpdateTask(task Task) error {
	// Запускаем транзактию
	tx, err := s.db.Begin(context.Background())
	// Обрабатываем ошибки
	if err != nil {
		return err
	}
	_, err = tx.Exec(context.Background(), `
		UPDATE tasks
		SET 
			closed = $1,
			author_id = $2,
			assigned_id = $3,
			title = $4,
			content = $5
		WHERE id = $6;
		`,
		task.Closed,
		task.AuthorID,
		task.AssignedID,
		task.Title,
		task.Content,
		task.ID,
	)
	// Если ошибка, откатывает транзакцию
	if err != nil {
		tx.Rollback(context.Background())
		return err
	}
	// Получаем текущие метки задачи
	curLabels, err := s.LabelsByTaskID(task.ID)
	if err != nil {
		return err
	}
	// Идентификаторы меток, которые надо добавить в таблицы соответствий задач и меток
	ladels2Add := other.Difference(task.LabelIDs, curLabels)
	for _, label := range ladels2Add {
		_, err = tx.Exec(context.Background(), `
			INSERT INTO tasks_labels(task_id,label_id)
			VALUES($1,$2)
			`,
			task.ID,
			label,
		)
		if err != nil {
			tx.Rollback(context.Background())
			return err
		}
	}
	// Идентификаторы меток, которые надо удалить из таблицы соответствий задач и меток
	labels2Delete := other.Difference(curLabels, task.LabelIDs)
	for _, label := range labels2Delete {
		_, err = tx.Exec(context.Background(), `
			DELETE FROM tasks_labels
			WHERE task_id = $1 AND label_id = $2
			`,
			task.ID,
			label,
		)
		if err != nil {
			tx.Rollback(context.Background())
			return err
		}
	}
	// Выполняем транзакцию
	tx.Commit(context.Background())
	return nil
}

// DeleteTask удаляет задачу по идентификатору, проверка существования задачи с идентификатором не производится, т.к. подразумевается,
// идентификатор задачи выбирается из структуры или списка структур, полученных методами Tasks, TasksByAuthor или TasksByLabel
// Используем транзакцию, т.к. необходимо гарантированно совершить операции удаления в нескольких таблицах
func (s *Storage) DeleteTask(taskID int) error {
	// Запускаем транзактию
	tx, err := s.db.Begin(context.Background())
	// Обрабатываем ошибки
	if err != nil {
		return err
	}
	// Удаляем записи их таблицы соответствия задачи и меток по идентификатору задачи,
	// делаем это до удаления записи из таблицы задача, т.к. таблица task_labels имеет внешний ключ в tasks(id) и будет ошибка
	_, err = tx.Exec(context.Background(), `
		DELETE FROM tasks_labels WHERE task_id = $1
		`,
		taskID,
	)
	// Если ошибка, откатывает транзакцию
	if err != nil {
		tx.Rollback(context.Background())
		return err
	}

	// Удаляем запись из таблицы задачи по индентификатору
	_, err = tx.Exec(context.Background(), `
		DELETE FROM tasks WHERE id = $1
		`,
		taskID,
	)
	// Если ошибка, откатывает транзакцию
	if err != nil {
		tx.Rollback(context.Background())
		return err
	}
	// Выполняем транзакцию
	tx.Commit(context.Background())
	return nil
}
