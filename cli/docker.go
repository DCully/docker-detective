package main

import (
	"archive/tar"
	"context"
	"database/sql"
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
	rootDirectoryFileId := SaveFile(db, fileSystemId, -1, "/", 0, true)
	parentId := rootDirectoryFileId
	for true {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Container extraction failed: %s", err.Error())
		}
		if header.Typeflag == tar.TypeReg {
			SaveFile(db, fileSystemId, parentId, header.FileInfo().Name(), header.Size, false)
		}
		if header.Typeflag == tar.TypeDir {
			SaveFile(db, fileSystemId, parentId, header.FileInfo().Name(), 0, true)
		}
		// TODO - update parent reference
	}
	// TODO - update size records on directory entries
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
