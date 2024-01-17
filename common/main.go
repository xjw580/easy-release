package common

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const (
	ProgramName          = "easy-release"
	ConfigFilePath       = "easy-release_static/config.json"
	GlobalConfigFilePath = "C:\\ProgramData\\" + ProgramName + "\\config.json"
)

type Logs interface {
	AppendLog(log string)
}
type Config struct {
	GiteeRepository  GitRepository
	GithubRepository GitRepository
}

type GitRepository struct {
	Owner    string `json:"name"`
	RepoName string `json:"repoName"`
	Token    string `json:"token"`
}

func ReadConfigFromFile() (Config, error) {
	var config Config
	_, err3 := os.Open(ConfigFilePath)
	if err3 != nil {
		_, err2 := createFile(GlobalConfigFilePath)
		if err2 == nil {
			_ = copyFile(GlobalConfigFilePath, ConfigFilePath)
		}
	}

	data, err := ioutil.ReadFile(ConfigFilePath)
	if err != nil {
		return config, err
	}
	config = Config{
		GithubRepository: GitRepository{},
		GiteeRepository:  GitRepository{},
	}
	err = json.Unmarshal(data, &config)
	return config, err
}
func createFile(filePath string) (bool, error) {
	open, err2 := os.Open(filePath)
	defer open.Close()
	if err2 != nil {
		// 创建文件所在目录，如果目录不存在的话
		err := os.MkdirAll(filepath.Dir(ConfigFilePath), os.ModePerm)
		if err != nil {
			log.Println("Error creating directory:", err)
			return false, err
		}

		// 创建文件
		file, err := os.Create(ConfigFilePath)
		if err != nil {
			log.Println("Error creating file:", err)
			return false, err
		}
		var config = Config{
			GithubRepository: GitRepository{
				Owner:    "",
				RepoName: "",
				Token:    "",
			},
			GiteeRepository: GitRepository{
				Owner:    "",
				RepoName: "",
				Token:    "",
			},
		}
		_ = WriteConfigToFile(config)
		defer file.Close()
		return true, nil
	}
	return false, nil
}
func WriteConfigToFile(config Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	ioutil.WriteFile(GlobalConfigFilePath, data, 0644)
	return ioutil.WriteFile(ConfigFilePath, data, 0644)
}

func copyFile(source, destination string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
