package main

import (
	"fmt"
	//"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"

	//"strings"
	"time"

	//lksdk "github.com/livekit/server-sdk-go"
	"github.com/livekit/protocol/auth"
)

// Generate JSON string with identity and accessToken fields
func GenerateJSONString(identity, accessToken string) (string, error) {
	data := struct {
		Identity    string `json:"identity"`
		AccessToken string `json:"accessToken"`
	}{
		Identity:    identity,
		AccessToken: accessToken,
	}

	jsonString, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to generate JSON string: %v", err)
	}

	return string(jsonString), nil
}

func getJoinToken(apiKey, apiSecret, room, identity string) (string, error) {
	canPublish := true
	canSubscribe := true

	at := auth.NewAccessToken(apiKey, apiSecret)
	grant := &auth.VideoGrant{
		RoomJoin:     true,
		Room:         room,
		CanPublish:   &canPublish,
		CanSubscribe: &canSubscribe,
	}
	at.AddGrant(grant).
		SetIdentity(identity).
		SetValidFor(time.Hour)

	jwtToken, err := at.ToJWT()
	if err != nil {
		fmt.Printf("Error in token generation: %s", err)
	}

	return GenerateJSONString(identity, jwtToken)
}

func main() {
	// Create a new logger
	logger := log.New(log.Writer(), "[SERVER] ", log.LstdFlags)

	// Get the port from the environment variable, defaulting to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Define the HTTP handler function
	token := func(w http.ResponseWriter, r *http.Request) {
		// Log the request method and path
		logger.Printf("Received %s request for %s", r.Method, r.URL.Path)

		// Get the parameters for the token generation
		queryParams := r.URL.Query()
		identity := "anonymous"
		if len(queryParams["identity"]) > 0 {
			identity = queryParams["identity"][0]
		}

		room := "room"
		if len(queryParams["roomName"]) > 0 {
			room = queryParams["roomName"][0]
		}

		key := os.Getenv("APIKEY")
		if key == "" {
			key = "key"
		}

		secret := os.Getenv("APISECRET")
		if secret == "" {
			secret = "key"
		}

		// Generate join JWT token
		jointoken, err := getJoinToken(key, secret, room, identity)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Fprintf(w, "%v", jointoken)

	}

	// Register the handler function
	http.HandleFunc("/api/token", token)

	// Start the server
	addr := ":" + port
	logger.Printf("Server started on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
