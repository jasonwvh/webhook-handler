package app

import (
	"database/sql"
	"log"

	"github.com/jasonwvh/webhook-handler/internal/models"
	"github.com/mattn/go-sqlite3"
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS work_items (
			id INTEGER PRIMARY KEY,
			url TEXT NOT NULL,
			seq INTEGER NOT NULL
		)
	`)
	if err != nil {
		return nil, err
	}

	return &SQLiteStorage{db: db}, nil
}

func (s *SQLiteStorage) GetWorkItem(id int) (*models.WorkItem, error) {
	var url string
	var seq int
	err := s.db.QueryRow("SELECT * FROM work_items WHERE id = ?", id).Scan(&id, &url, &seq)
	if err != nil {
		return nil, err
	}

	return &models.WorkItem{ID: id, URL: url, Seq: seq}, nil
}

func (s *SQLiteStorage) StoreWorkItem(workItem *models.WorkItem) error {
	statement, err := s.db.Prepare("INSERT INTO work_items (id, url, seq) VALUES (?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec(workItem.ID, workItem.URL, workItem.Seq)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint {
			log.Printf("work item already processed")
		}
		return err
	}

	log.Printf("stored item: %v", workItem)
	return nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
