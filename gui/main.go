package gui

import (
	"easy-release/release"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	_ "github.com/lxn/win"
	"log"
	"syscall"
)

type MyWindow struct {
	mainWin                       *walk.MainWindow
	goRadio, javaMavenRadio       *walk.RadioButton
	githubCheckBox, giteeCheckBox *walk.CheckBox
	versionLineEdit, typeLineEdit *walk.LineEdit
	logTextEdit, commentTextEdit  *walk.TextEdit
}

const (
	winWidth  = 800
	winHeight = 1000
)

var mv = new(MyWindow)

func init() {
	err2 := MainWindow{
		AssignTo: &mv.mainWin,
		Bounds: Rectangle{
			X:      int(getDisplayWidth()-winWidth) / 2,
			Y:      int(getDisplayHeight()-winHeight) / 2,
			Width:  winWidth,
			Height: winHeight,
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
						AssignTo: &mv.javaMavenRadio,
						Text:     "Java-Maven",
						Row:      1,
						Column:   1,
					},
					RadioButton{
						AssignTo: &mv.goRadio,
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
						AssignTo: &mv.giteeCheckBox,
						Text:     "Gitee",
						Row:      2,
						Column:   1,
					},
					CheckBox{
						AssignTo: &mv.githubCheckBox,
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
							LineEdit{
								AssignTo:  &mv.typeLineEdit,
								CueBanner: "例：jar&zip&exe",
							},
						},
						Row:        3,
						Column:     0,
						ColumnSpan: 3,
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
								AssignTo: &mv.versionLineEdit,
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
								AssignTo: &mv.commentTextEdit,
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
						AssignTo: &mv.logTextEdit,
						VScroll:  true,
						ReadOnly: true,
					},
				},
			},

			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						MaxSize: Size{Width: 50},
						Text:    "执行",
						OnClicked: func() {
							go func() {
								if mv.javaMavenRadio.Checked() {
									project := new(release.JavaMavenProject)
									project.PackageProject()
								} else if mv.goRadio.Checked() {
									project := new(release.GoProject)
									project.PackageProject()
								}
							}()
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
	release.SetLogs(new(GUILogs))
	mv.javaMavenRadio.SetChecked(true)
	commitMessage, _ := release.GetLatestCommitMessage()
	version, _ := release.ParseVersionAndPreRelease(commitMessage)
	_ = mv.versionLineEdit.SetText(version)
	_ = mv.commentTextEdit.SetText(commitMessage)
	setAlwaysOnTop(mv.mainWin.Handle(), true)
	mv.mainWin.Run()
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
	mv.logTextEdit.AppendText(log + "\r\n")
}

func setAlwaysOnTop(hwnd win.HWND, onTop bool) {
	if onTop {
		win.SetWindowPos(hwnd, win.HWND_TOPMOST, 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE)
	} else {
		win.SetWindowPos(hwnd, win.HWND_NOTOPMOST, 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE)
	}
}
