package main

import (
	"database/sql"
	"github.com/docker/docker/client"
	"os"
)

func getImageNameFromCLI() string {
	if len(os.Args) < 2 {
		panic("Pass the Docker image you wish to analyze as the command-line argument.")
	}
	return os.Args[1]
}

func loadImageData(imageId string, db *sql.DB, cli *client.Client, done chan<- bool) {
	loadFileSystemDataFromImage(cli, db, imageId)
	done <- true
}

func loadLayerData(imageId string, db *sql.DB, cli *client.Client, done chan<- bool) {
	loadFileSystemDataFromLayers(cli, db, imageId)
	done <- true
}

func parseDockerImageIntoDB(imageId string, db *sql.DB, cli *client.Client) {
	imageDone := make(chan bool, 1)
	go loadImageData(imageId, db, cli, imageDone)
	layerDone := make(chan bool, 1)
	go loadLayerData(imageId, db, cli, layerDone)
	<-imageDone
	<-layerDone
}

func main() {
	//cli := getDockerClient()
	//imageName := getImageNameFromCLI()
	//im := getImageIdFromImageName(cli, imageName)
	db := GetDB("alpine.db")
	//CreateTables(db)
	//parseDockerImageIntoDB(im, db, cli)
	serveWebApp("alpine:latest", db)
}
