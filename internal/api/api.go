package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go-project/internal/database"
)

var dbPool, _ = database.ConnectDB()

type MeshObjectResponse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	UploadTime string `json:"upload_time"`
	Data       string `json:"data"` 
}

func SaveMeshObjectHandler(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file upload", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	meshID, err := database.SaveMeshObject(dbPool, name, data)
	if err != nil {
		http.Error(w, "Failed to save object", http.StatusInternalServerError)
		return
	}

	response := map[string]int{"id": meshID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetMeshObjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	mesh, err := database.GetMeshObjectByID(dbPool, id)
	if err != nil {
		http.Error(w, "Object not found", http.StatusNotFound)
		return
	}

	response := MeshObjectResponse{
		ID:         mesh.ID,
		Name:       mesh.Name,
		UploadTime: mesh.UploadTime.Format("2006-01-02 15:04:05"),
		Data:       fmt.Sprintf("%x", mesh.Data), 
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}