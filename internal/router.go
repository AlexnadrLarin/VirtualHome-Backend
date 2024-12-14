package internal

import (
	"go-project/internal/api"

	"github.com/gorilla/mux"
)

func SetupRouter() (*mux.Router) {
	router := mux.NewRouter()

    //1 нейронка
	router.HandleFunc("/api/run-script", api.RunScript).Methods("POST")

    //2 нейронка
    router.HandleFunc("/api/newrun-script", api.ProcessAll).Methods("POST")

    router.HandleFunc("/api/mesh", api.SaveMeshObjectHandler).Methods("POST")
	router.HandleFunc("/api/mesh/{id:[0-9]+}", api.GetMeshObjectHandler).Methods("GET")
	router.HandleFunc("/api/upload", api.UploadImage).Methods("POST")

	router.HandleFunc("/api/register", api.RegisterHandler)
	router.HandleFunc("/api/login", api.LoginHandler)
	router.HandleFunc("/api/update-profile", api.UpdateProfileHandler)
	router.HandleFunc("/api/add-item", api.AddCatalogItemHandler)

	return router
}
