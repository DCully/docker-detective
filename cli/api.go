package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func enableCorsAndGet(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET")
}

func serveWebApp(imageName string, db *sql.DB) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		enableCorsAndGet(&w)
		fmt.Fprintf(w, "todo - serve the front end statically here")
	})
	http.HandleFunc("/name", func(w http.ResponseWriter, r *http.Request) {
		enableCorsAndGet(&w)
		o := make(map[string]string)
		o["imageName"] = imageName
		j, _ := json.Marshal(o)
		w.WriteHeader(200)
		fmt.Fprintf(w, string(j))
	})
	http.HandleFunc("/filesystems", func(w http.ResponseWriter, r *http.Request) {
		enableCorsAndGet(&w)
		// the 'image' layer can be identified by name, and the other layers
		// can be ordered by sorting their root file system IDs
		j, _ := json.Marshal(LoadLayers(db))
		w.WriteHeader(200)
		fmt.Fprintf(w, string(j))
	})
	http.HandleFunc("/dirData", func(w http.ResponseWriter, r *http.Request) {
		enableCorsAndGet(&w)
		id, _ := strconv.Atoi(r.URL.Query().Get("id"))
		jsonDirData, _ := json.Marshal(LoadDirectory(db, int64(id)))
		w.WriteHeader(200)
		fmt.Fprintf(w, string(jsonDirData))
	})
	log.Fatal(http.ListenAndServe(":1337", nil))
}
