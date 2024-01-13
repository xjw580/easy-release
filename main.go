package main

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	_ "github.com/lxn/win"
	"log"
	"syscall"
)

type MyWindow struct {
	javaMavenRadio  *walk.RadioButton
	goRadio         *walk.RadioButton
	mainWin         *walk.MainWindow
	giteeCheckBox   *walk.CheckBox
	githubCheckBox  *walk.CheckBox
	typeLineEdit    *walk.LineEdit
	versionLineEdit *walk.LineEdit
	commentTextEdit *walk.TextEdit
}

const (
	winWidth  = 800
	winHeight = 1000
)

func main() {
	mainWin := new(MyWindow)
	err2 := MainWindow{
		AssignTo: &mainWin.mainWin,
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
				Layout: Grid{Spacing: 15},
				Children: []Widget{
					Label{
						Text:   "构建：",
						Row:    1,
						Column: 0,
					},
					RadioButton{
						AssignTo: &mainWin.javaMavenRadio,
						Text:     "Java-Maven",
						Row:      1,
						Column:   1,
					},
					RadioButton{
						AssignTo: &mainWin.goRadio,
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
						AssignTo: &mainWin.giteeCheckBox,
						Text:     "Gitee",
						Row:      2,
						Column:   1,
					},
					CheckBox{
						AssignTo: &mainWin.githubCheckBox,
						Text:     "Github",
						Checked:  true,
						Row:      2,
						Column:   2,
					},

					Composite{
						Layout: HBox{},
						MaxSize: Size{
							Width: 125,
						},
						Children: []Widget{
							Label{
								Text: "发布资源文件类型：",
							},
							LineEdit{
								AssignTo:  &mainWin.typeLineEdit,
								CueBanner: "例：jar&zip&exe",
							},
						},
						Row:        3,
						Column:     0,
						ColumnSpan: 2,
					},

					Composite{
						Layout: VBox{},
						MaxSize: Size{
							Width: 125,
						},
						Children: []Widget{
							Label{
								Text: "版本号：",
							},
							LineEdit{
								AssignTo: &mainWin.versionLineEdit,
							},
						},
						Row:    4,
						Column: 0,
					},
					Composite{
						Layout: VBox{},
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
								AssignTo: &mainWin.commentTextEdit,
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

	mainWin.javaMavenRadio.SetChecked(true)
	mainWin.mainWin.Run()
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
