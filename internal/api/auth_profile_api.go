package api

import (
	"encoding/json"
	"net/http"

	"go-project/internal/database"
)


func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user struct {
		Nickname string `json:"nickname"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	userID, err := database.CreateUser(DbPool, user.Nickname, user.Password)
	if err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"user_id": userID})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user struct {
		Nickname string `json:"nickname"`
		Password string `json:"password"`
	}

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
	var profile database.Profile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	_, err := database.CreateProfile(DbPool, profile.UserID, profile.FirstName, profile.LastName, profile.Bio)
	if err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated successfully"})
}

func AddCatalogItemHandler(w http.ResponseWriter, r *http.Request) {
	var item database.Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	itemID, err := database.CreateItem(DbPool, item.CatalogID, item.Name, item.Object3D, item.Photo)
	if err != nil {
		http.Error(w, "Failed to add item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"item_id": itemID})
}
