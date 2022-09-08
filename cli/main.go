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

func parseDockerImageIntoDB(imageId string, db *sql.DB, cli *client.Client) {
	loadFileSystemDataFromImage(cli, db, imageId)
	loadFileSystemDataFromLayers(cli, db, imageId)
}

func serveWebApp(db *sql.DB) {
	// TODO - API + static React app
}

func main() {
	cli := getDockerClient()
	im := getImageIdFromImageName(cli, getImageNameFromCLI())
	db := GetDB("test.db")
	//db := GetDB(":memory:")
	CreateTables(db)
	parseDockerImageIntoDB(im, db, cli)
	serveWebApp(db)
}
