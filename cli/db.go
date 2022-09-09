package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type FileSystemEntry struct {
	Id    int64
	Name  string
	Size  int64
	IsDir bool
	Files []FileSystemEntry
}

type Layer struct {
	Id              int64
	RootDirectoryId int64
	Name            string
}

func GetDB(dataSourceName string) *sql.DB {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func _runCreate(conn *sql.DB, sql string) {
	statement, err := conn.Prepare(sql)
	defer statement.Close()
	if err != nil {
		log.Fatal(err.Error())
	}
	_, er := statement.Exec()
	if er != nil {
		log.Fatal(er.Error())
	}
}

func CreateTables(conn *sql.DB) {
	filesystemsTable := `CREATE TABLE filesystems (
    	"id" integer not null unique primary key autoincrement,
    	"name" text
    );`
	_runCreate(conn, filesystemsTable)
	filesTable := `CREATE TABLE files (
    	"id" integer not null unique primary key autoincrement,
		"fileSystemId" integer not null,
		"parentFileId" integer,
		"fileName" text not null,
		"totalSize" integer not null,
		"isDirectory" integer not null,
		UNIQUE (fileSystemId, parentFileId, fileName),
		FOREIGN KEY(parentFileId) REFERENCES files(id),
		FOREIGN KEY(fileSystemId) REFERENCES filesystems(id)
	  );`
	_runCreate(conn, filesTable)
}

func SaveFileSystem(conn *sql.DB, name string) int64 {
	s := `INSERT INTO filesystems(name) values (?)`
	statement, err := conn.Prepare(s)
	defer statement.Close()
	if err != nil {
		log.Fatalln(err)
	}
	rs, er := statement.Exec(name)
	if er != nil {
		log.Fatal(er.Error())
	}
	newFileSystemId, e := rs.LastInsertId()
	if e != nil {
		log.Fatal(e.Error())
	}
	return newFileSystemId
}

func SaveFile(conn *sql.DB, fileSystemId int64, parentFileId int64, name string, totalSize int64, isDirectory bool) int64 {
	s := `INSERT INTO files(fileSystemId, parentFileId, fileName, totalSize, isDirectory) VALUES (?, ?, ?, ?, ?) returning id;`
	statement, err := conn.Prepare(s)
	if err != nil {
		log.Fatal(err.Error())
	}
	rows, er := statement.Query(fileSystemId, parentFileId, name, totalSize, isDirectory)
	defer statement.Close()
	if er != nil {
		log.Fatal(er.Error())
	}
	var result int64
	rows.Next()
	rows.Scan(&result)
	rows.Close()
	return result
}

func _loadFile(conn *sql.DB, fileId int64) FileSystemEntry {
	s := `select id, fileSystemId, parentFileId, fileName, totalSize, isDirectory from files where id=?;`
	statement, err := conn.Prepare(s)
	defer statement.Close()
	if err != nil {
		log.Fatal(err.Error())
	}
	rows, e := statement.Query(fileId)
	if e != nil {
		log.Fatal(e.Error())
	}
	results := make([]FileSystemEntry, 1)
	for rows.Next() {
		var id, fileSystemId, parentFileId, totalSize, isDirectory int64
		var fileName string
		err = rows.Scan(&id, &fileSystemId, &parentFileId, &fileName, &totalSize, &isDirectory)
		if err != nil {
			log.Fatal(err)
		}
		isDir := false
		if isDirectory > 0 {
			isDir = true
		}
		results[0] = FileSystemEntry{
			id,
			fileName,
			totalSize,
			isDir,
			nil,
		}
	}
	return results[0]
}

func LoadDirectory(conn *sql.DB, directoryId int64) FileSystemEntry {
	s := `select id, fileSystemId, parentFileId, fileName, totalSize, isDirectory from files where parentFileId=?;`
	statement, err := conn.Prepare(s)
	defer statement.Close()
	if err != nil {
		log.Fatal(err.Error())
	}
	rows, e := statement.Query(directoryId)
	if e != nil {
		log.Fatal(e.Error())
	}
	results := make([]FileSystemEntry, 0)
	for rows.Next() {
		var id, fileSystemId, parentFileId, totalSize, isDirectory int64
		var fileName string
		err = rows.Scan(&id, &fileSystemId, &parentFileId, &fileName, &totalSize, &isDirectory)
		if err != nil {
			log.Fatal(err)
		}
		isDir := false
		if isDirectory > 0 {
			isDir = true
		}
		results = append(results, FileSystemEntry{
			id,
			fileName,
			totalSize,
			isDir,
			nil,
		})
	}
	d := _loadFile(conn, directoryId)
	d.Files = results
	return d
}

func LoadLayers(conn *sql.DB) []Layer {
	s := `select 
		filesystems.id as fileSystemId,
		files.id as rootDirectoryId,
		filesystems.name as fileSystemName
	from filesystems
		inner join files on files.fileSystemId = filesystems.Id where files.parentFileId=-1;`
	st, err := conn.Prepare(s)
	defer st.Close()
	if err != nil {
		log.Fatal(err.Error())
	}
	rows, error := st.Query()
	if error != nil {
		log.Fatal(error.Error())
	}
	results := make([]Layer, 0)
	for rows.Next() {
		var fileSystemId, fileId int64
		var fileSystemName string
		err = rows.Scan(&fileSystemId, &fileId, &fileSystemName)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, Layer{
			fileSystemId,
			fileId,
			fileSystemName,
		})
	}
	return results
}
