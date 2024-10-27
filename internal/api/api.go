package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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

func UploadImage(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	examplesDir := "/home/ubuntu/Neiro/examples"

	destPath := filepath.Join(examplesDir, filepath.Base(header.Filename))
	out, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "Unable to create the file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Unable to write to the file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded successfully"))
}

func RunScript(w http.ResponseWriter, r *http.Request) {

	var data RequestData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil || data.Filename == "" {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	projectDir := "/home/ubuntu"
	scriptDir := projectDir + "/Neiro"
	scriptPath := "run.py"
	outputDir := "output/"

	if err := os.Chdir(scriptDir); err != nil {
		log.Printf("Error changing directory: %v", err)
		http.Error(w, fmt.Sprintf("Error changing directory: %v", err), http.StatusInternalServerError)
		return
	}

	cmd := exec.Command("bash", "-c", fmt.Sprintf("source %s/myenv/bin/activate && python %s %s --output-dir %s",
		projectDir, scriptPath, data.Filename, outputDir))

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error running script: %v", err)
		log.Printf("Script output: %s", output)
		http.Error(w, fmt.Sprintf("Error running script: %v\nOutput:\n%s", err, output), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}
