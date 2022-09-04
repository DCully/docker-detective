package main

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"io"
	"log"
	"os"
	"strings"
)

type File struct {
	Size int64
	Name string
	path string
}

type Directory struct {
	Files       []File
	Directories map[string]*Directory
}

func (directory *Directory) AddFile(file File) {
	directory.Files = append(directory.Files, file)
}

func getImageFromCLI() string {
	p := argparse.NewParser("docker-detective", "Visualize the contents of Docker images")
	i := p.String("i", "image",
		&argparse.Options{Required: true, Help: "The name of the (local) Docker image to inspect"})
	err := p.Parse(os.Args)
	if err != nil {
		log.Fatalln(p.Usage(err))
	}
	return *i
}

func getMarshaledJSONString(o any) string {
	j, err := json.Marshal(o)
	if err != nil {
		log.Fatalln(err)
	}
	return string(j)
}

func getFileSystemAsJSONFromTarReader(reader *tar.Reader) string {
	// This reference to top-level root node of the filesystem tree does not change.
	root := Directory{Files: make([]File, 0), Directories: make(map[string]*Directory, 0)}

	// This preorder traverses the filesystem tree depth-first.
	files := make([]File, 0)
	for true {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Container extraction failed: %s", err.Error())
		}
		if header.Typeflag == tar.TypeReg {
			files = append(files, File{path: header.Name, Name: header.FileInfo().Name(), Size: header.Size})
		}
		if header.Typeflag != tar.TypeDir {
			continue
		}
		dirNames := strings.Split(strings.TrimSuffix(header.Name, "/"), "/")
		parent := &root
		for _, dirName := range dirNames {
			dir := Directory{Files: make([]File, 0), Directories: make(map[string]*Directory, 0)}
			_, dirIsPresent := parent.Directories[dirName]
			if !dirIsPresent {
				parent.Directories[dirName] = &dir
			}
			parent = parent.Directories[dirName]
		}
	}
	for _, file := range files {
		dirNames := strings.Split(file.path, "/")
		dirForFile := &root
		for i := 0; i < len(dirNames)-1; i++ {
			dirForFile = dirForFile.Directories[dirNames[i]]
		}
		dirForFile.AddFile(file)
	}
	return getMarshaledJSONString(root)
}

func getImageFileSystemAsJSON(imageName string) string {
	// Get a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatalln(err)
	}

	// Create a container from the specified image
	c, err := cli.ContainerCreate(
		context.Background(),
		&container.Config{Image: imageName},
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
	return getFileSystemAsJSONFromTarReader(reader)
}

func main() {
	imageName := getImageFromCLI()
	j := getImageFileSystemAsJSON(imageName)
	fmt.Println(j)
}
