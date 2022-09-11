package main

import (
	"archive/tar"
	"context"
	"database/sql"
	"encoding/json"
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

type DirTuple struct {
	parentId int64
	fileName string
}

func loadFileSystemDataFromTarReader(reader *tar.Reader, db *sql.DB, fileSystemName string) {

	fileSystemId := SaveFileSystem(db, fileSystemName)
	rootDirectoryId := SaveFile(db, fileSystemId, -1, "/", 0, true)

	setOfExistingDirs := make(map[DirTuple]int64, 0)

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

		// Save all the directories from the file path - we are maintaining the
		//in-memory set of records to prevent us from re-saving files.
		parentDirectoryId := rootDirectoryId
		for _, dirName := range dirPathSegments {

			// one way or another this DirTuple is ending up in setOfExistingDirs
			keyForDirExistenceSet := DirTuple{parentDirectoryId, dirName}

			// check if it's already in there
			_, present := setOfExistingDirs[keyForDirExistenceSet]

			if !present {
				// if it's not in there yet, we need to hit the DB to save a record of it
				dirFileId := SaveFile(db, fileSystemId, parentDirectoryId, dirName, 0, true)
				setOfExistingDirs[keyForDirExistenceSet] = dirFileId
			}

			// now this dir is definitely in the DB and our in-memory set
			// pull out its ID and push it onto the stack in case we need to
			// use the stack to bubble up a file's size to its parents
			dirId, _ := setOfExistingDirs[keyForDirExistenceSet]
			stack.Push(dirId)

			// also update the parentDirectoryId variable so subsequent iterations,
			// if any, correctly process directories as children of this iteration's directory
			parentDirectoryId = dirId
		}

		if header.Typeflag == tar.TypeReg {

			// if this header is actually for a file, save a record of that file, too
			SaveFile(db, fileSystemId, parentDirectoryId, header.FileInfo().Name(), header.FileInfo().Size(), false)

			// Update all of this file's parent directories' total sizes with this file's size
			dirId, stackIsEmpty := stack.Pop()
			for ; !stackIsEmpty; dirId, stackIsEmpty = stack.Pop() {
				IncrementFileSize(db, dirId, header.FileInfo().Size())
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

type HistoryEntry struct {
	Created     string
	Created_by  string
	empty_layer bool
}

type Config struct {
	History []HistoryEntry
}

func loadFileSystemDataFromLayers(cli *client.Client, db *sql.DB, imageId string) {
	resp, err := cli.ImageSave(context.Background(), []string{imageId})
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Close()
	reader := tar.NewReader(resp)
	var imageConfig Config
	for true {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if strings.HasSuffix(header.Name, ".tar") {
			layerId := strings.Split(header.Name, "/")[0]
			loadFileSystemDataFromTarReader(tar.NewReader(reader), db, layerId)
		} else if header.Name != "manifest.json" && strings.HasSuffix(header.Name, ".json") {
			_bytes := make([]byte, header.Size)
			_, err := reader.Read(_bytes)
			if err != io.EOF {
				log.Fatalln(err.Error())
			}
			e := json.Unmarshal(_bytes, &imageConfig)
			if e != nil {
				return
			}
			history := make([]HistoryEntry, 0)
			for _, h := range imageConfig.History {
				if !h.empty_layer {
					history = append(history, h)
				}
			}
			imageConfig.History = history
		}
	}
	layers := LoadLayers(db)
	for i, layer := range layers {
		history := imageConfig.History[i]
		SetFileSystemCommand(db, layer.Id, history.Created_by)
	}
}
