package gui

import (
	"easy-release/common"
	"easy-release/release"
	"encoding/json"
	_ "encoding/json"
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	_ "github.com/lxn/win"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

type MyWindow struct {
	mainWin                                                                                           *walk.MainWindow
	goRadio, javaMavenRadio                                                                           *walk.RadioButton
	githubCheckBox, giteeCheckBox, pushCheckBox, packageCheckBox, releaseCheckBox, allProgramCheckBox *walk.CheckBox
	versionLineEdit                                                                                   *walk.LineEdit
	logTextEdit, commentTextEdit                                                                      *walk.TextEdit
}

const (
	winWidth       = 800
	winHeight      = 1000
	configFilePath = "C:\\ProgramData\\" + common.ProgramName + "\\config.json"
)

var mw = new(MyWindow)

func init() {
	err2 := MainWindow{
		Title:    common.ProgramName,
		AssignTo: &mw.mainWin,
		Bounds: Rectangle{
			X:      int(getDisplayWidth()-winWidth) / 2,
			Y:      int(getDisplayHeight()-winHeight) / 2,
			Width:  winWidth,
			Height: winHeight,
		},

		MenuItems: []MenuItem{
			Menu{
				Text: "仓库设置",
				Items: []MenuItem{
					Action{
						Text: "Github",
						OnTriggered: func() {
							githubSettings()
						},
					},
					Action{
						Text: "Gitee",
						OnTriggered: func() {
							giteeSettings()
						},
					},
				},
			},
		},

		Font: Font{
			PointSize: 11,
		},
		Background: SolidColorBrush{
			Color: walk.RGB(224, 240, 253),
		},
		Layout: VBox{},
		Children: []Widget{
			Composite{
				Layout: Grid{Spacing: 10},
				Children: []Widget{
					Label{
						Text:   "构建：",
						Row:    1,
						Column: 0,
					},
					RadioButton{
						AssignTo: &mw.javaMavenRadio,
						Text:     "Java-Maven",
						Row:      1,
						Column:   1,
					},
					RadioButton{
						AssignTo: &mw.goRadio,
						Text:     "Go",
						Row:      1,
						Column:   2,
					},

					Label{
						Text:   "发布：",
						Row:    2,
						Column: 0,
					},
					CheckBox{
						AssignTo: &mw.giteeCheckBox,
						Text:     "Gitee",
						Row:      2,
						Column:   1,
					},
					CheckBox{
						AssignTo: &mw.githubCheckBox,
						Text:     "Github",
						Checked:  true,
						Row:      2,
						Column:   2,
					},

					Composite{
						Layout: HBox{Margins: Margins{Left: 1}},
						MaxSize: Size{
							Width: 125,
						},
						Children: []Widget{
							Label{
								Text: "发布资源文件类型：",
							},
						},
						Row:        3,
						Column:     0,
						ColumnSpan: 1,
					},
					CheckBox{
						Text:    ".zip",
						Checked: true,
						Row:     3,
						Column:  1,
					},
					CheckBox{
						Text:    ".jar",
						Checked: true,
						Row:     3,
						Column:  2,
					},
					CheckBox{
						Text:   ".exe",
						Row:    3,
						Column: 3,
					},

					Composite{
						Layout: VBox{Margins: Margins{Left: 1}},
						MaxSize: Size{
							Width: 125,
						},
						Children: []Widget{
							Label{
								Text: "版本号：",
							},
							LineEdit{
								AssignTo: &mw.versionLineEdit,
							},
						},
						Row:    4,
						Column: 0,
					},
					Composite{
						Layout: VBox{Margins: Margins{Left: 1}},
						MaxSize: Size{
							Height: 160,
						},
						MinSize: Size{
							Height: 160,
						},
						Children: []Widget{
							Label{
								Text: "版本说明：",
							},
							TextEdit{
								AssignTo: &mw.commentTextEdit,
								VScroll:  true,
							},
						},
						Row:        4,
						Column:     1,
						ColumnSpan: 5,
					},
				},
			},

			Composite{
				Layout: VBox{},
				Children: []Widget{
					Label{Text: "输出日志："},
					TextEdit{
						AssignTo: &mw.logTextEdit,
						VScroll:  true,
						ReadOnly: true,
					},
				},
			},

			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						Text: "过程：",
					},
					CheckBox{
						AssignTo: &mw.pushCheckBox,
						Text:     "推送",
						Checked:  true,
					},
					CheckBox{
						AssignTo: &mw.packageCheckBox,
						Text:     "打包",
						Checked:  true,
					},
					CheckBox{
						AssignTo: &mw.releaseCheckBox,
						Text:     "发布",
						Checked:  true,
					},
					PushButton{
						MaxSize: Size{Width: 50},
						Text:    "执行",
						OnClicked: func() {
							go func() {
								setAlwaysOnTop(mw.mainWin.Handle(), true)
								if mw.javaMavenRadio.Checked() {
									project := new(release.JavaMavenProject)
									project.PackageProject()
								} else if mw.goRadio.Checked() {
									project := new(release.GoProject)
									project.PackageProject()
								}
								setAlwaysOnTop(mw.mainWin.Handle(), false)
							}()
						},
					},
					CheckBox{
						AssignTo: &mw.allProgramCheckBox,
						Text:     "全部",
						Checked:  true,
					},
				},
			},
		},
	}.Create()
	if err2 != nil {
		log.Println(err2)
		return
	}
	initMsg()
	mw.mainWin.Run()
}

func giteeSettings() {

}

func initMsg() {
	release.RequireLogs(new(GUILogs))
	mw.javaMavenRadio.SetChecked(true)
	commitMessage, _ := release.GetLatestCommitMessage()
	version, _ := release.ParseVersionAndPreRelease(commitMessage)
	_ = mw.versionLineEdit.SetText(version)
	_ = mw.commentTextEdit.SetText(commitMessage)
}

/*
*
获取显示器宽度
*/
func getDisplayWidth() uintptr {
	w, _, _ := syscall.NewLazyDLL(`User32.dll`).NewProc(`GetSystemMetrics`).Call(uintptr(0))
	return w
}

/*
*
获取显示器高度
*/
func getDisplayHeight() uintptr {
	h, _, _ := syscall.NewLazyDLL(`User32.dll`).NewProc(`GetSystemMetrics`).Call(uintptr(1))
	return h
}

type GUILogs struct {
}

func (logs GUILogs) AppendLog(log string) {
	mw.logTextEdit.AppendText(log + "\r\n")
}

func setAlwaysOnTop(hwnd win.HWND, onTop bool) {
	if onTop {
		win.SetWindowPos(hwnd, win.HWND_TOPMOST, 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE)
	} else {
		win.SetWindowPos(hwnd, win.HWND_NOTOPMOST, 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE)
	}
}
func githubSettings() {
	var githubSettingsMW *walk.MainWindow
	var ownerLineEdit, repoNameLineEdit, tokenLineEdit *walk.LineEdit
	loadedConfig, _ := readConfigFromFile(configFilePath)
	_ = MainWindow{
		AssignTo: &githubSettingsMW,
		Title:    common.ProgramName + "-Github设置",
		Font: Font{
			PointSize: 11,
		},
		Background: SolidColorBrush{
			Color: walk.RGB(224, 240, 253),
		},
		Bounds: Rectangle{
			Width: 550,
			X:     500,
			Y:     500,
		},
		Layout: Grid{Margins: Margins{Top: 10, Left: 10, Right: 10, Bottom: 0}},
		Children: []Widget{
			Label{
				Text:   "owner:",
				Row:    0,
				Column: 0,
			},
			LineEdit{
				AssignTo: &ownerLineEdit,
				Row:      0,
				Column:   1,
			},

			Label{
				Text:   "repoName:",
				Row:    1,
				Column: 0,
			},
			LineEdit{
				AssignTo: &repoNameLineEdit,
				Row:      1,
				Column:   1,
			},

			Label{
				Text:   "token:",
				Row:    2,
				Column: 0,
			},
			LineEdit{
				AssignTo: &tokenLineEdit,
				Row:      2,
				Column:   1,
			},

			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text: "保存",
						MaxSize: Size{
							Width: 50,
						},
						OnClicked: func() {
							loadedConfig.GithubRepository.Owner = ownerLineEdit.Text()
							loadedConfig.GithubRepository.RepoName = repoNameLineEdit.Text()
							loadedConfig.GithubRepository.Token = tokenLineEdit.Text()
							writeConfigToFile(loadedConfig, configFilePath)
							log.Println("保存成功")
						},
					},
				},
				Row:        3,
				Column:     0,
				ColumnSpan: 2,
			},
		},
	}.Create()

	ownerLineEdit.SetText(loadedConfig.GithubRepository.Owner)
	repoNameLineEdit.SetText(loadedConfig.GithubRepository.RepoName)
	tokenLineEdit.SetText(loadedConfig.GithubRepository.Token)
	fmt.Println(loadedConfig)
	githubSettingsMW.Run()
}

func readConfigFromFile(filename string) (Config, error) {
	var config Config
	open, err2 := os.Open(configFilePath)
	defer open.Close()
	if err2 != nil {
		// 创建文件所在目录，如果目录不存在的话
		err := os.MkdirAll(filepath.Dir(configFilePath), os.ModePerm)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			return config, err
		}

		// 创建文件
		file, err := os.Create(configFilePath)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return config, err
		}
		defer file.Close()
		config = Config{
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
		writeConfigToFile(config, configFilePath)

		return config, nil
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(data, &config)
	return config, err
}
func writeConfigToFile(config Config, filename string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
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
