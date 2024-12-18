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
			id        SERIAL PRIMARY KEY,
			x         FLOAT NOT NULL,
			y         FLOAT NOT NULL,
			direction FLOAT NOT NULL,
			price     FLOAT NOT NULL,
			ts        TIMESTAMP NOT NULL
		);`

	if _, err := db.db.Exec(query); err != nil {
		return fmt.Errorf("failed to create positions table: %w", err)
	}

	return nil
}

func (db *dbManager) savePosition(p position) error {
	const q = /* sql */ `
		INSERT INTO positions
			(x, y, direction, price, ts)
		VALUES
			(?, ?, ?, ?, ?)
	`

	stmt, err := db.db.Prepare(q)
	if err != nil {
		return fmt.Errorf("error preparing insert stmt: %w", err)
	}

	if _, err = stmt.Exec(p.X, p.Y, p.Direction, p.Price, p.Timestamp); err != nil {
		return fmt.Errorf("error executing position insert: %w", err)
	}

	return nil
}

func (db *dbManager) fetchPositions(id int) ([]position, error) {
	const q = /* sql */ `
		SELECT
			id, x, y, direction, price, ts
		FROM
			positions
		WHERE id > ?
		ORDER BY id DESC;
	`

	rows, err := db.db.Query(q, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	positions := make([]position, 0)
	for rows.Next() {
		var p position
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
