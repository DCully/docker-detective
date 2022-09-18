package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strconv"
)

//go:embed build
var UI embed.FS

func writeHttpResponse(w *http.ResponseWriter, responseCode int, byteStr []byte) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET")
	(*w).WriteHeader(responseCode)
	_, err := fmt.Fprintf(*w, string(byteStr))
	if err != nil {
		log.Println("An error occurred writing the 200 response - ", err.Error())
	}
}

func getStaticFS() http.FileSystem {
	fileSystem, err := fs.Sub(UI, "build")
	if err != nil {
		log.Fatalln("Error loading static web app FS:", err.Error())
	}
	return http.FS(fileSystem)
}

func serveWebApp(imageName string, db *sql.DB) {
	http.Handle("/", http.FileServer(getStaticFS()))
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
