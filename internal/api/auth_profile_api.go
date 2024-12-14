package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go-project/internal/database"
)

type User struct {
    ID          int    `json:"id"`
    Nickname    string `json:"nickname"`
    Password    string `json:"password"`
}

type Profile struct {
    ID         int    `json:"id"`
    UserID     int    `json:"user_id"`
    FirstName  string `json:"first_name"`
    LastName   string `json:"last_name"`
    Bio        string `json:"bio"`
}

type Item struct {
    ID         int       `json:"id"`
    CatalogID  int       `json:"catalog_id"`
    Name       string    `json:"name"`
    Object3D   []byte    `json:"object_3d"`
    Photo      []byte    `json:"photo"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	userID, err := database.CreateUser(DbPool, user.Nickname, user.Password)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"user_id": userID})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	existingUser, err := database.GetUserByNickName(DbPool, user.Nickname)
	if err != nil || existingUser.Password != user.Password {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}

func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
    var req Profile
    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    if req.UserID == 0 {
        http.Error(w, "Invalid user_id", http.StatusBadRequest)
        return
    }

	_, err = database.CreateProfile(DbPool, req.UserID, req.FirstName, req.LastName, req.Bio)
	if err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated successfully"})
}

func AddCatalogItemHandler(w http.ResponseWriter, r *http.Request) {
	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		fmt.Print(err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	fmt.Println(item.CatalogID, item.Name, item.Object3D, item.Photo)
	itemID, err := database.CreateItem(DbPool, item.CatalogID, item.Name, item.Object3D, item.Photo)
	if err != nil {
		fmt.Print(err)
		http.Error(w, "Failed to add item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"item_id": itemID})
}
