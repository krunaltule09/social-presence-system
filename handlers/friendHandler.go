// handlers/friend_handler.go

package handlers

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"
)

// User represents a user.
type User struct {
	gorm.Model
	Username string
	// Add other user-related fields as needed
}

// Friendship represents a friend relationship.
type Friendship struct {
	gorm.Model
	UserID   uint
	FriendID uint
	Friend   User // Friend is a reference to the User model
}

func AddFriendRequest(w http.ResponseWriter, r *http.Request) {
	// Parse request data (sender ID, recipient ID, etc.)
	// Store the friend request in the database
	// Return appropriate response (success or error)
}

// handlers/friend_handler.go

func AcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	// Parse request data (request ID, user ID, etc.)
	// Update friend request status to "accepted" in the database
	// Return appropriate response
}

func RejectFriendRequest(w http.ResponseWriter, r *http.Request) {
	// Similar logic as AcceptFriendRequest
}

// handlers/friend_handler.go

func RemoveFriend(w http.ResponseWriter, r *http.Request) {
	// Parse request data (user ID, friend ID, etc.)
	// Remove the friend from the user's friend list in the database
	// Return appropriate response
}

// handlers/friend_handler.go
func GetFriendList(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request data (user ID, etc.)
		userID := uint(1) // Replace with actual user ID from authentication

		// Fetch the friend list from the database
		var friends []Friendship
		if err := db.Where("user_id = ?", userID).Find(&friends).Error; err != nil {
			http.Error(w, "Error fetching friend list", http.StatusInternalServerError)
			return
		}

		// Extract friend usernames
		var friendUsernames []string
		for _, friend := range friends {
			friendUsernames = append(friendUsernames, friend.Friend.Username)
		}

		// Return the friend list as a JSON response
		response := map[string]interface{}{
			"friends": friendUsernames,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
