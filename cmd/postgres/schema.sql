-- удаление таблиц, если существуют
drop table if exists tasks_labels,tasks,labels,users;

-- таблица пользователей
create table if not exists users(
	id serial primary key,
	name text not null
);

-- таблица меток
create table if not exists labels(
	id serial primary key,
	name text not null
);

--таблица задач
create table if not exists tasks(
	id serial primary key,
	opened bigint DEFAULT extract(epoch from now()),
	closed bigint default 0,
	author_id integer references users(id) default 0,
	assigned_id integer references users(id) default 0,
	title text not null,
	content text not null
);

-- таблица соответсвия для меток и задач
create table if not exists tasks_labels(
	task_id integer references tasks(id),
	label_id integer references labels(id)
);

-- пользователь по умолчанию с id = 0
INSERT INTO users (id, name) VALUES (0, 'default'), (1,'Andrey');

-- демонстрационные метки
INSERT INTO labels (name) VALUES ('important'),('project'),('no task');