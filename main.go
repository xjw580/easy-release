package main

import (
	"easy-release/gui"
	"embed"
	_ "embed"
	"log"
	"os"
)

//go:embed easy-release_static/*
var content embed.FS

func main() {
	saveAll()
	gui.ShowMain()
}

func createFile(path string, data []byte) {
	_, err2 := os.Open(path)
	if err2 != nil {
		err := saveEmbeddedFile(path, data)
		log.Println(err)
	}
}
func saveAll() {
	file, _ := content.ReadFile("easy-release_static/favicon.ico")
	file1, _ := content.ReadFile("easy-release_static/loading.png")
	file2, _ := content.ReadFile("easy-release_static/ok.png")
	file3, _ := content.ReadFile("easy-release_static/fail.png")
	createFile("easy-release_static/favicon.ico", file)
	createFile("easy-release_static/loading.png", file1)
	createFile("easy-release_static/ok.png", file2)
	createFile("easy-release_static/fail.png", file3)
}

func saveEmbeddedFile(filename string, data []byte) error {
	_ = os.Mkdir("easy-release_static", os.ModePerm)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}
