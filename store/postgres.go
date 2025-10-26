package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"taskflow/models"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db_url string) (TaskStore, error) {
	db, err := sql.Open("pgx", db_url)
	if err != nil {
		log.Printf("ERROR: unable to open database connection: %v\n", err)
		return nil, fmt.Errorf("ERROR: unable to open database connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Printf("ERROR: unable to ping the database: %v\n", err)
		return nil, fmt.Errorf("ERROR: unable to ping the database: %v", err)
	}

	log.Println("successfuly connected to PostgreSQL")
	return &PostgresStore{
		db: db,
	}, nil
}

func (ps *PostgresStore) Create(task *models.Task) {
	sql_statement := `
		INSERT INTO tasks (id, payload, status, error, created_at) 
		VALUES($1, $2, $3, $4, $5) `

	payload_bytes, err := json.Marshal(task.Payload)
	if err != nil {
		log.Printf("ERROR: unable to parse the payload: %v\n", err)
		return
	}

	_, err = ps.db.Exec(sql_statement, task.ID, payload_bytes, task.Status, "", task.CreatedAt)
	if err != nil {
		log.Printf("ERROR: unable to add task to databse: %v\n", err)
		return
	}
}

func (ps *PostgresStore) Get(id string) (*models.Task, error) {
	sql_statement := `
	SELECT id, payload, result, status, error, created_at, finished_at from tasks 
	WHERE id = $1`

	var task models.Task
	var result_bytes, payload_bytes []byte
	var finished_at_null sql.NullTime
	var error_null sql.NullString

	err := ps.db.QueryRow(sql_statement, id).Scan(
		&task.ID,
		&payload_bytes,
		&result_bytes,
		&task.Status,
		&error_null,
		&task.CreatedAt,
		&finished_at_null,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("ERROR: No tasks found with given id: %v", err)
			return nil, fmt.Errorf("ERROR: No tasks found with given id: %v", err)
		} else {
			log.Printf("ERROR: unable to retrieve the task from database: %v", err)
			return nil, fmt.Errorf("ERROR: unable to retrieve the task from database: %v", err)
		}
	}

	if payload_bytes != nil {
		if err := json.Unmarshal(payload_bytes, &task.Payload); err != nil {
			log.Printf("ERROR: unable to parse the payload to json: %v", err)
			return nil, fmt.Errorf("ERROR: unable to parse the payload to json: %v", err)
		}
	}

	task.Result = result_bytes

	if finished_at_null.Valid {
		task.FinishedAt = finished_at_null.Time
	}

	if error_null.Valid {
		task.Error = error_null.String
	}

	return &task, nil
}

func (ps *PostgresStore) GetAll() []*models.Task {
	sql_statement := `
	SELECT id, payload, result, status, error, created_at, finished_at from tasks`

	tasks := make([]*models.Task, 0)
	rows, err := ps.db.Query(sql_statement)
	if err != nil {
		log.Printf("ERROR: unable to fetch the tasks from database: %v", err)
		return tasks
	}

	defer rows.Close()

	for rows.Next() {
		var task models.Task
		var payload_bytes, result_bytes []byte
		var finished_at_null sql.NullTime
		var error_null sql.NullString

		if err := rows.Scan(&task.ID, &payload_bytes, &result_bytes, &task.Status, &error_null, &task.CreatedAt, &finished_at_null); err != nil {
			log.Printf("ERROR: failed to scan row: %v", err)
			continue
		}

		if payload_bytes != nil {
			if err := json.Unmarshal(payload_bytes, &task.Payload); err != nil {
				log.Printf("ERROR: unable to parse the payload to json: %v", err)
				continue
			}
		}

		task.Result = result_bytes

		if finished_at_null.Valid {
			task.FinishedAt = finished_at_null.Time
		}

		if error_null.Valid {
			task.Error = error_null.String
		}

		tasks = append(tasks, &task)
	}

	return tasks
}

func (ps *PostgresStore) Update(task *models.Task) error {
	sql_statement := `
	UPDATE tasks SET status=$2, result=$3, error=$4, finished_at=$5 
	WHERE id=$1`

	_, err := ps.db.Exec(sql_statement, task.ID, task.Status, task.Result, task.Error, task.FinishedAt)
	return err
}
