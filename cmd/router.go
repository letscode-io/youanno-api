package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	timecodeParser "youannoapi/cmd/timecode_parser"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func startHttpServer() {
	log.Println("Starting development server at http://127.0.0.1:8080/")
	router := mux.NewRouter().StrictSlash(true)

	handler := cors.Default().Handler(router)

	router.HandleFunc("/", handleHome)

	// annotations
	router.HandleFunc("/annotations", createAnnotation).Methods("POST")
	router.HandleFunc("/annotations/{videoId}", getAnnotations)

	router.HandleFunc("/parse_description/{videoId}", parseDescription)

	log.Fatal(http.ListenAndServe(":8080", handler))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func parseDescription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoId := vars["videoId"]

	description := getVideoDescription(videoId)
	timeCodes := timecodeParser.Parse(description)

	json.NewEncoder(w).Encode(timeCodes)
}
