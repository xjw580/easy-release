package main

import (
	_ "easy-release/gui"
	_ "embed"
)

//go:embed favicon.ico
var embeddedFile []byte

func main() {

}

//func main() {
//	// 保存嵌入的文件到磁盘
//	err := saveEmbeddedFile("embedded_favicon.ico", embeddedFile)
//	if err != nil {
//		fmt.Println("Error saving embedded file:", err)
//		return
//	}
//
//	fmt.Println("Embedded file saved successfully.")
//}
//
//func saveEmbeddedFile(filename string, data []byte) error {
//	file, err := os.Create(filename)
//	if err != nil {
//		return err
//	}
//	defer file.Close()
//
//	_, err = file.Write(data)
//	return err
//}
