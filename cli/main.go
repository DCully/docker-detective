package main

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"io"
	"log"
	"os"
	"strings"
)

type FileSystemEntry struct {
	Name     string             `json:"name"`
	Value    int64              `json:"value,omitempty"`
	Children []*FileSystemEntry `json:"children,omitempty"`
	path     string
}

func getImageNameFromCLI() string {
	if len(os.Args) < 2 {
		panic("Pass the Docker image you wish to analyze as the command-line argument.")
	}
	return os.Args[1]
}

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

func printMarshaledJSONString(o any) {
	// TODO - need to make this stream, instead of doing it all in memory
	j, err := json.Marshal(o)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(j))
}

func contains(children []*FileSystemEntry, childName string) (*FileSystemEntry, bool) {
	for _, c := range children {
		if c.Name == childName {
			return c, true
		}
	}
	return nil, false
}

func getFileSystemAsRootfsFromTarReader(reader *tar.Reader) FileSystemEntry {
	// This reference to top-level root node of the filesystem tree does not change.
	root := FileSystemEntry{Name: "/", Children: make([]*FileSystemEntry, 0)}

	// This preorder traverses the filesystem tree depth-first.
	// On this pass, we just add the directories into our tree.
	files := make([]*FileSystemEntry, 0)
	for true {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Container extraction failed: %s", err.Error())
		}
		if header.Typeflag == tar.TypeReg {
			files = append(files, &FileSystemEntry{
				Name:  header.FileInfo().Name(),
				Value: header.Size,
				path:  header.Name,
			})
		}
		if header.Typeflag != tar.TypeDir {
			// If it's not a file or a directory, just ignore it.
			continue
		}
		dirNames := strings.Split(strings.TrimSuffix(header.Name, "/"), "/")
		parent := &root
		for _, dirName := range dirNames {
			dir, present := contains(parent.Children, dirName)
			if !present {
				dir = &FileSystemEntry{
					Name:     dirName,
					Children: make([]*FileSystemEntry, 0),
				}
			}
			parent.Children = append(parent.Children, dir)
			parent = dir
		}
	}

	// Now add the files into the tree.
	for _, file := range files {
		dirNames := strings.Split(file.path, "/")
		dirForFile := &root
		for i := 0; i < len(dirNames)-1; i++ {
			dirForFile, _ = contains(dirForFile.Children, dirNames[i])
		}
		dirForFile.Children = append(dirForFile.Children, file)
	}
	return root
}

func getImageData(cli *client.Client, imageId string) FileSystemEntry {
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
	return getFileSystemAsRootfsFromTarReader(reader)
}

func readBytesAsJsonObject(reader *tar.Reader, numBytes int64) map[string]any {
	bytes := make([]byte, numBytes)
	_, err := reader.Read(bytes)
	if err != io.EOF {
		log.Fatalln(err)
	}
	jsonObj := make(map[string]any)
	err = json.Unmarshal(bytes, &jsonObj)
	if err != nil {
		log.Fatalln(err)
	}
	return jsonObj
}

func readBytesAsJsonArray(reader *tar.Reader, numBytes int64) []any {
	bytes := make([]byte, numBytes)
	_, err := reader.Read(bytes)
	if err != io.EOF {
		log.Fatalln(err)
	}
	jsonArr := make([]any, 0)
	err = json.Unmarshal(bytes, &jsonArr)
	if err != nil {
		log.Fatalln(err)
	}
	return jsonArr
}

func getLayerData(cli *client.Client, imageId string) map[string]any {
	resp, err := cli.ImageSave(context.Background(), []string{imageId})
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Close()
	reader := tar.NewReader(resp)
	layerIdsToRootFSs := make(map[string]FileSystemEntry, 0)
	manifest := make([]any, 0)
	config := make(map[string]any)
	for true {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if strings.HasSuffix(header.Name, ".tar") {
			layerReader := tar.NewReader(reader)
			layerRootDirectory := getFileSystemAsRootfsFromTarReader(layerReader)
			layerId := strings.Split(header.Name, "/")[0]
			layerIdsToRootFSs[layerId] = layerRootDirectory
		} else if header.Name == "manifest.json" {
			manifest = readBytesAsJsonArray(reader, header.Size)
		} else if strings.HasSuffix(header.Name, ".json") {
			config = readBytesAsJsonObject(reader, header.Size)
		}
	}
	result := make(map[string]any)
	result["layerIdsToRootFSs"] = layerIdsToRootFSs
	result["manifest"] = manifest
	result["config"] = config
	return result
}

func main() {
	cli := getDockerClient()
	imageName := getImageNameFromCLI()
	imageId := getImageIdFromImageName(cli, imageName)
	//layerData := getLayerData(cli, imageId)
	imageData := getImageData(cli, imageId)
	//data := make(map[string]any)
	//data["layers"] = layerData
	//data["image"] = imageData
	printMarshaledJSONString(imageData)
}
