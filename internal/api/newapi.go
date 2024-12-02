package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"mime/multipart"
	"net/http"
	"time"

	"github.com/joho/godotenv"
)

var (
	baseURL string
	apiKey  string
	dest    string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	baseURL = os.Getenv("BASE_URL")
	apiKey = os.Getenv("API_KEY")
	dest = os.Getenv("DEST_PATH")

	if baseURL == "" || apiKey == "" || dest == "" {
		log.Fatal("BASE_URL, API_KEY, or DEST_PATH not set in .env")
	}
}

type UploadResponse struct {
	Code int `json:"code"`
	Data struct {
		ImageToken string `json:"image_token"`
	} `json:"data"`
}

type TaskResponse struct {
	Code int `json:"code"`
	Data struct {
		TaskID string `json:"task_id"`
	} `json:"data"`
}

type FinalResponse struct {
	Code int `json:"code"`
	Data struct {
		TaskID          string `json:"task_id"`
		Type            string `json:"type"`
		Status          string `json:"status"`
		Input           Input  `json:"input"`
		Output          Output `json:"output"`
		Progress        int    `json:"progress"`
		CreateTime      int64  `json:"create_time"`
		QueuingNum      int    `json:"queuing_num"`
		RunningLeftTime int    `json:"running_left_time"`
		Result          Result `json:"result"`
	} `json:"data"`
}

type Input struct {
	OriginalModelID string `json:"original_model_id"`
	Format          string `json:"format"`
	Quad            bool   `json:"quad"`
	FaceLimit       int    `json:"face_limit"`
}

type Output struct {
	Model string `json:"model"`
}

type Result struct {
	Model struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"model"`
}

func UploadFile(w http.ResponseWriter, r *http.Request) (*UploadResponse, error) {
	file, handler, err := r.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("error retrieving file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", handler.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %v", err)
	}

	if _, err = io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file content: %v", err)
	}
	writer.Close()

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/upload", baseURL), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	log.Println("firstPart: ", resp)

	var uploadResp UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return nil, fmt.Errorf("failed to parse upload response: %v", err)
	}

	return &uploadResp, nil
}

func CreateImageToModelTask(w http.ResponseWriter, imageToken string) (*TaskResponse, error) {
	data := map[string]interface{}{
		"type": "image_to_model",
		"file": map[string]string{
			"type":       "jpg",
			"file_token": imageToken,
		},
	}

	jsonData, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/task", baseURL), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	log.Println("2 Part req: ", req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	log.Println("secondPart: ", resp)

	var taskResp TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		return nil, fmt.Errorf("failed to parse task response: %v", err)
	}

	return &taskResp, nil
}

func CreateModelConversionTask(w http.ResponseWriter, taskID string) (*TaskResponse, error) {
	data := []byte(fmt.Sprintf(`{
		"type": "convert_model",
		"format": "USDZ",
		"original_model_task_id": "%s",
		"quad": true,
		"face_limit": 5000
	}`, taskID))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/task", baseURL), bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	log.Println("3 Part req: ", req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	log.Println("3 Part resp: ", resp)

	var convResp TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&convResp); err != nil {
		return nil, fmt.Errorf("failed to parse conversion task response: %v", err)
	}

	return &convResp, nil
}

func PollTask(w http.ResponseWriter, taskID string) (string, error) {
	client := &http.Client{}
	for {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/task/%s", baseURL, taskID), nil)
		if err != nil {
			return "", fmt.Errorf("failed to create request: %v", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to send request: %v", err)
		}
		defer resp.Body.Close()
		log.Println("4 Part: ", resp)

		var finalResp FinalResponse
		if err := json.NewDecoder(resp.Body).Decode(&finalResp); err != nil {
			return "", fmt.Errorf("failed to parse response: %v", err)
		}

		log.Println("4 Part status", finalResp.Data.Status)
		if finalResp.Data.Status == "success" {
			return finalResp.Data.Result.Model.URL, nil
		}

		time.Sleep(5 * time.Second)
	}
}

func downloadFile(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to get the file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: status code %d", resp.StatusCode)
	}

	outFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to copy the file content: %w", err)
	}

	fmt.Println("File downloaded successfully!")
	return nil
}

func ProcessAll(w http.ResponseWriter, r *http.Request) {
	log.Println("1 Part started")
	uploadResp, err := UploadFile(w, r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to upload file: %v", err), http.StatusInternalServerError)
		return
	}
	log.Println("1 Part finished")

	log.Println("2 Part started")
	taskResp, err := CreateImageToModelTask(w, uploadResp.Data.ImageToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create image-to-model task: %v", err), http.StatusInternalServerError)
		return
	}
	log.Println("2 Part finished")

	time.Sleep(20 * time.Second)
	log.Println("3 Part started")
	convResp, err := CreateModelConversionTask(w, taskResp.Data.TaskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create model conversion task: %v", err), http.StatusInternalServerError)
		return
	}
	log.Println("3 Part finished")

	log.Println("4 Part started")
	modelURL, err := PollTask(w, convResp.Data.TaskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to poll task: %v", err), http.StatusInternalServerError)
		return
	}
	log.Println("4 Part finished")

	log.Println("5 Part started")
	err = downloadFile(modelURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to download file: %v", err), http.StatusInternalServerError)
		return
	}
	log.Println("5 Part finished")

	saveData := SaveRequestData{FilePath: dest, Name: "GeneratedObject"}
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

	response := ResponseData{
		MeshData:    meshJSON,
	}
	responseBytes, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseBytes)
}
