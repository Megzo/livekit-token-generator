package main

import (
    "fmt"
	"encoding/base64"
	"encoding/json"
	"log"
	"os"
    "net/http"
	"strings"
	"time"

	//lksdk "github.com/livekit/server-sdk-go"
	"github.com/livekit/protocol/auth"
)

// Decodes a JWT token into a JSON string
func DecodeJWT(token string) (string, error) {
	// Split the token into its three parts: header, payload, and signature
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid JWT token format")
	}

	// Decode and parse the payload part
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode JWT payload: %v", err)
	}

	var payloadMap map[string]interface{}
	err = json.Unmarshal(payloadBytes, &payloadMap)
	if err != nil {
		return "", fmt.Errorf("failed to parse JWT payload: %v", err)
	}

	// Encode the payload as a JSON string
	payloadJSON, err := json.MarshalIndent(payloadMap, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to encode JWT payload as JSON: %v", err)
	}

	return string(payloadJSON), nil
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

	return DecodeJWT(jwtToken)
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