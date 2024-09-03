DROP TABLE IF EXISTS tasks_labels, tasks, labels, users;

-- пользователи системы
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

-- метки задач
CREATE TABLE labels (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

-- задачи
CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    opened BIGINT NOT NULL DEFAULT extract(epoch from now()), -- время создания задачи
    closed BIGINT DEFAULT 0, -- время выполнения задачи
    author_id INTEGER REFERENCES users(id) DEFAULT 0, -- автор задачи
    assigned_id INTEGER REFERENCES users(id) DEFAULT 0, -- ответственный
    title TEXT, -- название задачи
    content TEXT -- задачи
);

-- связь многие - ко- многим между задачами и метками
CREATE TABLE tasks_labels (
    task_id INTEGER REFERENCES tasks(id),
    label_id INTEGER REFERENCES labels(id)
);


INSERT INTO users (name) VALUES ('Alice');
INSERT INTO users (name) VALUES ('Bob');

-- Заполнение таблицы меток
INSERT INTO labels (name) VALUES ('Bug');
INSERT INTO labels (name) VALUES ('Feature');

-- Заполнение таблицы задач
INSERT INTO tasks (opened, closed, author_id, assigned_id, title, content) 
VALUES (extract(epoch from now()), 0, 1, 2, 'Fix login issue', 'The login form throws an error on submit');

INSERT INTO tasks (opened, closed, author_id, assigned_id, title, content) 
VALUES (extract(epoch from now()), 0, 2, 1, 'Add search feature', 'Implement search functionality for the dashboard');

-- Заполнение таблицы tasks_labels, связывающей задачи с метками
INSERT INTO tasks_labels (task_id, label_id) VALUES (1, 1); -- Задача 1 имеет метку Bug
INSERT INTO tasks_labels (task_id, label_id) VALUES (2, 2); -- Задача 2 имеет метку Feature


