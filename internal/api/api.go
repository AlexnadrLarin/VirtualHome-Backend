package api

import (
	"bytes"
	"encoding/base64"
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

type SaveRequestData struct {
	FilePath string `json:"file_path"`
	Name     string `json:"name"`
}

type ResponseData struct {
    MeshData      json.RawMessage `json:"mesh_data"`
    PhotoBase64   string          `json:"photo_base64,omitempty"`
}

func RunScript(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to run script")
	var data RequestData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil || data.Filename == "" {
		log.Printf("Invalid request data: %v", err)
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	log.Printf("Request data: %+v", data)

	projectDir := "/home/ubuntu"
	scriptDir := projectDir + "/Neiro"
	scriptPath := "run.py"
	outputDir := "output/0"
	outputDirPy := "output"
	outputFilePath := fmt.Sprintf("%s/%s/mesh.usd", scriptDir, outputDir)

	photoFilePath := filepath.Join(scriptDir, outputDir, "example2.jpg")
	log.Printf("Script output file path: %s", outputFilePath)

	if err := os.Chdir(scriptDir); err != nil {
		log.Printf("Error changing directory to %s: %v", scriptDir, err)
		http.Error(w, fmt.Sprintf("Error changing directory: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("Changed directory to %s", scriptDir)

	cmd := exec.Command("bash", "-c", fmt.Sprintf("python %s %s --output-dir %s --bake-texture",
		scriptPath, data.Filename, outputDirPy))
	log.Printf("Executing command: %s", cmd.String())

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error running script: %v", err)
		log.Printf("Script output: %s", output)
		http.Error(w, fmt.Sprintf("Error running script: %v\nOutput:\n%s", err, output), http.StatusInternalServerError)
		return
	}
	log.Printf("Script executed successfully, output: %s", output)

	saveData := SaveRequestData{FilePath: outputFilePath, Name: "GeneratedObject"}
	saveDataBytes, _ := json.Marshal(saveData)

	resp, err := http.Post("http://90.156.217.78:8080/api/mesh", "application/json", bytes.NewBuffer(saveDataBytes))
	if err != nil {
		log.Printf("Failed to send save mesh request: %v", err)
		http.Error(w, "Failed to save mesh object 1 part", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	log.Printf("Received response from save mesh API with status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		log.Printf("Save mesh API returned an error: %d", resp.StatusCode)
		http.Error(w, "Failed to save mesh object impossible", resp.StatusCode)
		return
	}

	var saveResponse map[string]int
	if err := json.NewDecoder(resp.Body).Decode(&saveResponse); err != nil {
		log.Printf("Failed to decode save response: %v", err)
		http.Error(w, "Failed to decode save response", http.StatusInternalServerError)
		return
	}

	getMeshURL := fmt.Sprintf("http://90.156.217.78:8080/api/mesh/%d", saveResponse["id"])
	log.Printf("Fetching saved mesh data from URL: %s", getMeshURL)

	meshResp, err := http.Get(getMeshURL)
	if err != nil {
		log.Printf("Failed to fetch saved mesh data: %v", err)
		http.Error(w, "Failed to retrieve saved object", http.StatusInternalServerError)
		return
	}
	defer meshResp.Body.Close()

	if meshResp.StatusCode != http.StatusOK {
		log.Printf("Error fetching saved mesh object: %d", meshResp.StatusCode)
		http.Error(w, "Failed to retrieve object details", meshResp.StatusCode)
		return
	}

	meshData, err := ioutil.ReadAll(meshResp.Body)
	if err != nil {
		log.Printf("Failed to read mesh data: %v", err)
		http.Error(w, "Failed to read mesh data", http.StatusInternalServerError)
		return
	}
	log.Println("Successfully read mesh data")

	var meshJSON json.RawMessage
	if err := json.Unmarshal(meshData, &meshJSON); err != nil {
		log.Printf("Failed to parse mesh data: %v", err)
		http.Error(w, "Failed to parse mesh data", http.StatusInternalServerError)
		return
	}
	log.Println("Successfully parsed mesh data")

	photoData, err := ioutil.ReadFile(photoFilePath)
	if err != nil {
		log.Printf("Failed to read photo file at %s: %v", photoFilePath, err)
		http.Error(w, "Failed to read photo file", http.StatusInternalServerError)
		return
	}
	photoBase64 := base64.StdEncoding.EncodeToString(photoData)
	log.Println("Successfully encoded photo to base64")

	response := ResponseData{
		MeshData:    meshJSON,
		PhotoBase64: photoBase64,
	}
	responseBytes, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseBytes)
}

func SaveMeshObjectHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to save mesh object")

	var requestData SaveRequestData
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil || requestData.FilePath == "" || requestData.Name == "" {
		log.Printf("Invalid request data: %v", err)
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadFile(requestData.FilePath)
	if err != nil {
		log.Printf("Failed to read file at %s: %v", requestData.FilePath, err)
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	log.Println("Successfully read file data")

	meshID, err := database.SaveMeshObject(dbPool, requestData.Name, data)
	if err != nil {
		log.Printf("Failed to save object to database: %v", err)
		http.Error(w, "Failed to save object", http.StatusInternalServerError)
		return
	}
	log.Printf("Successfully saved mesh object with ID: %d", meshID)

	response := map[string]int{"id": meshID}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to send response: %v", err)
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
		return
	}
	log.Println("Response sent successfully")
}

func GetMeshObjectHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to get mesh object")

	vars := mux.Vars(r)
	idStr := vars["id"]
	log.Printf("Extracted ID from URL: %s", idStr)

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Invalid ID format: %v", err)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	log.Printf("Parsed ID: %d", id)

	mesh, err := database.GetMeshObjectByID(dbPool, id)
	if err != nil {
		log.Printf("Failed to fetch object with ID %d: %v", id, err)
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
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to send response: %v", err)
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
		return
	}
	log.Println("Response sent successfully")
}

func UploadImage(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request to upload an image")

	file, header, err := r.FormFile("image")
	if err != nil {
		log.Printf("Error retrieving the file from request: %v", err)
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	log.Printf("Received file: %s", header.Filename)

	examplesDir := "/home/ubuntu/Neiro/examples"
	destPath := filepath.Join(examplesDir, filepath.Base(header.Filename))
	log.Printf("Saving file to: %s", destPath)

	out, err := os.Create(destPath)
	if err != nil {
		log.Printf("Unable to create the file: %v", err)
		http.Error(w, "Unable to create the file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err = io.Copy(out, file); err != nil {
		log.Printf("Unable to write to the file: %v", err)
		http.Error(w, "Unable to write to the file", http.StatusInternalServerError)
		return
	}
	log.Printf("File saved successfully: %s", destPath)

	w.WriteHeader(http.StatusOK)
	responseMessage := "File uploaded successfully"
	if _, err := w.Write([]byte(responseMessage)); err != nil {
		log.Printf("Failed to send response: %v", err)
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
		return
	}
	log.Println("Response sent successfully")
}
