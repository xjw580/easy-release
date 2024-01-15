package release

import (
	"bufio"
	"context"
	"easy-release/common"
	"fmt"
	"github.com/go-resty/resty/v2"
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
	config, _ := common.ReadConfigFromFile()
	giteeRepository = &config.GiteeRepository
	githubRepository = &config.GithubRepository
}

var (
	guiLogs          common.Logs
	giteeRepository  *common.GitRepository
	githubRepository *common.GitRepository
)

type GitPlatform string

const (
	GiteePlatform  = "Gitee"
	GithubPlatform = "Github"
)

type ProjectType interface {
	PushPlatform(gitPlatform []GitPlatform)
	PackageProject()
	ReleasePackage(fileTypes []string, commitMessage, releaseVersion string, gitPlatform []GitPlatform)
}

type JavaMavenProject struct {
}

func (project JavaMavenProject) PushPlatform(gitPlatform []GitPlatform) {
	pushAll(gitPlatform)
}
func (project JavaMavenProject) PackageProject() {
	guiLogs.AppendLog("++++++++++++++++++++开始打包++++++++++++++++++++")
	err := execCMD("mvn", "clean", "package")
	if err == nil {
		guiLogs.AppendLog("++++++++++++++++++++打包成功")
	} else {
		guiLogs.AppendLog("打包失败,err:" + err.Error())
	}
}
func (project JavaMavenProject) ReleasePackage(fileTypes []string, commitMessage, releaseVersion string, gitPlatform []GitPlatform) {
	releaseAll(fileTypes, commitMessage, releaseVersion, "target", gitPlatform)
}

type GoProject struct {
}

func (project GoProject) PushPlatform(gitPlatform []GitPlatform) {
	pushAll(gitPlatform)
}
func (project GoProject) PackageProject() {
	guiLogs.AppendLog("++++++++++++++++++++开始打包++++++++++++++++++++")
	err := execCMD("go", "build", "-ldflags", "-H=windowsgui")
	if err == nil {
		guiLogs.AppendLog("++++++++++++++++++++打包成功")
	} else {
		guiLogs.AppendLog("打包失败,err:" + err.Error())
	}
}
func (project GoProject) ReleasePackage(fileTypes []string, commitMessage, releaseVersion string, gitPlatform []GitPlatform) {
	releaseAll(fileTypes, commitMessage, releaseVersion, "", gitPlatform)
}

func releaseAll(fileTypes []string, commitMessage, releaseTag, packageDir string, gitPlatform []GitPlatform) {
	for _, platform := range gitPlatform {
		switch platform {
		case GiteePlatform:
			releaseGitee(fileTypes, commitMessage, releaseTag, packageDir)
		case GithubPlatform:
			releaseGithub(fileTypes, commitMessage, releaseTag, packageDir)
		}
	}

}

/*
*
return tagName, name, body, prerelease, packageFile
*/
func getReleaseMsg(fileTypes []string, commitMessage, releaseVersion, packageDir string) (string, string, string, bool, []string) {
	_, prerelease := ParseVersionAndPreRelease(commitMessage)
	commitMessage = "#### " + commitMessage
	packageFile := getJavaMavenPackageFile(fileTypes, packageDir)
	fileName := filepath.Base(packageFile[0])
	index := strings.LastIndex(fileName, ".")
	return releaseVersion, fileName[:index], commitMessage, prerelease, packageFile
}

func releaseGithub(fileTypes []string, commitMessage, releaseVersion, packageDir string) {
	guiLogs.AppendLog("++++++++++++++++++++开始发布到Github++++++++++++++++++++")
	tagName, name, body, prerelease, packageFile := getReleaseMsg(fileTypes, commitMessage, releaseVersion, packageDir)
	// 创建 GitHub 客户端
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubRepository.Token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	// 获取已有的 Release，如果不存在则创建
	release, _, err := client.Repositories.GetReleaseByTag(ctx, githubRepository.Owner, githubRepository.RepoName, releaseVersion)
	// 创建新的 Release
	createRelease := &github.RepositoryRelease{
		TagName:         github.String(tagName),
		Name:            github.String(name),
		TargetCommitish: github.String("master"), // 或者你的默认分支
		Body:            github.String(body),
		//Draft:           github.Bool(true),
		Prerelease: github.Bool(prerelease),
	}
	if err != nil {
		guiLogs.AppendLog(err.Error())
		release, _, err = client.Repositories.CreateRelease(ctx, githubRepository.Owner, githubRepository.RepoName, createRelease)
		if err != nil {
			guiLogs.AppendLog("创建release失败，err:" + err.Error())
			return
		}
		guiLogs.AppendLog("++++++++++++++++++++创建release成功")
	} else {
		guiLogs.AppendLog("已存在release")
		client.Repositories.EditRelease(ctx, githubRepository.Owner, githubRepository.RepoName, release.GetID(), createRelease)
	}
	guiLogs.AppendLog("开始上传")
	for _, s := range packageFile {
		uploadFile(s, client, ctx, release, *githubRepository)
	}
	guiLogs.AppendLog("上传结束")
}

func releaseGitee(fileTypes []string, commitMessage, releaseVersion, packageDir string) {
	guiLogs.AppendLog("++++++++++++++++++++开始发布到Gitee++++++++++++++++++++")
	tagName, name, body, prerelease, _ := getReleaseMsg(fileTypes, commitMessage, releaseVersion, packageDir)
	createReleaseURL := fmt.Sprintf("https://gitee.com/api/v5/repos/%s/%s/releases?access_token=%s", giteeRepository.Owner, giteeRepository.RepoName, giteeRepository.Token)
	createReleaseResponse, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"tag_name":         tagName,
			"name":             name,
			"body":             body,
			"target_commitish": "master",
			"prerelease":       prerelease,
		}).
		Post(createReleaseURL)

	if err != nil {
		guiLogs.AppendLog("Failed to create release,err:" + err.Error())
		return
	}

	// 检查创建 Release 的响应状态码
	if createReleaseResponse.StatusCode() != 201 {
		guiLogs.AppendLog("Failed to create release. Status code:" + strconv.Itoa(createReleaseResponse.StatusCode()))
		guiLogs.AppendLog("Response body:" + createReleaseResponse.String())
		return
	}
	guiLogs.AppendLog("++++++++++++++++++++创建release成功")
}

func pushAll(gitPlatform []GitPlatform) {
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
		guiLogs.AppendLog("++++++++++++++++++++推送到" + origin + "成功")
	} else {
		guiLogs.AppendLog("推送到" + origin + "失败,err：" + err.Error())
	}
}

func uploadFile(path string, client *github.Client, ctx context.Context, release *github.RepositoryRelease, repository common.GitRepository) {
	// 打开本地文件
	file, err := os.Open(path)
	if err != nil {
		guiLogs.AppendLog("无法打开文件：" + err.Error())
		return
	}
	defer file.Close()
	fileName := filepath.Base(file.Name())
	guiLogs.AppendLog("上传文件：" + fileName)
	// 创建上传选项
	opts := &github.UploadOptions{Name: fileName}

	// 上传文件到 GitHub Release
	asset, _, err := client.Repositories.UploadReleaseAsset(ctx, repository.Owner, repository.RepoName, release.GetID(), opts, file)
	if err != nil {
		guiLogs.AppendLog(fileName + "上传失败,err:" + err.Error())
		return
	}

	guiLogs.AppendLog("++++++++++++++++++++" + fileName + "上传成功, Asset ID:" + strconv.Itoa(int(asset.GetID())))
}

func getJavaMavenPackageFile(fileTypes []string, packageDir string) []string {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		guiLogs.AppendLog("Error getting current directory:" + err.Error())
		return nil
	}

	// 构建 target 目录的完整路径
	targetDir := filepath.Join(currentDir, packageDir)

	// 获取 target 目录下的所有文件
	files, err := getFilesInDirectory(targetDir, fileTypes)
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
