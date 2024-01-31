package api

import (
	"encoding/json"
	"log"
	"net/http"

	"go.woodpecker-ci.org/woodpecker/v2/server/queue" // replace with the actual import path
)

func GetQueueStats() {
	http.HandleFunc("/api/queue/stats", func(w http.ResponseWriter, r *http.Request) {
		var stats queue.InfoT
		err := json.NewEncoder(w).Encode(stats)
		if err != nil {
			// handle the error
			log.Printf("Error encoding JSON: %v", err)
			return
		}
	})
}
