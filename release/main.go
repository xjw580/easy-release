package release

import (
	"bufio"
	"context"
	"easy-release/common"
	"errors"
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
	"syscall"
	"unicode"
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

const (
	GiteePlatform  = "Gitee"
	GithubPlatform = "Github"
)

type GitPlatform string

type ProjectType interface {
	PushPlatform(gitPlatform []GitPlatform) bool
	PackageProject() bool
	ReleasePackage(fileTypes []string, commitMessage, releaseVersion string, gitPlatform []GitPlatform, isPreRelease bool) bool
	DeployPackage() bool
}

type JavaMavenProject struct {
}

func (project JavaMavenProject) PushPlatform(gitPlatform []GitPlatform) bool {
	return pushAll(gitPlatform)
}
func (project JavaMavenProject) PackageProject() bool {
	guiLogs.AppendLog("++++++++++++++++++++开始打包++++++++++++++++++++")
	err := execCMD("mvn", "clean", "package")
	if err == nil {
		guiLogs.AppendLog("==========>打包成功<==========👌👌👌")
		return true
	} else {
		guiLogs.AppendLog("=====>打包失败,err:" + err.Error())
		return false
	}
}
func (project JavaMavenProject) ReleasePackage(fileTypes []string, commitMessage, releaseVersion string, gitPlatform []GitPlatform, isPreRelease bool) bool {
	return releaseAll(fileTypes, commitMessage, releaseVersion, "target", gitPlatform, isPreRelease)
}
func (project JavaMavenProject) DeployPackage() bool {
	guiLogs.AppendLog("++++++++++++++++++++开始部署++++++++++++++++++++")
	err := execCMD("mvn", "deploy")
	if err == nil {
		guiLogs.AppendLog("==========>部署成功<==========👌👌👌")
		return true
	} else {
		guiLogs.AppendLog("=====>部署失败,err:" + err.Error())
		return false
	}
}

type GoProject struct {
}

func (project GoProject) PushPlatform(gitPlatform []GitPlatform) bool {
	return pushAll(gitPlatform)
}
func (project GoProject) PackageProject() bool {
	guiLogs.AppendLog("++++++++++++++++++++开始打包++++++++++++++++++++")
	delExeFile()
	err := execCMD("go", "build")
	if err == nil {
		guiLogs.AppendLog("==========>打包成功<==========👌👌👌")
		return true
	} else {
		guiLogs.AppendLog("=====>打包失败,err:" + err.Error())
		return false
	}
}
func (project GoProject) ReleasePackage(fileTypes []string, commitMessage, releaseVersion string, gitPlatform []GitPlatform, isPreRelease bool) bool {
	return releaseAll(fileTypes, commitMessage, releaseVersion, "", gitPlatform, isPreRelease)
}
func (project GoProject) DeployPackage() bool {
	return true
}

// ParseVersionAndPreRelease @return: 版本号,是否预览版
func ParseVersionAndPreRelease(commitMessage string) (string, bool) {
	lines := strings.Split(commitMessage, "\n")
	if len(lines) > 0 {
		// 解析版本号，假设版本号在提交信息的第一行
		version := strings.TrimSpace(lines[0])

		// 判断是否预发布
		prerelease := !strings.HasSuffix(version, "GA") && !unicode.IsDigit(rune(version[len(version)-1]))

		return version, prerelease
	}

	// 如果没有版本号，默认使用 "v1.0.0" 并标记为非预发布
	return "v1.0.0", false
}

func GetLatestCommitMessage() (string, error) {
	cmd := exec.Command("git", "log", "-1", "--pretty=%B")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func RequireLogs(logs common.Logs) {
	guiLogs = logs
}

func delExeFile() {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	// 遍历当前目录下的所有文件
	err = filepath.Walk(currentDir, func(path string, info os.FileInfo, err error) error {
		// 检查文件是否是exe文件
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".exe") {
			// 删除exe文件
			err := os.Remove(path)
			if err != nil {
				fmt.Println("Error deleting file:", err)
				return err
			}
			fmt.Println("Deleted:", path)
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error walking the path:", err)
		return
	}

	fmt.Println("Deletion complete.")
}

func releaseAll(fileTypes []string, commitMessage, releaseTag, packageDir string, gitPlatform []GitPlatform, isPreRelease bool) bool {
	var wg sync.WaitGroup
	var result = true
	for _, platform := range gitPlatform {
		switch platform {
		case GiteePlatform:
			wg.Add(1)
			go func() {
				defer wg.Done()
				if !releaseGitee(fileTypes, commitMessage, releaseTag, packageDir, isPreRelease) {
					result = false
				}
			}()
		case GithubPlatform:
			wg.Add(1)
			go func() {
				defer wg.Done()
				if !releaseGithub(fileTypes, commitMessage, releaseTag, packageDir, isPreRelease) {
					result = false
				}
			}()
		}
	}
	wg.Wait()
	return result
}

/*
*
return tagName, name, body, prerelease, packageFile
*/
func getReleaseMsg(fileTypes []string, commitMessage, releaseVersion, packageDir string) (string, string, string, bool, []string, error) {
	version, prerelease := ParseVersionAndPreRelease(commitMessage)
	commitMessage = "#### " + commitMessage
	packageFile := getJavaMavenPackageFile(fileTypes, packageDir)
	if packageFile == nil || len(packageFile) == 0 {
		return "", "", "", false, nil, errors.New("packageFile为空")
	}
	fileName := filepath.Base(packageFile[0])
	index := strings.LastIndex(fileName, ".")
	title := fileName[:index]
	if !strings.Contains(title, "v") {
		title = title + "_" + version
	}
	return releaseVersion, title, commitMessage, prerelease, packageFile, nil
}

func releaseGithub(fileTypes []string, commitMessage, releaseVersion, packageDir string, isPreRelease bool) bool {
	guiLogs.AppendLog("++++++++++++++++++++开始发布到Github++++++++++++++++++++")
	tagName, name, body, _, packageFile, err := getReleaseMsg(fileTypes, commitMessage, releaseVersion, packageDir)
	if err != nil {
		return false
	}
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
		TargetCommitish: github.String(getCurrentBranch()), // 或者你的默认分支
		Body:            github.String(body),
		//Draft:           github.Bool(true),
		Prerelease: github.Bool(isPreRelease),
	}
	if err != nil {
		guiLogs.AppendLog(err.Error())
		release, _, err = client.Repositories.CreateRelease(ctx, githubRepository.Owner, githubRepository.RepoName, createRelease)
		if err != nil {
			guiLogs.AppendLog("Github=====>创建release失败，err:" + err.Error())
			return false
		}
		guiLogs.AppendLog("Github==========>创建release成功<==========👌👌👌")
	} else {
		guiLogs.AppendLog("Github=====>已存在release")
		client.Repositories.EditRelease(ctx, githubRepository.Owner, githubRepository.RepoName, release.GetID(), createRelease)
	}
	guiLogs.AppendLog("Github=====>开始上传")
	var result = true
	for _, s := range packageFile {
		if !uploadFileToGithubRelease(s, client, ctx, release, *githubRepository) {
			result = false
		}
	}
	guiLogs.AppendLog("Github=====>上传结束")
	return result
}

func releaseGitee(fileTypes []string, commitMessage, releaseVersion, packageDir string, isPreRelease bool) bool {
	guiLogs.AppendLog("++++++++++++++++++++开始发布到Gitee++++++++++++++++++++")
	tagName, name, body, _, _, err := getReleaseMsg(fileTypes, commitMessage, releaseVersion, packageDir)
	if err != nil {
		return false
	}
	createReleaseURL := fmt.Sprintf("https://gitee.com/api/v5/repos/%s/%s/releases?access_token=%s", giteeRepository.Owner, giteeRepository.RepoName, giteeRepository.Token)
	createReleaseResponse, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"tag_name":         tagName,
			"name":             name,
			"body":             body,
			"target_commitish": getCurrentBranch(),
			"prerelease":       isPreRelease,
		}).
		Post(createReleaseURL)

	if err != nil {
		guiLogs.AppendLog("Gitee=====>Failed to create release,err:" + err.Error())
		return false
	}

	// 检查创建 Release 的响应状态码
	if createReleaseResponse.StatusCode() != 201 {
		guiLogs.AppendLog("Gitee=====>Failed to create release. Status code:" + strconv.Itoa(createReleaseResponse.StatusCode()))
		guiLogs.AppendLog("Gitee=====>Response body:" + createReleaseResponse.String())
		return false
	}
	guiLogs.AppendLog("Gitee==========>创建release成功<==========👌👌👌")
	return true
}

func pushAll(gitPlatform []GitPlatform) bool {
	var wg sync.WaitGroup
	var result = true
	branch := getCurrentBranch()
	for _, platform := range gitPlatform {
		switch platform {
		case GiteePlatform:
			wg.Add(1)
			go func() {
				defer wg.Done()
				guiLogs.AppendLog("++++++++++++++++++++开始Push到Gitee++++++++++++++++++++")
				if !push("Gitee", branch, branch) {
					result = false
				}
			}()
		case GithubPlatform:
			wg.Add(1)
			go func() {
				defer wg.Done()
				guiLogs.AppendLog("++++++++++++++++++++开始Push到Github++++++++++++++++++++")
				if !push("Github", branch, branch) {
					result = false
				}
			}()
		}
	}
	wg.Wait()
	return result
}

func push(origin, localBranch, remoteBranch string) bool {
	err := execCMD("git", "push", origin, localBranch+":"+remoteBranch)
	if err == nil {
		guiLogs.AppendLog("==========>推送到" + origin + "成功<==========👌👌👌")
		return true
	} else {
		guiLogs.AppendLog("=====>推送到" + origin + "失败,err：" + err.Error())
		return false
	}
}

func uploadFileToGithubRelease(path string, client *github.Client, ctx context.Context, release *github.RepositoryRelease, repository common.GitRepository) bool {
	// 打开本地文件
	file, err := os.Open(path)
	if err != nil {
		guiLogs.AppendLog("Github=====>无法打开上传文件：" + err.Error())
		return false
	}
	defer file.Close()
	fileName := filepath.Base(file.Name())
	guiLogs.AppendLog("Github=====>上传文件：" + fileName)
	// 创建上传选项
	opts := &github.UploadOptions{Name: fileName}

	// 上传文件到 GitHub Release
	asset, _, err := client.Repositories.UploadReleaseAsset(ctx, repository.Owner, repository.RepoName, release.GetID(), opts, file)
	if err != nil {
		guiLogs.AppendLog("Github=====>" + fileName + "上传失败,err:" + err.Error())
		return false
	}

	guiLogs.AppendLog("Github==========>" + fileName + "上传成功, Asset ID:" + strconv.Itoa(int(asset.GetID())) + "<==========👌👌👌")
	return true
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
	// 隐藏窗口
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
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

func getFilesInDirectory(dirPath string, extensions []string) ([]string, error) {
	var files []string

	// 遍历目录下的文件
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 检查文件是否是目标扩展名之一
		for _, ext := range extensions {
			lowerName := strings.ToLower(info.Name())
			if strings.HasSuffix(lowerName, ext) && !strings.Contains(lowerName, "javadoc") && !strings.Contains(lowerName, "sources") {
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

// 获取当前git分支
func getCurrentBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")

	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return "master"
	}

	branch := strings.TrimSpace(string(output))
	return branch
}
