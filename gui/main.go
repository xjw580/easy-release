package gui

import (
	"easy-release/common"
	"easy-release/release"
	_ "encoding/json"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	_ "github.com/lxn/win"
	"log"
	"syscall"
)

type MyWindow struct {
	mainWin                                                                                                                                              *walk.MainWindow
	goRadio, javaMavenRadio                                                                                                                              *walk.RadioButton
	githubCheckBox, giteeCheckBox, pushCheckBox, packageCheckBox, releaseCheckBox, allProgramCheckBox, zipFileCheckBox, jarFileCheckBox, exeFileCheckBox *walk.CheckBox
	versionLineEdit                                                                                                                                      *walk.LineEdit
	logTextEdit, commitMsgTextEdit                                                                                                                       *walk.TextEdit
}

const (
	winWidth  = 800
	winHeight = 1000
)

var mw = new(MyWindow)

func init() {
	image, _ := walk.Resources.Image("embedded_favicon.ico")
	err2 := MainWindow{
		Title:    common.ProgramName,
		Icon:     image,
		AssignTo: &mw.mainWin,
		Bounds: Rectangle{
			X:      int(getDisplayWidth()-winWidth) / 2,
			Y:      int(getDisplayHeight()-winHeight-40) / 2,
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
						OnClicked: func() {
							if mw.javaMavenRadio.Checked() {
								mw.zipFileCheckBox.SetChecked(true)
								mw.jarFileCheckBox.SetChecked(true)
								mw.exeFileCheckBox.SetChecked(false)
							}
						},
					},
					RadioButton{
						AssignTo: &mw.goRadio,
						Text:     "Go",
						Row:      1,
						Column:   2,
						OnClicked: func() {
							if mw.goRadio.Checked() {
								mw.zipFileCheckBox.SetChecked(false)
								mw.jarFileCheckBox.SetChecked(false)
								mw.exeFileCheckBox.SetChecked(true)
							}
						},
					},

					Label{
						Text:   "发布：",
						Row:    2,
						Column: 0,
					},
					CheckBox{
						AssignTo: &mw.giteeCheckBox,
						Text:     "Gitee",
						Checked:  true,
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
						AssignTo: &mw.zipFileCheckBox,
						Text:     ".zip",
						Checked:  true,
						Row:      3,
						Column:   1,
					},
					CheckBox{
						AssignTo: &mw.jarFileCheckBox,
						Text:     ".jar",
						Checked:  true,
						Row:      3,
						Column:   2,
					},
					CheckBox{
						AssignTo: &mw.exeFileCheckBox,
						Text:     ".exe",
						Row:      3,
						Column:   3,
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
								OnKeyPress: func(key walk.Key) {
									if key.String() == "Return" {
										_ = mw.commitMsgTextEdit.SetFocus()
									}
								},
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
								AssignTo: &mw.commitMsgTextEdit,
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
					Composite{
						Layout: HBox{},
						Children: []Widget{
							Label{Text: "输出日志："},
							PushButton{
								MaxSize: Size{
									Width: 110,
								},
								Text: "刷新版本信息",
								OnClicked: func() {
									go setVersionMsg()
								},
							},
							PushButton{
								MaxSize: Size{
									Width: 80,
								},
								Text: "清空日志",
								OnClicked: func() {
									_ = mw.logTextEdit.SetText("")
								},
							},
						},
					},
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
						OnClicked: func() {
							progressCheck()
						},
					},
					CheckBox{
						AssignTo: &mw.packageCheckBox,
						Text:     "打包",
						OnClicked: func() {
							progressCheck()
						},
					},
					CheckBox{
						AssignTo: &mw.releaseCheckBox,
						Text:     "发布",
						OnClicked: func() {
							progressCheck()
						},
					},
					PushButton{
						MaxSize: Size{Width: 50},
						Text:    "执行",
						OnClicked: func() {
							go func() {
								setAlwaysOnTop(mw.mainWin.Handle(), true)
								mw.logTextEdit.AppendText("++++++++++++++++++++开始执行++++++++++++++++++++\r\n")
								/*项目类型*/
								var project release.ProjectType
								if mw.javaMavenRadio.Checked() {
									project = new(release.JavaMavenProject)
								} else if mw.goRadio.Checked() {
									project = new(release.GoProject)
								}
								/*package*/
								if mw.packageCheckBox.Checked() {
									project.PackageProject()
								}
								/*git平台*/
								var platforms []release.GitPlatform
								if mw.githubCheckBox.Checked() {
									platforms = append(platforms, release.GithubPlatform)
								}
								if mw.giteeCheckBox.Checked() {
									platforms = append(platforms, release.GiteePlatform)
								}
								/*push*/
								if mw.pushCheckBox.Checked() {
									project.PushPlatform(platforms)
								}
								/*release*/
								if mw.releaseCheckBox.Checked() {
									var fileTypes []string
									if mw.zipFileCheckBox.Checked() {
										fileTypes = append(fileTypes, mw.zipFileCheckBox.Text())
									}
									if mw.jarFileCheckBox.Checked() {
										fileTypes = append(fileTypes, mw.jarFileCheckBox.Text())
									}
									if mw.exeFileCheckBox.Checked() {
										fileTypes = append(fileTypes, mw.exeFileCheckBox.Text())
									}
									project.ReleasePackage(fileTypes, mw.commitMsgTextEdit.Text(), mw.versionLineEdit.Text(), platforms)
								}
								mw.logTextEdit.AppendText("++++++++++++++++++++执行完毕++++++++++++++++++++\r\n")
								mw.logTextEdit.AppendText("\r\n                            .__          __             .___\r\n  ____  ____   _____ ______ |  |   _____/  |_  ____   __| _/\r\n_/ ___\\/  _ \\ /     \\\\____ \\|  | _/ __ \\   __\\/ __ \\ / __ | \r\n\\  \\__(  <_> )  Y Y  \\  |_> >  |_\\  ___/|  | \\  ___// /_/ | \r\n \\___  >____/|__|_|  /   __/|____/\\___  >__|  \\___  >____ | \r\n     \\/            \\/|__|             \\/          \\/     \\/ \r\n")
								setAlwaysOnTop(mw.mainWin.Handle(), false)
							}()
						},
					},
					CheckBox{
						AssignTo: &mw.allProgramCheckBox,
						Text:     "全部",
						OnClicked: func() {
							if mw.allProgramCheckBox.Checked() {
								mw.pushCheckBox.SetChecked(true)
								mw.packageCheckBox.SetChecked(true)
								mw.releaseCheckBox.SetChecked(true)
							} else {
								mw.pushCheckBox.SetChecked(false)
								mw.packageCheckBox.SetChecked(false)
								mw.releaseCheckBox.SetChecked(false)
							}
						},
					},
				},
			},
		},
	}.Create()
	if err2 != nil {
		log.Println(err2)
		return
	}
	go initMsg()
	mw.mainWin.Run()
}

func progressCheck() {
	if mw.pushCheckBox.Checked() && mw.packageCheckBox.Checked() && mw.releaseCheckBox.Checked() {
		mw.allProgramCheckBox.SetChecked(true)
	} else {
		mw.allProgramCheckBox.SetChecked(false)
	}
}

func initMsg() {
	release.RequireLogs(new(GUILogs))
	mw.javaMavenRadio.SetChecked(true)
	setVersionMsg()
}
func setVersionMsg() {
	commitMessage, _ := release.GetLatestCommitMessage()
	version, _ := release.ParseVersionAndPreRelease(commitMessage)
	_ = mw.versionLineEdit.SetText(version)
	_ = mw.commitMsgTextEdit.SetText(commitMessage)
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

func giteeSettings() {
	config, _ := common.ReadConfigFromFile()
	showGitSettingsWindow("Gitee设置", &config.GiteeRepository, &config)
}
func githubSettings() {
	config, _ := common.ReadConfigFromFile()
	showGitSettingsWindow("Github设置", &config.GithubRepository, &config)
}

func showGitSettingsWindow(title string, repository *common.GitRepository, config *common.Config) {
	var gitSettingsMW *walk.MainWindow
	var ownerLineEdit, repoNameLineEdit, tokenLineEdit *walk.LineEdit
	_ = MainWindow{
		AssignTo: &gitSettingsMW,
		Title:    common.ProgramName + "-" + title,
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
				OnKeyPress: func(key walk.Key) {
					if key.String() == "Return" {
						_ = repoNameLineEdit.SetFocus()
					}
				},
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
				OnKeyPress: func(key walk.Key) {
					if key.String() == "Return" {
						_ = tokenLineEdit.SetFocus()
					}
				},
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
				OnKeyPress: func(key walk.Key) {
					if key.String() == "Return" {
						_ = ownerLineEdit.SetFocus()
					}
				},
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
							repository.Owner = ownerLineEdit.Text()
							repository.RepoName = repoNameLineEdit.Text()
							repository.Token = tokenLineEdit.Text()
							_ = common.WriteConfigToFile(*config)
							log.Println("保存成功")
							_ = gitSettingsMW.Close()
						},
					},
				},
				Row:        3,
				Column:     0,
				ColumnSpan: 2,
			},
		},
	}.Create()

	ownerLineEdit.SetText(repository.Owner)
	repoNameLineEdit.SetText(repository.RepoName)
	tokenLineEdit.SetText(repository.Token)
	gitSettingsMW.Run()
}
