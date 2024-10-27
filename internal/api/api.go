package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"go-project/internal/database"

	"github.com/gorilla/mux"
)

var dbPool, _ = database.ConnectDB()

type MeshObjectResponse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	UploadTime string `json:"upload_time"`
	Data       string `json:"data"` 
}

type RequestData struct {
    Filename string `json:"filename"`
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

func RunScript(w http.ResponseWriter, r *http.Request) {

    var data RequestData
    err := json.NewDecoder(r.Body).Decode(&data)
    if err != nil || data.Filename == "" {
        http.Error(w, "Invalid request data", http.StatusBadRequest)
        return
    }

	scriptDir := "/"
    scriptPath := "run.py"
    outputDir := "output/"

    cwd, _ := os.Getwd()
    fmt.Printf("Current working directory: %s\n", cwd)

	if err := os.Chdir(scriptDir); err != nil {
        http.Error(w, fmt.Sprintf("Error changing directory: %v", err), http.StatusInternalServerError)
        return
    }

	cwd, _ = os.Getwd()
    fmt.Printf("Changed working directory to: %s\n", cwd)

    cmd := exec.Command("python3", scriptPath, data.Filename, "--output-dir", outputDir)

    output, err := cmd.CombinedOutput()
    if err != nil {
		log.Printf("Script output: %s", output)
        http.Error(w, fmt.Sprintf("Error running script: %v", err), http.StatusInternalServerError)
        return
    }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}