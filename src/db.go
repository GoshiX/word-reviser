package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type DB struct {
	db *sql.DB
}

func initDB() DB {

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	dbPass := os.Getenv("DB_PASSWORD")

	psqlInfo := fmt.Sprintf(`host=%s port=%s user=%s 
		dbname=%s password=%s sslmode=disable`,
		dbHost, dbPort, dbUser, dbName, dbPass)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Println(err)
	}

	DBres := DB{db}
	DBres.CreateTables()

	return DBres

}

func (db *DB) Close() {
	db.db.Close()
}

func (db *DB) CreateTables() {
	res, err := db.db.Query(`
	CREATE TABLE IF NOT EXISTS words 
		(id SERIAL PRIMARY KEY, 
		word_en VARCHAR(255) NOT NULL, 
		word_ru VARCHAR(255) NOT NULL,
		aithor_id INT NOT NULL,
		show BOOLEAN NOT NULL DEFAULT TRUE,
		total INT NOT NULL DEFAULT 0, 
		corr INT NOT NULL DEFAULT 0
	);`)
	if err != nil {
		log.Println(err)
	}
	res.Close()

}

func (db *DB) AddWord(word_en string, word_ru string, author_id int) error {
	result, err := db.db.Exec("SELECT * FROM words WHERE word_en = $1 AND aithor_id = $2", word_en, author_id)
	if err != nil {
		log.Println(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println(err)
	}
	if rowsAffected != 0 {
		return errors.New("such word already exists")
	}
	_, err = db.db.Exec("INSERT INTO words (word_en, word_ru, aithor_id) VALUES ($1, $2, $3)", word_en, word_ru, author_id)
	return err
}

func (db *DB) GetRandomWord(id int) Word {
	var word Word
	err := db.db.QueryRow("SELECT * FROM words WHERE show = true AND aithor_id = $1 ORDER BY RANDOM() LIMIT 1", id).Scan(&word.Id, &word.WordEn, &word.WordRu, &word.Author, &word.Show, &word.Total, &word.Corr)
	if err != nil {
		log.Println(err)
	}
	return word
}

func (db *DB) GetWord(id int) Word {
	var word Word
	err := db.db.QueryRow("SELECT * FROM words WHERE id = $1", id).Scan(&word.Id, &word.WordEn, &word.WordRu, &word.Author, &word.Show, &word.Total, &word.Corr)
	if err != nil {
		log.Println(err)
	}
	return word
}

func (db *DB) UpdateWord(word Word) error {
	_, err := db.db.Exec("UPDATE words SET show = $1, total = $2, corr = $3 WHERE id = $4", word.Show, word.Total, word.Corr, word.Id)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (db *DB) GetWordsAmount(id int) (int, int) {
	var total, learning int
	err := db.db.QueryRow("SELECT COUNT(*) FROM words WHERE aithor_id = $1", id).Scan(&total)
	if err != nil {
		log.Println(err)
	}

	err = db.db.QueryRow("SELECT COUNT(*) FROM words WHERE aithor_id = $1 AND show = true", id).Scan(&learning)
	if err != nil {
		log.Println(err)
	}
	return total, learning
}
