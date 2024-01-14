package release

import (
	"bufio"
	"context"
	"easy-release/common"
	"fmt"
	"github.com/go-toast/toast"
	"github.com/google/go-github/v58/github"
	"golang.org/x/oauth2"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

func init() {
	giteePlatformMsg = &GitPlatformMsg{
		owner:    "zergqueen",
		repoName: "Hearthstone-Script",
		token:    "34142252982fbc15356ffa7407fe0d9f",
	}
	githubPlatformMsg = &GitPlatformMsg{
		owner:    "xjw580",
		repoName: "Hearthstone-Script",
		token:    "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",
	}
}

var (
	guiLogs           common.Logs
	giteePlatformMsg  *GitPlatformMsg
	githubPlatformMsg *GitPlatformMsg
)

type GitPlatform string

const (
	GiteePlatform  = "Gitee"
	GithubPlatform = "Github"
)

type GitPlatformMsg struct {
	owner, repoName, token string
}

type ProjectType interface {
	PushPlatform(gitPlatform ...GitPlatform)
	PackageProject()
	ReleasePackage(fileTypes, commitMessage, releaseTag string, gitPlatform ...GitPlatform)
}
type JavaMavenProject struct {
}
type GoProject struct {
	test string
}

func (project JavaMavenProject) PushPlatform(gitPlatform ...GitPlatform) {
	pushAll(gitPlatform...)
}
func (project JavaMavenProject) PackageProject() {
	guiLogs.AppendLog("++++++++++++++++++++开始打包++++++++++++++++++++")
	err := execCMD("mvn", "clean", "package")
	if err == nil {
		notice("打包成功")
	} else {
		notice("打包失败")
	}
}
func (project JavaMavenProject) ReleasePackage(fileTypes, commitMessage, releaseTag string, gitPlatform ...GitPlatform) {
	releaseAll(fileTypes, commitMessage, releaseTag, "target", gitPlatform...)
}
func releaseAll(fileTypes, commitMessage, releaseTag, packageDir string, gitPlatform ...GitPlatform) {
	for _, platform := range gitPlatform {
		switch platform {
		case GiteePlatform:
			releaseGitee(fileTypes, commitMessage, releaseTag, packageDir)
		case GithubPlatform:
			releaseGithub(fileTypes, commitMessage, releaseTag, packageDir)
		}
	}

}

func releaseGithub(fileTypes, commitMessage, releaseTag, packageDir string) {
	guiLogs.AppendLog("++++++++++++++++++++开始发布到Github++++++++++++++++++++")
	_, prerelease := ParseVersionAndPreRelease(commitMessage)
	commitMessage = "#### " + commitMessage

	guiLogs.AppendLog("prerelease：" + strconv.FormatBool(prerelease))

	packageFile := getJavaMavenPackageFile(fileTypes, packageDir)
	// 创建 GitHub 客户端
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubPlatformMsg.token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// 获取已有的 Release，如果不存在则创建
	release, _, err := client.Repositories.GetReleaseByTag(ctx, githubPlatformMsg.owner, githubPlatformMsg.repoName, releaseTag)
	if err != nil {
		fmt.Printf("Error getting releasePackage: %v\n", err)
		// 创建新的 Release
		createRelease := &github.RepositoryRelease{
			TagName:         github.String(releaseTag),
			Name:            github.String(packageFile[0] + "_" + releaseTag),
			TargetCommitish: github.String("master"), // 或者你的默认分支
			Body:            github.String(commitMessage),
			//Draft:           github.Bool(true),
			Prerelease: github.Bool(prerelease),
		}

		release, _, err = client.Repositories.CreateRelease(ctx, githubPlatformMsg.owner, githubPlatformMsg.repoName, createRelease)
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

func releaseGitee(fileTypes, commitMessage, releaseTag, packageDir string) {
	guiLogs.AppendLog("++++++++++++++++++++开始发布到Gitee++++++++++++++++++++")
}

func pushAll(gitPlatform ...GitPlatform) {
	for _, platform := range gitPlatform {
		switch platform {
		case GiteePlatform:
			guiLogs.AppendLog("++++++++++++++++++++开始Push到Gitee++++++++++++++++++++")
			push("Gitee", "master", "master")
		case GithubPlatform:
			guiLogs.AppendLog("++++++++++++++++++++开始Push到Github++++++++++++++++++++")
			push("Github", "master", "master")
		}
	}
}

func push(origin, localBranch, remoteBranch string) {
	err := execCMD("git", "push", origin, localBranch+":"+remoteBranch)
	if err == nil {
		notice("推送到" + origin + "成功")
	} else {
		notice("推送到" + origin + "失败,err：" + err.Error())
	}
}

func (project GoProject) PackageProject() {
	guiLogs.AppendLog("++++++++++++++++++++开始打包++++++++++++++++++++")
	err := execCMD("go", "build", "-ldflags", "-H=windowsgui")
	if err == nil {
		notice("打包成功")
	} else {
		notice("打包失败")
	}
}

func ReleasePackage() {

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
	guiLogs.AppendLog("fileName：" + fileName)
	// 创建上传选项
	opts := &github.UploadOptions{Name: fileName}

	// 上传文件到 GitHub Release
	asset, _, err := client.Repositories.UploadReleaseAsset(ctx, "", "", release.GetID(), opts, file)
	if err != nil {
		fmt.Printf("Error uploading file: %v\n", err)
		notice(fileName + "上传失败")
		return
	}

	fmt.Printf("File uploaded successfully! Asset ID: %d\n", asset.GetID())
	notice(fileName + "上传成功")
}

func getJavaMavenPackageFile(fileTypes, packageDir string) []string {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		guiLogs.AppendLog("Error getting current directory:" + err.Error())
		return nil
	}

	// 构建 target 目录的完整路径
	targetDir := filepath.Join(currentDir, packageDir)

	split := strings.Split(fileTypes, "&")
	for i := range split {
		split[i] = "." + split[i]
	}
	// 获取 target 目录下的所有文件
	files, err := getFilesInDirectory(targetDir, split)
	if err != nil {
		guiLogs.AppendLog("Error getting files:" + err.Error())
		return nil
	}
	return files
}

func execCMD(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		guiLogs.AppendLog(err.Error())
		return err
	}

	// 启动命令
	err = cmd.Start()
	if err != nil {
		guiLogs.AppendLog(err.Error())
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// 启动 goroutine 实时读取输出
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			guiLogs.AppendLog(scanner.Text())
		}
	}()

	// 等待命令执行完成
	err = cmd.Wait()
	if err != nil {
		guiLogs.AppendLog(err.Error())
		return err
	}

	// 等待输出读取 goroutine 完成
	wg.Wait()
	return nil
}

func ParseVersionAndPreRelease(commitMessage string) (string, bool) {
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

func GetLatestCommitMessage() (string, error) {
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

func RequireLogs(logs common.Logs) {
	guiLogs = logs
}

func notice(content string) {
	notification := toast.Notification{
		AppID:   "Microsoft.Windows.Shell.RunDialog",
		Title:   common.ProgramName,
		Message: content,
	}
	guiLogs.AppendLog("++++++++++++++++++++" + content + "++++++++++++++++++++")
	err := notification.Push()
	if err != nil {
		guiLogs.AppendLog(err.Error())
	}
}
