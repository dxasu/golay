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
		catLabel := canvas.NewText(string(cat), color.NRGBA{R: 60, G: 80, B: 160, A: 255})
		catLabel.TextSize = 11
		catLabel.TextStyle = fyne.TextStyle{Bold: true}
		items = append(items,
			widget.NewSeparator(),
			container.NewPadded(catLabel),
		)

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
	icon.CornerRadius = 3
	icon.SetMinSize(fyne.NewSize(18, 18))

	nameLabel := canvas.NewText(info.DisplayName, color.NRGBA{R: 30, G: 30, B: 50, A: 255})
	nameLabel.TextSize = 12

	engLabel := canvas.NewText(info.Name, color.NRGBA{R: 130, G: 130, B: 150, A: 200})
	engLabel.TextSize = 9

	labels := container.NewVBox(
		container.NewPadded(nameLabel),
		container.NewPadded(engLabel),
	)
	row := container.NewBorder(nil, nil, container.NewPadded(icon), nil, labels)

	btn := widget.NewButton("", func() {
		if tb.OnAddWidget != nil {
			tb.OnAddWidget(info.Type)
		}
	})
	btn.Importance = widget.LowImportance

	return container.NewStack(btn, container.NewPadded(row))
}
