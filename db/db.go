package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init() {
	// Koneksi ke database default `postgres`
	connStr := "user=dimas password=root dbname=postgres sslmode=disable"
	dbTemp, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Gagal koneksi ke DB postgres:", err)
	}
	defer dbTemp.Close()

	// Cek apakah database `golang_log` sudah ada
	var exists bool
	err = dbTemp.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = 'golang_log')").Scan(&exists)
	if err != nil {
		log.Fatal("Gagal cek database:", err)
	}

	if !exists {
		_, err = dbTemp.Exec("CREATE DATABASE golang_log")
		if err != nil {
			log.Fatal("Gagal membuat database golang_log:", err)
		}
		fmt.Println("Database golang_log berhasil dibuat")
	} else {
		fmt.Println("Database golang_log sudah ada")
	}

	// Sekarang koneksi ke `golang_log`
	connStr = "user=dimas password=root dbname=golang_log sslmode=disable"
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Gagal koneksi ke database golang_log:", err)
	}
	err = DB.Ping()
	if err != nil {
		log.Fatal("Ping ke golang_log gagal:", err)
	}
	fmt.Println("Koneksi ke database golang_log berhasil")

	// Buat tabel `users`
	createTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		password TEXT NOT NULL
	);
	`
	_, err = DB.Exec(createTable)
	if err != nil {
		log.Fatal("Gagal membuat tabel users:", err)
	}
	fmt.Println("Tabel users siap")
}
