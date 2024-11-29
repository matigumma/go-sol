package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"matu/gosol/types"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	storage := &Storage{db: db}
	if err := storage.initDB(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *Storage) initDB() error {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS mints (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		address TEXT UNIQUE NOT NULL
	);
	CREATE TABLE IF NOT EXISTS reports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		mint_id INTEGER,
		symbol TEXT,
		name TEXT,
		uri TEXT,
		mutable BOOLEAN,
		update_authority TEXT,
		score INTEGER,
		detected_at DATETIME,
		FOREIGN KEY(mint_id) REFERENCES mints(id)
	);
	`
	_, err := s.db.Exec(sqlStmt)
	return err
}

func (s *Storage) AddReport(mintAddress string, report types.Report) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	// Insertar o ignorar el mint
	_, err = tx.Exec("INSERT OR IGNORE INTO mints(address) VALUES(?)", mintAddress)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Obtener el ID del mint
	var mintID int
	err = tx.QueryRow("SELECT id FROM mints WHERE address = ?", mintAddress).Scan(&mintID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Insertar el reporte
	_, err = tx.Exec(
		"INSERT INTO reports(mint_id, symbol, name, uri, mutable, update_authority, score, detected_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?)",
		mintID, report.TokenMeta.Symbol, report.TokenMeta.Name, report.TokenMeta.URI, report.TokenMeta.Mutable,
		report.TokenMeta.UpdateAuthority, report.Score, time.Now(),
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *Storage) GetMintState() (map[string][]types.Report, error) {
	mintState := make(map[string][]types.Report)

	rows, err := s.db.Query(`
		SELECT m.address, r.symbol, r.name, r.uri, r.mutable, r.update_authority, r.score, r.detected_at
		FROM reports r
		JOIN mints m ON r.mint_id = m.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var address, symbol, name, uri, updateAuthority string
		var mutable bool
		var score int
		var detectedAt time.Time

		err = rows.Scan(&address, &symbol, &name, &uri, &mutable, &updateAuthority, &score, &detectedAt)
		if err != nil {
			return nil, err
		}

		report := types.Report{
			TokenMeta: types.TokenMeta{
				Symbol:          symbol,
				Name:            name,
				URI:             uri,
				Mutable:         mutable,
				UpdateAuthority: updateAuthority,
			},
			Score:      score,
			DetectedAt: detectedAt,
		}

		mintState[address] = append(mintState[address], report)
	}

	return mintState, nil
}
