package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type MeshObject struct {
	ID         int
	Name       string
	Data       []byte
	UploadTime time.Time
}

func ConnectDB() (*pgxpool.Pool, error) {
	err := godotenv.Load()
    if err != nil {
        fmt.Errorf("Ошибка при загрузке файла .env: %v", err)
    }

	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	user := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	dbname := os.Getenv("DATABASE")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",  
		user, password, host, port, dbname)
	config, err := pgxpool.ParseConfig(dsn)

	if err != nil {
		return nil, fmt.Errorf("unable to parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create pool: %w", err)
	}

	fmt.Println("Successfully connected!")
	return pool, nil
}

func SaveMeshObject(db *pgxpool.Pool, name string, data []byte) (int, error) {
	var id int
	query := `INSERT INTO mesh_objects (name, data) VALUES ($1, $2) RETURNING id`
	err := db.QueryRow(context.Background(), query, name, data).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetMeshObjectByID(db *pgxpool.Pool, id int) (*MeshObject, error) {
	query := `SELECT id, name, data, upload_time FROM mesh_objects WHERE id = $1`
	row := db.QueryRow(context.Background(), query, id)

	var mesh MeshObject
	err := row.Scan(&mesh.ID, &mesh.Name, &mesh.Data, &mesh.UploadTime)
	if err != nil {
		return nil, err
	}

	return &mesh, nil
}
