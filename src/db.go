package src

import (
	"database/sql"
	"errors"
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
	createPositions := /* sql */ `
		CREATE TABLE IF NOT EXISTS positions (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			blck      INTEGER NOT NULL, -- the block number
			x         FLOAT NOT NULL,
			y         FLOAT NOT NULL,
			direction FLOAT NOT NULL,
			price     FLOAT NOT NULL,
			ts        TIMESTAMP NOT NULL
		);`

	if _, err := db.db.Exec(createPositions); err != nil {
		return fmt.Errorf("failed to create positions table: %w", err)
	}

	createBlocksChecked := /* sql */ `
		CREATE TABLE IF NOT EXISTS blocks_checked (
			blck INTEGER PRIMARY KEY
		);`

	if _, err := db.db.Exec(createBlocksChecked); err != nil {
		return fmt.Errorf("failed to create blocks_checked table: %w", err)
	}

	// insert 0 as the first block checked if it doesn't exist
	const q = /* sql */ `
		INSERT OR IGNORE INTO blocks_checked (blck) VALUES (0);
	`

	if _, err := db.db.Exec(q); err != nil {
		return fmt.Errorf("failed to insert 0 into blocks_checked: %w", err)
	}

	return nil
}

func (db *dbManager) savePosition(p position) error {
	const q = /* sql */ `
		INSERT INTO positions
			(blck, x, y, direction, price, ts)
		VALUES
			(?, ?, ?, ?, ?, ?)
	`

	stmt, err := db.db.Prepare(q)
	if err != nil {
		return fmt.Errorf("error preparing insert stmt: %w", err)
	}

	if _, err = stmt.Exec(p.block, p.X, p.Y, p.Direction, p.Price, p.Timestamp); err != nil {
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
		ORDER BY id ASC;
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

func (db *dbManager) getLatestPosition() (position, error) {
	const q = /* sql */ `
		SELECT
			id, x, y, direction, price, ts
		FROM positions
		WHERE id = (SELECT MAX(id) FROM positions);
	`

	var p position
	if err := db.db.QueryRow(q).Scan(&p.ID, &p.X, &p.Y, &p.Direction, &p.Price, &p.Timestamp); err != nil {
		// check for now rows
		if errors.Is(err, sql.ErrNoRows) {
			return position{}, nil
		}
		return position{}, fmt.Errorf("error getting latest position: %w", err)
	}

	return p, nil
}

func (db *dbManager) saveBlockChecked(blck int) error {
	const q = /* sql */ `
		INSERT INTO blocks_checked (blck) VALUES (?);
	`

	stmt, err := db.db.Prepare(q)
	if err != nil {
		return fmt.Errorf("error preparing insert stmt: %w", err)
	}

	if _, err = stmt.Exec(blck); err != nil {
		return fmt.Errorf("error executing block insert: %w", err)
	}

	return nil
}

func (db *dbManager) getLatestBlockChecked() (int, error) {
	const q = /* sql */ `
		SELECT MAX(blck) FROM blocks_checked;
	`

	var blck int
	if err := db.db.QueryRow(q).Scan(&blck); err != nil {
		// check for no rows
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("error getting latest block checked: %w", err)
	}

	return blck, nil
}

func (db *dbManager) Close() {
	if err := db.db.Close(); err != nil {
		log.Fatal(err)
	}
}
