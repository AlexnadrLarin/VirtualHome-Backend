package internal

import (
	"go-project/internal/api"

	"github.com/gorilla/mux"
)

func SetupRouter() (*mux.Router) {
	router := mux.NewRouter()

	router.HandleFunc("/api/mesh", api.SaveMeshObjectHandler).Methods("POST")
	router.HandleFunc("/api/mesh/{id:[0-9]+}", api.GetMeshObjectHandler).Methods("GET")
	router.HandleFunc("/api/upload", api.UploadImage).Methods("POST")
	router.HandleFunc("/api/run-script", api.RunScript).Methods("POST")

	return router
}
