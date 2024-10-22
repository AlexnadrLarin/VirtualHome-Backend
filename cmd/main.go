package main

import (
	"log"
	"time"
	"net/http"

	"go-project/internal"
)

type MeshObject struct {
	ID         int
	Name       string
	Data       []byte
	UploadTime time.Time
}

func main() {
	router := internal.SetupRouter()
	log.Fatal(http.ListenAndServe(":8080", router))
}