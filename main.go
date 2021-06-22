package main

import (
	"log"
	"net/http"
)

func home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Write([]byte("Hello from SuiteNet"))
}

func showMaintenanceRequest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Display a specific maintenance request..."))
}

func createMaintenanceRequest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Create a new maintenance request..."))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", home)
	mux.HandleFunc("/maintenanceRequest", showMaintenanceRequest)
	mux.HandleFunc("/maintenanceRequest/create", createMaintenanceRequest)

	log.Println("Starting server on :4000")
	err := http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}
