package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/mico/golay/designer"
)

func main() {
	a := app.NewWithID("com.golay.designer")
	a.Settings().SetTheme(newGolayTheme())

	w := a.NewWindow("golay — Go 可视化 GUI 设计器")
	w.Resize(fyne.NewSize(1200, 760))
	w.SetMaster()

	d := designer.NewApp(w)
	d.Build()

	w.ShowAndRun()
}
