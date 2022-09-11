package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func serveWebApp(imageName string, db *sql.DB) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "todo - serve the front end statically here")
	})
	http.HandleFunc("/name", func(w http.ResponseWriter, r *http.Request) {
		o := make(map[string]string)
		o["imageName"] = imageName
		j, _ := json.Marshal(o)
		w.WriteHeader(200)
		fmt.Fprintf(w, string(j))
	})
	http.HandleFunc("/layers", func(w http.ResponseWriter, r *http.Request) {
		// the 'image' layer can be identified by name, and the other layers
		// can be ordered by sorting their root file system IDs
		j, _ := json.Marshal(LoadLayers(db))
		w.WriteHeader(200)
		fmt.Fprintf(w, string(j))
	})
	http.HandleFunc("/dirData", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(r.URL.Query().Get("id"))
		jsonDirData, _ := json.Marshal(LoadDirectory(db, int64(id)))
		w.WriteHeader(200)
		fmt.Fprintf(w, string(jsonDirData))
	})
	log.Fatal(http.ListenAndServe(":1337", nil))
}
