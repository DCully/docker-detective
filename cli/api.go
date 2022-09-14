package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func writeHttpResponse(w *http.ResponseWriter, responseCode int, byteStr []byte) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET")
	(*w).WriteHeader(responseCode)
	_, err := fmt.Fprintf(*w, string(byteStr))
	if err != nil {
		log.Println("An error occurred writing the 200 response - ", err.Error())
	}
}

func serveWebApp(imageName string, db *sql.DB) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeHttpResponse(&w, 200, []byte("todo - serve the front end statically here"))
	})
	http.HandleFunc("/name", func(w http.ResponseWriter, r *http.Request) {
		o := make(map[string]string)
		o["imageName"] = imageName
		j, _ := json.Marshal(o)
		writeHttpResponse(&w, 200, j)
	})
	http.HandleFunc("/filesystems", func(w http.ResponseWriter, r *http.Request) {
		j, _ := json.Marshal(LoadLayers(db))
		writeHttpResponse(&w, 200, j)
	})
	http.HandleFunc("/dirData", func(w http.ResponseWriter, r *http.Request) {
		id, atoiErr := strconv.Atoi(r.URL.Query().Get("id"))
		if atoiErr != nil {
			writeHttpResponse(&w, 400, []byte("Bad id"))
		}
		jsonDirData, _ := json.Marshal(LoadDirectory(db, int64(id)))
		writeHttpResponse(&w, 200, jsonDirData)
	})
	log.Fatal(http.ListenAndServe(":1337", nil))
}
