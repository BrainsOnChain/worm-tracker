package src

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type dbManager struct {
	db *sql.DB
}

func NewDBManager(dataSourceName string) (*dbManager, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}
	return &dbManager{db: db}, nil
}

func (db *dbManager) CreatePositionsTable() error {
	query := /* sql */ `
		CREATE TABLE IF NOT EXISTS positions (
			id SERIAL PRIMARY KEY,
			x FLOAT NOT NULL,
			y FLOAT NOT NULL,
			direction FLOAT NOT NULL,
			price FLOAT NOT NULL,
			timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`

	if _, err := db.db.Exec(query); err != nil {
		return fmt.Errorf("failed to create positions table: %w", err)
	}

	return nil
}

func (db *dbManager) savePosition(id int, position string) error {
	stmt, err := db.db.Prepare("INSERT INTO positions(id, position) VALUES(?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(id, position)
	if err != nil {
		return err
	}
	return nil
}

func (db *dbManager) fetchPositions(id int) ([]Position, error) {
	const q = /* sql */ `
		SELECT
			id,
			x,
			y,
			direction,
			price,
			timestamp
		FROM
			positions
		WHERE id > ?;
	`

	rows, err := db.db.Query(q, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	positions := make([]Position, 0)
	for rows.Next() {
		var p Position
		if err := rows.Scan(&p.ID, &p.X, &p.Y, &p.Direction, &p.Price, &p.Timestamp); err != nil {
			return nil, err
		}
		positions = append(positions, p)
	}

	return positions, nil
}

func (db *dbManager) Close() {
	if err := db.db.Close(); err != nil {
		log.Fatal(err)
	}
}
