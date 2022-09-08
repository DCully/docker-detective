package main

import (
	"archive/tar"
	"context"
	"database/sql"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"io"
	"log"
	"strings"
)

func getDockerClient() *client.Client {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatalln(err)
	}
	return cli
}

func getImageIdFromImageName(cli *client.Client, imageName string) string {
	filters := filters.NewArgs()
	filters.Add("reference", imageName)
	opts := types.ImageListOptions{All: false, Filters: filters}
	imageIds, err := cli.ImageList(context.Background(), opts)
	if err != nil {
		log.Fatalln(err)
	}
	if len(imageIds) < 1 {
		log.Fatalln("Could not find image named " + imageName)
	}
	return imageIds[0].ID
}

func loadFileSystemDataFromTarReader(reader *tar.Reader, db *sql.DB, fileSystemName string) {

	fileSystemId := SaveFileSystem(db, fileSystemName)
	rootDirectoryId := SaveFile(db, fileSystemId, -1, "/", 0, true)

	for true {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Container extraction failed: %s", err.Error())
		}
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeDir {
			continue
		}

		stack := DirectoryIdStack{}
		stack.Push(rootDirectoryId)

		// First, make sure all directories in this file path are saved.
		dirPathSegments := strings.Split(header.Name, "/")

		// If this is a file, the last element is the file name.
		// If this is a directory, the last element is a blank string.
		// Either way, slice it off so that we don't save it as a directory record.
		dirPathSegments = dirPathSegments[:len(dirPathSegments)-1]

		// Save all the directories from the file path - this works because SaveFile is idempotent.
		parentFileId := rootDirectoryId
		for _, dirName := range dirPathSegments {
			parentFileId = SaveFile(db, fileSystemId, parentFileId, dirName, 0, true)
			stack.Push(parentFileId)
		}

		// Now, if this is a file, save a record for the file itself,
		// and bubble its file size up into all of its parent directory size sums.
		if header.Typeflag == tar.TypeReg {
			SaveFile(db, fileSystemId, parentFileId, header.FileInfo().Name(), header.FileInfo().Size(), false)
			for !stack.IsEmpty() {
				id, _ := stack.Pop()
				fmt.Println("id ", id, " name ", header.Name)
				IncrementFileTotalSize(db, id, header.FileInfo().Size())
			}
		}
	}
}

func loadFileSystemDataFromImage(cli *client.Client, db *sql.DB, imageId string) {
	// Create a container from the specified image
	c, err := cli.ContainerCreate(
		context.Background(),
		&container.Config{Image: imageId},
		nil,
		nil,
		nil,
		"")
	if err != nil {
		log.Fatalln(err)
	}
	defer cli.ContainerRemove(
		context.Background(),
		c.ID,
		types.ContainerRemoveOptions{RemoveVolumes: false, RemoveLinks: false, Force: true})

	// Get a tar archive stream of the container and convert the FS to JSON
	resp, err := cli.ContainerExport(context.Background(), c.ID)
	defer resp.Close()
	reader := tar.NewReader(resp)
	loadFileSystemDataFromTarReader(reader, db, "image")
}

func loadFileSystemDataFromLayers(cli *client.Client, db *sql.DB, imageId string) {
	resp, err := cli.ImageSave(context.Background(), []string{imageId})
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Close()
	reader := tar.NewReader(resp)
	for true {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if strings.HasSuffix(header.Name, ".tar") {
			layerId := strings.Split(header.Name, "/")[0]
			loadFileSystemDataFromTarReader(tar.NewReader(reader), db, layerId)
		}
	}
	// TODO parse/save the manifest, so we have layer ordering and names
}
