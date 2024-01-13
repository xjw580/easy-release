package gui

import (
	"bufio"
	"context"
	"fmt"
	"github.com/go-toast/toast"
	"github.com/google/go-github/v58/github"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"golang.org/x/oauth2"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const (
	ProgramName = "easy-release"
)

var MyWindow *walk.MainWindow
var LogTextEdit *walk.TextEdit
var ExecBtn *walk.PushButton
var (
	ReleaseGithub *walk.CheckBox
	ReviewContent *walk.CheckBox
	ReleaseGitee  *walk.CheckBox
)

func init() {
	image, _ := walk.Resources.Image("favicon.ico")
	run, err := MainWindow{
		AssignTo: &MyWindow,
		Title:    ProgramName,
		Bounds: Rectangle{
			X:      1000,
			Y:      100,
			Width:  800,
			Height: 1000,
		},
		Icon: image,
		Font: Font{
			PointSize: 11,
		},
		Layout: VBox{},
		Background: SolidColorBrush{
			Color: walk.RGB(224, 240, 253),
		},
		Children: []Widget{
			Composite{
				Layout: HBox{Alignment: AlignHNearVNear},
				MaxSize: Size{
					Height: 110,
				},
				Background: SolidColorBrush{
					Color: walk.RGB(200, 200, 200),
				},
				Children: []Widget{
					Composite{
						Alignment: AlignHNearVNear,
						Layout:    HBox{},
						Children: []Widget{
							CheckBox{
								AssignTo: &ReviewContent,
								Text:     "审查版本说明",
								Checked:  true,
							},
						},
					},
					Composite{
						Background: SolidColorBrush{
							Color: walk.RGB(220, 220, 220),
						},
						Alignment: AlignHNearVNear,
						Layout:    HBox{},
						Children: []Widget{
							Label{
								Text: "发布平台：",
							},
							CheckBox{
								AssignTo: &ReleaseGitee,
								Text:     "Gitee",
								Checked:  true,
							},
							CheckBox{
								AssignTo: &ReleaseGithub,
								Text:     "Github",
								Checked:  true,
							},
						},
					},
					Composite{
						Alignment: AlignHNearVNear,
						Background: SolidColorBrush{
							Color: walk.RGB(220, 220, 220),
						},
						Layout: HBox{},
						Children: []Widget{
							Label{
								Text: "日志：",
							},
						},
					},
				},
			},
			TextEdit{
				AssignTo: &LogTextEdit,
				ReadOnly: false,
				VScroll:  true,
				HScroll:  true,
			},
			Composite{
				MaxSize: Size{
					Height: 50,
				},
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						MaxSize: Size{
							Width: 50,
						},
						AssignTo: &ExecBtn,
						Text:     "执行",
						OnClicked: func() {
							go func() {
								ExecBtn.SetEnabled(false)
								LogToGUI("开始执行")

								if ReleaseGitee.Checked() {
									PushGitee()
								}

								if ReleaseGithub.Checked() {
									PushGithub()
								}

								PackageProject()

								ReleasePackage()

								LogToGUI("执行完毕")
								ExecBtn.SetEnabled(true)
							}()
						},
						Background: SolidColorBrush{
							Color: walk.RGB(128, 206, 128),
						},
					},
				},
			},
		},
	}.Run()
	log.Println("exit code:", run)
	if err != nil {
		log.Println(err)
		return
	}
}

const (
	// 设置 GitHub 仓库信息
	OWNER     = "xjw580"
	REPO_NAME = "Hearthstone-Script"
	// 设置 GitHub 访问令牌
	TOKEN        = "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"
	PACKAGE_NAME = "hs-script"
)

func notice(content string) {
	notification := toast.Notification{
		AppID:   "Microsoft.Windows.Shell.RunDialog",
		Title:   ProgramName,
		Message: content,
	}
	err := notification.Push()
	if err != nil {
		log.Fatalln(err)
	}
	LogToGUI(content)
}

func PushGithub() {
	LogToGUI("开始Push到Github")
	push("Github", "master", "master")
}

func PushGitee() {
	LogToGUI("开始Push到Gitee")
	push("Gitee", "master", "master")
}

func push(origin, localBranch, remoteBranch string) {
	err := execCMD("git", "push", origin, localBranch+":"+remoteBranch)
	if err == nil {
		notice("推送到" + origin + "成功")
	} else {
		notice("推送到" + origin + "失败\n" + err.Error())
	}
}

func PackageProject() {
	LogToGUI("开始打包")
	err := execCMD("mvn", "clean", "package")
	if err == nil {
		notice("打包成功")
	} else {
		notice("打包失败")
	}
}

func ReleasePackage() {
	LogToGUI("开始发布")
	commitMessage, _ := getLatestCommitMessage()
	releaseTag, prerelease := parseVersionAndPreRelease(commitMessage)
	commitMessage = "#### " + commitMessage

	LogToGUI("commitMessage：" + commitMessage)
	LogToGUI("releaseTag：" + releaseTag)
	LogToGUI("prerelease：" + strconv.FormatBool(prerelease))

	packageFile := getPackageFile()

	// 创建 GitHub 客户端
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: TOKEN})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// 获取已有的 Release，如果不存在则创建
	release, _, err := client.Repositories.GetReleaseByTag(ctx, OWNER, REPO_NAME, releaseTag)
	if err != nil {
		fmt.Printf("Error getting releasePackage: %v\n", err)

		// 创建新的 Release
		createRelease := &github.RepositoryRelease{
			TagName:         github.String(releaseTag),
			Name:            github.String(PACKAGE_NAME + "_" + releaseTag),
			TargetCommitish: github.String("master"), // 或者你的默认分支
			Body:            github.String(commitMessage),
			//Draft:           github.Bool(true),
			Prerelease: github.Bool(prerelease),
		}

		release, _, err = client.Repositories.CreateRelease(ctx, OWNER, REPO_NAME, createRelease)
		if err != nil {
			fmt.Printf("Error creating releasePackage: %v\n", err)
			notice("创建release失败")
			return
		}
		notice("创建release成功")
	}

	for _, s := range packageFile {
		uploadFile(s, client, ctx, release)
	}

}

func uploadFile(path string, client *github.Client, ctx context.Context, release *github.RepositoryRelease) {
	// 打开本地文件
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()
	fileName := filepath.Base(file.Name())
	LogToGUI("fileName：" + fileName)
	// 创建上传选项
	opts := &github.UploadOptions{Name: fileName}

	// 上传文件到 GitHub Release
	asset, _, err := client.Repositories.UploadReleaseAsset(ctx, OWNER, REPO_NAME, release.GetID(), opts, file)
	if err != nil {
		fmt.Printf("Error uploading file: %v\n", err)
		notice(fileName + "上传失败")
		return
	}

	fmt.Printf("File uploaded successfully! Asset ID: %d\n", asset.GetID())
	notice(fileName + "上传成功")
}

func getPackageFile() []string {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		LogToGUI("Error getting current directory:" + err.Error())
		return nil
	}

	// 构建 target 目录的完整路径
	targetDir := filepath.Join(currentDir, "target")

	// 获取 target 目录下的所有文件
	files, err := getFilesInDirectory(targetDir, []string{".zip", ".jar"})
	if err != nil {
		LogToGUI("Error getting files:" + err.Error())
		return nil
	}
	return files
}

func execCMD(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	// 启动命令
	err = cmd.Start()
	if err != nil {
		LogToGUI(err.Error())
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// 启动 goroutine 实时读取输出
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			LogToGUI(scanner.Text())
		}
	}()

	// 等待命令执行完成
	err = cmd.Wait()
	if err != nil {
		LogToGUI(err.Error())
		log.Fatal(err)
	}

	// 等待输出读取 goroutine 完成
	wg.Wait()
	return nil
}

func parseVersionAndPreRelease(commitMessage string) (string, bool) {
	lines := strings.Split(commitMessage, "\n")
	if len(lines) > 0 {
		// 解析版本号，假设版本号在提交信息的第一行
		version := strings.TrimSpace(lines[0])

		// 判断是否预发布
		prerelease := strings.HasSuffix(version, "DEV") || strings.HasSuffix(version, "BETA")

		return version, prerelease
	}

	// 如果没有版本号，默认使用 "v1.0.0" 并标记为非预发布
	return "v1.0.0", false
}

func getLatestCommitMessage() (string, error) {
	cmd := exec.Command("git", "log", "-1", "--pretty=%B")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getFilesInDirectory(dirPath string, extensions []string) ([]string, error) {
	var files []string

	// 遍历目录下的文件
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 检查文件是否是目标扩展名之一
		for _, ext := range extensions {
			if strings.HasSuffix(strings.ToLower(info.Name()), ext) {
				files = append(files, path)
				break
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func LogToGUI(log string) {
	LogTextEdit.AppendText(log + "\r\n")
}
