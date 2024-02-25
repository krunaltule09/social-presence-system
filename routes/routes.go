// routes.go

package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// SetupRoutes registers the API routes.
func SetupRoutes(db *gorm.DB) *mux.Router {
	r := mux.NewRouter()

	// Register your routes here
	r.HandleFunc("/api/friends/add", AddFriendRequest(db)).Methods(http.MethodPost)
	r.HandleFunc("/api/friends/requests", AcceptFriendRequest(db)).Methods(http.MethodPost)
	r.HandleFunc("/api/friends/requests", RejectFriendRequest(db)).Methods(http.MethodPost)
	r.HandleFunc("/api/friends/remove", RemoveFriend(db)).Methods(http.MethodDelete)
	r.HandleFunc("/api/friends/list/{user_id}", GetFriendList(db)).Methods(http.MethodGet)

	return r
}
