package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type DBFile struct {
	Id           int64
	FileSystemId int64
	ParentFileId int64
	FileName     string
	TotalSize    int64
	IsDirectory  bool
}

type DirectoryData struct {
	Dir   DBFile
	Files []DBFile
}

func runSQL(conn *sql.DB, sql string) {
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
	runSQL(conn, filesystemsTable)
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
	runSQL(conn, filesTable)
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
	s := `INSERT OR IGNORE INTO files(fileSystemId, parentFileId, fileName, totalSize, isDirectory) VALUES (?, ?, ?, ?, ?);`
	statement, err := conn.Prepare(s)
	if err != nil {
		log.Fatal(err.Error())
	}
	_, er := statement.Exec(fileSystemId, parentFileId, name, totalSize, isDirectory)
	statement.Close()
	if er != nil {
		log.Fatal(er.Error())
	}
	s2 := `select (id) from files where fileSystemId=? and parentFileId=? and fileName=?`
	statement2, _ := conn.Prepare(s2)
	rows, _ := statement2.Query(fileSystemId, parentFileId, name)
	defer statement2.Close()
	var result int64
	rows.Next()
	rows.Scan(&result)
	rows.Close()
	return result
}

func IncrementFileTotalSize(conn *sql.DB, fileId int64, totalSize int64) {
	// TODO - make this actually increment things
	sql := `update files set totalSize=? where id=?;`
	s, _ := conn.Prepare(sql)
	defer s.Close()
	_, e := s.Exec(totalSize, fileId)
	if e != nil {
		log.Fatal(e.Error())
	}
}

func LoadFile(conn *sql.DB, fileId int64) DBFile {
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
	results := make([]DBFile, 1)
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
		results[0] = DBFile{
			id,
			fileSystemId,
			parentFileId,
			fileName,
			totalSize,
			isDir,
		}
	}
	return results[0]
}

func LoadDirectoryData(conn *sql.DB, directoryId int64) DirectoryData {
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
	results := make([]DBFile, 0)
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
		results = append(results, DBFile{
			id,
			fileSystemId,
			parentFileId,
			fileName,
			totalSize,
			isDir,
		})
	}
	return DirectoryData{
		Dir:   LoadFile(conn, directoryId),
		Files: results,
	}
}

func GetDB(dataSourceName string) *sql.DB {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
