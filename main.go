package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"

	"github.com/gorilla/mux"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"unique"`
	Email    string `gorm:"unique"`
	Password string
}

type Friendship struct {
	gorm.Model
	UserID   uint
	FriendID uint
}

type FriendRequest struct {
	gorm.Model
	SenderID    uint
	RecipientID uint
	Status      string // 'pending', 'accepted', 'rejected'
}

type Response struct {
	UserID   string
	FriendID string
}

// CustomClaims represents the custom claims in the JWT token.
type CustomClaims struct {
	UserID string `json:"sub"`
	jwt.StandardClaims
}

type RequestBody struct {
	UserID string `json:"user_id"`
}

type RequestBody2 struct {
	RecipientID string `json:"recipient_id"`
}

var secretKey = []byte("wefrhuioehruiovgerngvergheruvghergueweweuihuwe")

func generateJWT(userID string) (string, error) {

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func tokenHandler(writer http.ResponseWriter, request *http.Request) {
	// Extract the user ID from the request (you can customize this part)

	var requestBody RequestBody
	if err := json.NewDecoder(request.Body).Decode(&requestBody); err != nil {
		http.Error(writer, "Invalid request data", http.StatusBadRequest)
		return
	}

	token, err := generateJWT(requestBody.UserID)
	if err != nil {
		http.Error(writer, "Error generating token", http.StatusInternalServerError)
		return
	}

	response := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(response)
}

func validateToken(signedToken string) (string, error) {

	if len(signedToken) > 7 && signedToken[:7] == "Bearer " {
		signedToken = signedToken[7:]
	}
	parsedToken, err := jwt.Parse(signedToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		username := claims["user_id"].(string)
		return username, nil
	}
	return "", errors.New("Invalid token")
}

func verifyJWT(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		// Extract the token from the request (e.g., from headers)
		tokenString := request.Header.Get("Authorization")

		// Validate the token and get the user ID
		userID, err := validateToken(tokenString)
		if err != nil {
			http.Error(writer, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Attach the user ID to the request context
		ctx := context.WithValue(request.Context(), "UserID", userID)
		next(writer, request.WithContext(ctx))
	}
}

func AddFriendRequest(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody RequestBody2
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, "Invalid request data", http.StatusBadRequest)
			return
		}

		UserIDFromContext := r.Context().Value("UserID").(string)

		// Validate sender and recipient IDs (e.g., check if they exist in the database)
		sender := User{}
		reciever := User{}

		convertedRec, _ := strconv.ParseUint(requestBody.RecipientID, 10, 64)
		convertedSen, _ := strconv.ParseUint(UserIDFromContext, 10, 64)

		rec := db.Table("users").Select("*").Where("user_id=?", uint(convertedRec)).Scan(&reciever)
		sen := db.Table("users").Select("*").Where("user_id=?", uint(convertedSen)).Scan(&sender)

		if rec.Error != nil || sen.Error != nil {
			http.Error(w, "One of the user does not exist", http.StatusInternalServerError)
			return
		}

		entry := FriendRequest{SenderID: uint(convertedRec), RecipientID: uint(convertedSen), Status: "Pending"}

		tx := db.Create(&entry)

		if tx.Error != nil {
			http.Error(w, "Error storing friend request", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.WriteHeader(http.StatusCreated)

	}
}

func AcceptFriendRequest(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request FriendRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request data", http.StatusBadRequest)
			return
		}

		// Fetch the friend request from the database based on request ID
		var existingRequest FriendRequest
		if err := db.First(&existingRequest, request.ID).Error; err != nil {
			http.Error(w, "Friend request not found", http.StatusNotFound)
			return
		}

		// Update the friend request status to "accepted"
		existingRequest.Status = "accepted"
		if err := db.Save(&existingRequest).Error; err != nil {
			http.Error(w, "Error updating friend request status", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(existingRequest)
	}
}

func RejectFriendRequest(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request FriendRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request data", http.StatusBadRequest)
			return
		}

		// Fetch the friend request from the database based on request ID
		var existingRequest FriendRequest
		if err := db.First(&existingRequest, request.ID).Error; err != nil {
			http.Error(w, "Friend request not found", http.StatusNotFound)
			return
		}

		// Update the friend request status to "rejected"
		existingRequest.Status = "rejected"
		if err := db.Save(&existingRequest).Error; err != nil {
			http.Error(w, "Error updating friend request status", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(existingRequest)
	}
}

func RemoveFriend(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request Friendship
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request data", http.StatusBadRequest)
			return
		}

		// Fetch the friend relationship from the database based on friend ID
		var existingFriendship Friendship
		if err := db.First(&existingFriendship, request.ID).Error; err != nil {
			http.Error(w, "Friend relationship not found", http.StatusNotFound)
			return
		}

		// Delete the friend relationship from the database
		if err := db.Delete(&existingFriendship).Error; err != nil {
			http.Error(w, "Error deleting friend relationship", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(existingFriendship)
	}
}

func GetFriendList(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		userID := vars["user_id"]
		UserIDFromContext := r.Context().Value("UserID")

		response := []Response{}

		fmt.Println(UserIDFromContext)
		var friends []Friendship
		if err := db.Table("friendships").Where("user_id=?", userID).Select("user_id", "friend_id").Scan(&response).Error; err != nil {
			http.Error(w, "Error fetching friend list", http.StatusInternalServerError)
			return
		}

		var friendUsernames []uint
		for _, friend := range friends {
			friendUsernames = append(friendUsernames, friend.UserID)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func SetupRoutes(db *gorm.DB) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/api/friends/add", verifyJWT(AddFriendRequest(db))).Methods(http.MethodPost)
	r.HandleFunc("/api/friends/requests", AcceptFriendRequest(db)).Methods(http.MethodPost)
	r.HandleFunc("/api/friends/requests", RejectFriendRequest(db)).Methods(http.MethodPost)
	r.HandleFunc("/api/friends/remove", RemoveFriend(db)).Methods(http.MethodDelete)
	r.HandleFunc("/api/friends/list/{user_id}", verifyJWT(GetFriendList(db))).Methods(http.MethodGet)
	r.HandleFunc("/generate-token", tokenHandler).Methods(http.MethodPost)

	return r
}

func main() {
	dsn := "root:password@tcp(localhost:3306)/socialpresence?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}

	r := mux.NewRouter()

	r = SetupRoutes(db)

	port := "8080"
	fmt.Printf("Server started on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))

	fmt.Println("Connected to the MySQL database successfully!")

}
