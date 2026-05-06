package designer

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/mico/golay/model"
)

// Toolbox 控件工具箱面板（按分类分组）
type Toolbox struct {
	widget.BaseWidget
	OnAddWidget func(t model.WidgetType)
}

// NewToolbox 创建工具箱
func NewToolbox(onAdd func(t model.WidgetType)) *Toolbox {
	tb := &Toolbox{OnAddWidget: onAdd}
	tb.ExtendBaseWidget(tb)
	return tb
}

func (tb *Toolbox) CreateRenderer() fyne.WidgetRenderer {
	groups := model.WidgetsByCategory()
	items := []fyne.CanvasObject{}

	for _, cat := range model.CategoryOrder {
		ws, ok := groups[cat]
		if !ok || len(ws) == 0 {
			continue
		}
		// 分类标题
		catBg := canvas.NewRectangle(color.NRGBA{R: 210, G: 218, B: 245, A: 255})
		catBg.SetMinSize(fyne.NewSize(0, 18))
		catLabel := canvas.NewText(string(cat), color.NRGBA{R: 40, G: 60, B: 150, A: 255})
		catLabel.TextSize = 9
		catLabel.TextStyle = fyne.TextStyle{Bold: true}
		items = append(items, container.NewStack(catBg, container.NewCenter(catLabel)))

		for _, info := range ws {
			info := info
			btn := tb.makeItem(info)
			items = append(items, btn)
		}
	}

	scroll := container.NewVScroll(container.NewVBox(items...))
	bg := canvas.NewRectangle(color.NRGBA{R: 232, G: 236, B: 250, A: 255})
	return widget.NewSimpleRenderer(container.NewMax(bg, scroll))
}

func (tb *Toolbox) makeItem(info model.WidgetTypeInfo) fyne.CanvasObject {
	iconC := color.NRGBA{R: info.Color[0], G: info.Color[1], B: info.Color[2], A: 255}

	icon := canvas.NewRectangle(iconC)
	icon.CornerRadius = 2
	icon.SetMinSize(fyne.NewSize(10, 10))

	nameLabel := canvas.NewText(info.DisplayName, color.NRGBA{R: 30, G: 30, B: 50, A: 255})
	nameLabel.TextSize = 11

	engLabel := canvas.NewText("  "+info.Name, color.NRGBA{R: 140, G: 140, B: 158, A: 200})
	engLabel.TextSize = 9

	// 单行：[■] 显示名  英文名
	iconBox := container.NewGridWrap(fyne.NewSize(20, 20), container.NewCenter(icon))
	textRow := container.NewHBox(nameLabel, engLabel)
	row := container.NewBorder(nil, nil, iconBox, nil, textRow)

	// 固定行高 26px
	minH := canvas.NewRectangle(color.Transparent)
	minH.SetMinSize(fyne.NewSize(0, 26))

	btn := widget.NewButton("", func() {
		if tb.OnAddWidget != nil {
			tb.OnAddWidget(info.Type)
		}
	})
	btn.Importance = widget.LowImportance

	return container.NewStack(minH, btn, row)
}
