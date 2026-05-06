package designer

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/mico/golay/model"
)

// PropertiesPanel 属性与事件面板（右侧）
type PropertiesPanel struct {
	widget.BaseWidget
	current        *model.DesignWidget
	currentForm    *model.FormDef
	OnModified     func(dw *model.DesignWidget)
	OnFormModified func(form *model.FormDef)
	OnOpenInEditor func(dw *model.DesignWidget, eventName string)
	content        *fyne.Container
	dc             *DesignCanvas
}

func NewPropertiesPanel(dc *DesignCanvas) *PropertiesPanel {
	pp := &PropertiesPanel{dc: dc}
	pp.ExtendBaseWidget(pp)
	return pp
}

func (pp *PropertiesPanel) CreateRenderer() fyne.WidgetRenderer {
	// 根据预设状态决定初始内容
	var initial fyne.CanvasObject
	switch {
	case pp.current != nil:
		initial = pp.buildWidgetPanel(pp.current)
	case pp.currentForm != nil:
		initial = pp.buildFormPanel(pp.currentForm)
	default:
		initial = pp.buildFormEmpty()
	}
	pp.content = container.NewMax(initial)
	bg := canvas.NewRectangle(color.NRGBA{R: 244, G: 245, B: 252, A: 255})
	return widget.NewSimpleRenderer(container.NewMax(bg, pp.content))
}

// setContent 安全地设置面板内容（content 可能尚未初始化）
func (pp *PropertiesPanel) setContent(obj fyne.CanvasObject) {
	if pp.content == nil {
		// CreateRenderer 尚未被调用，预设状态即可，渲染时会用到
		return
	}
	pp.content.Objects = []fyne.CanvasObject{obj}
	pp.content.Refresh()
}

// ShowWidget 展示控件属性
func (pp *PropertiesPanel) ShowWidget(dw *model.DesignWidget) {
	pp.current = dw
	pp.currentForm = nil
	pp.setContent(pp.buildWidgetPanel(dw))
}

// ShowFormProps 展示 Form 属性（未选中控件时）
func (pp *PropertiesPanel) ShowFormProps(form *model.FormDef) {
	pp.currentForm = form
	pp.current = nil
	pp.setContent(pp.buildFormPanel(form))
}

// ClearWidget 回到 Form 属性（若有当前 Form）
func (pp *PropertiesPanel) ClearWidget() {
	pp.current = nil
	if pp.dc != nil && pp.dc.GetCurrentForm() != nil {
		pp.ShowFormProps(pp.dc.GetCurrentForm())
	} else {
		pp.currentForm = nil
		pp.setContent(pp.buildFormEmpty())
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Form 属性面板
// ─────────────────────────────────────────────────────────────────────────────

func (pp *PropertiesPanel) buildFormEmpty() fyne.CanvasObject {
	t := canvas.NewText("暂无设计内容", color.NRGBA{R: 140, G: 145, B: 170, A: 255})
	t.TextSize = 12
	t.Alignment = fyne.TextAlignCenter
	return container.NewCenter(t)
}

func (pp *PropertiesPanel) buildFormPanel(form *model.FormDef) fyne.CanvasObject {
	title := pp.sectionHeader("🪟  " + form.Name + "  [窗体]")
	tabs := container.NewAppTabs(
		container.NewTabItem("属性", pp.buildFormPropsTab(form)),
		container.NewTabItem("事件", pp.buildEventsTab(form.Events, func(ev model.EventBinding) {
			if pp.OnOpenInEditor != nil {
				pp.OnOpenInEditor(nil, ev.EventName)
			}
		}, func(i int, v bool) {
			form.Events[i].Enabled = v
			if pp.OnFormModified != nil {
				pp.OnFormModified(form)
			}
		}, func(i int, v string) {
			form.Events[i].HandlerFn = v
		})),
	)
	return container.NewBorder(title, nil, nil, nil, tabs)
}

func (pp *PropertiesPanel) buildFormPropsTab(form *model.FormDef) fyne.CanvasObject {
	props := form.GetFormProperties()
	form2 := widget.NewForm()

	for _, p := range props {
		p := p
		var item *widget.FormItem
		switch p.Type {
		case "bool":
			chk := widget.NewCheck("", func(v bool) {
				form.SetFormProperty(p.Name, boolVal(v))
				if pp.OnFormModified != nil {
					pp.OnFormModified(form)
				}
			})
			chk.Checked = p.Value == "true"
			item = widget.NewFormItem(p.Name, chk)
		case "options":
			sel := widget.NewSelect(p.Options, func(v string) {
				form.SetFormProperty(p.Name, v)
				if pp.OnFormModified != nil {
					pp.OnFormModified(form)
				}
			})
			sel.Selected = p.Value
			item = widget.NewFormItem(p.Name, sel)
		default:
			entry := widget.NewEntry()
			entry.SetText(p.Value)
			entry.OnChanged = func(v string) {
				form.SetFormProperty(p.Name, v)
				if pp.OnFormModified != nil {
					pp.OnFormModified(form)
				}
			}
			item = widget.NewFormItem(p.Name, entry)
		}
		item.HintText = p.Description
		form2.AppendItem(item)
	}
	return container.NewVScroll(container.NewPadded(form2))
}

// ─────────────────────────────────────────────────────────────────────────────
// 控件属性面板
// ─────────────────────────────────────────────────────────────────────────────

func (pp *PropertiesPanel) buildWidgetPanel(dw *model.DesignWidget) fyne.CanvasObject {
	info := model.GetWidgetTypeInfo(dw.Type)
	title := pp.sectionHeader(fmt.Sprintf("📦  %s  [%s]", dw.Name, info.DisplayName))

	tabs := container.NewAppTabs(
		container.NewTabItem("属性", pp.buildPropsTab(dw)),
		container.NewTabItem("事件", pp.buildEventsTab(dw.Events,
			func(ev model.EventBinding) {
				if pp.OnOpenInEditor != nil {
					pp.OnOpenInEditor(dw, ev.EventName)
				}
			},
			func(i int, v bool) {
				dw.Events[i].Enabled = v
				pp.notifyModified(dw)
			},
			func(i int, v string) {
				dw.Events[i].HandlerFn = v
			},
		)),
		container.NewTabItem("布局", pp.buildLayoutTab(dw)),
	)
	return container.NewBorder(title, nil, nil, nil, tabs)
}

func (pp *PropertiesPanel) buildPropsTab(dw *model.DesignWidget) fyne.CanvasObject {
	form := widget.NewForm()
	for i := range dw.Properties {
		idx := i
		p := &dw.Properties[idx]
		var item *widget.FormItem
		switch p.Type {
		case "bool":
			chk := widget.NewCheck("", func(v bool) {
				dw.Properties[idx].Value = boolVal(v)
				if p.Name == "Name" {
					dw.Name = dw.Properties[idx].Value
				}
				pp.notifyModified(dw)
			})
			chk.Checked = p.Value == "true"
			item = widget.NewFormItem(p.Name, chk)
		case "options":
			sel := widget.NewSelect(p.Options, func(v string) {
				dw.Properties[idx].Value = v
				pp.notifyModified(dw)
			})
			sel.Selected = p.Value
			item = widget.NewFormItem(p.Name, sel)
		default:
			entry := widget.NewEntry()
			entry.SetText(p.Value)
			entry.OnChanged = func(v string) {
				dw.Properties[idx].Value = v
				if p.Name == "Name" {
					dw.Name = v
				}
				pp.notifyModified(dw)
			}
			item = widget.NewFormItem(p.Name, entry)
		}
		item.HintText = p.Description
		form.AppendItem(item)
	}
	return container.NewVScroll(container.NewPadded(form))
}

func (pp *PropertiesPanel) buildLayoutTab(dw *model.DesignWidget) fyne.CanvasObject {
	makeEntry := func(label string, getV func() float32, setV func(float32)) *widget.FormItem {
		e := widget.NewEntry()
		e.SetText(fmt.Sprintf("%.0f", getV()))
		e.OnChanged = func(v string) {
			if f, err := strconv.ParseFloat(strings.TrimSpace(v), 32); err == nil {
				setV(float32(f))
				pp.notifyModified(dw)
				if pp.dc != nil {
					pp.dc.Refresh()
				}
			}
		}
		return widget.NewFormItem(label, e)
	}
	form := widget.NewForm(
		makeEntry("X", func() float32 { return dw.X }, func(v float32) { dw.X = v }),
		makeEntry("Y", func() float32 { return dw.Y }, func(v float32) { dw.Y = v }),
		makeEntry("宽度", func() float32 { return dw.W }, func(v float32) { dw.W = v }),
		makeEntry("高度", func() float32 { return dw.H }, func(v float32) { dw.H = v }),
	)
	delBtn := widget.NewButtonWithIcon("删除此控件", theme.DeleteIcon(), func() {
		if pp.dc != nil {
			pp.dc.DeleteSelected()
		}
	})
	delBtn.Importance = widget.DangerImportance
	return container.NewVBox(container.NewPadded(form), widget.NewSeparator(), container.NewPadded(delBtn))
}

// ─────────────────────────────────────────────────────────────────────────────
// 事件列表（通用，可用于 DesignWidget 和 FormDef）
// ─────────────────────────────────────────────────────────────────────────────

// buildEventsTab 构建紧凑的事件表格
// onOpen: 双击事件名回调
// onToggle: 启用/禁用回调（i=索引, v=bool）
// onRename: 修改函数名回调
func (pp *PropertiesPanel) buildEventsTab(
	events []model.EventBinding,
	onOpen func(ev model.EventBinding),
	onToggle func(i int, v bool),
	onRename func(i int, v string),
) fyne.CanvasObject {
	if len(events) == 0 {
		t := canvas.NewText("该控件无可绑定事件", color.NRGBA{R: 140, G: 145, B: 170, A: 200})
		t.TextSize = 12
		return container.NewCenter(t)
	}

	// 表头
	header := pp.eventTableHeader()
	rows := []fyne.CanvasObject{header}

	var specIdx, univIdx []int
	for i, ev := range events {
		if isUniversalEvent(ev.EventName) {
			univIdx = append(univIdx, i)
		} else {
			specIdx = append(specIdx, i)
		}
	}

	addSection := func(title string, indices []int) {
		if len(indices) == 0 {
			return
		}
		secBg := canvas.NewRectangle(color.NRGBA{R: 220, G: 225, B: 245, A: 255})
		secBg.SetMinSize(fyne.NewSize(0, 16))
		secTxt := canvas.NewText(title, color.NRGBA{R: 50, G: 60, B: 140, A: 255})
		secTxt.TextSize = 9
		secTxt.TextStyle = fyne.TextStyle{Bold: true}
		rows = append(rows, container.NewStack(secBg, container.NewCenter(secTxt)))

		for _, idx := range indices {
			rows = append(rows, pp.eventRow(idx, events[idx], onOpen, onToggle, onRename))
		}
	}

	addSection("专属事件", specIdx)
	addSection("通用事件", univIdx)

	hint := canvas.NewText("双击事件名 → 在编辑器中跳转到对应函数", color.NRGBA{R: 130, G: 135, B: 165, A: 180})
	hint.TextSize = 9
	rows = append(rows, container.NewPadded(hint))

	return container.NewVScroll(container.NewVBox(rows...))
}

// eventTableHeader 事件表格表头
func (pp *PropertiesPanel) eventTableHeader() fyne.CanvasObject {
	const hdrH = float32(20)
	hBg := canvas.NewRectangle(color.NRGBA{R: 60, G: 70, B: 150, A: 255})
	hBg.SetMinSize(fyne.NewSize(0, hdrH))

	mkHdr := func(t string, w float32) fyne.CanvasObject {
		txt := canvas.NewText(t, color.NRGBA{R: 230, G: 235, B: 255, A: 255})
		txt.TextSize = 9
		txt.TextStyle = fyne.TextStyle{Bold: true}
		if w > 0 {
			return container.NewGridWrap(fyne.NewSize(w, hdrH), txt)
		}
		return txt
	}
	row := container.NewBorder(nil, nil,
		mkHdr("启用", 34),
		mkHdr("函数名", 120),
		mkHdr("事件名 / 签名", 0),
	)
	return container.NewStack(hBg, container.NewPadded(row))
}

// eventRow 单个事件行（紧凑，固定行高 26px）
func (pp *PropertiesPanel) eventRow(
	idx int,
	ev model.EventBinding,
	onOpen func(ev model.EventBinding),
	onToggle func(i int, v bool),
	onRename func(i int, v string),
) fyne.CanvasObject {
	const rowH = float32(26)

	// 行背景
	rowBg := canvas.NewRectangle(color.NRGBA{R: 248, G: 249, B: 255, A: 255})
	if ev.Enabled {
		rowBg.FillColor = color.NRGBA{R: 232, G: 240, B: 255, A: 255}
	}
	rowBg.SetMinSize(fyne.NewSize(0, rowH))

	// 左色条（启用时蓝色）
	sideBar := canvas.NewRectangle(color.Transparent)
	sideBar.SetMinSize(fyne.NewSize(3, rowH))
	if ev.Enabled {
		sideBar.FillColor = color.NRGBA{R: 60, G: 120, B: 240, A: 255}
	}

	// 启用 checkbox
	check := widget.NewCheck("", func(v bool) {
		onToggle(idx, v)
		pp.content.Refresh()
	})
	check.Checked = ev.Enabled

	// 事件名（可双击）+ 签名内联小字
	nameLabel := newDoubleTapLabel(ev.EventName, func() {
		if onOpen != nil {
			onOpen(ev)
		}
	})
	nameLabel.TextStyle = fyne.TextStyle{Bold: ev.Enabled}

	sigTxt := canvas.NewText("  "+kindLabelStr(ev.Kind), color.NRGBA{R: 140, G: 145, B: 175, A: 180})
	sigTxt.TextSize = 9

	nameRow := container.NewHBox(nameLabel, sigTxt)

	// 函数名输入框
	fnEntry := widget.NewEntry()
	fnEntry.SetText(ev.HandlerFn)
	fnEntry.PlaceHolder = "函数名"
	fnEntry.OnChanged = func(v string) { onRename(idx, v) }

	row := container.NewBorder(nil, nil,
		container.NewHBox(sideBar, container.NewGridWrap(fyne.NewSize(30, rowH), check)),
		container.NewGridWrap(fyne.NewSize(120, rowH), fnEntry),
		container.NewCenter(nameRow),
	)

	sep := canvas.NewRectangle(color.NRGBA{R: 210, G: 215, B: 235, A: 255})
	sep.SetMinSize(fyne.NewSize(0, 1))

	return container.NewVBox(container.NewStack(rowBg, row), sep)
}

// ─────────────────────────────────────────────────────────────────────────────
// 辅助
// ─────────────────────────────────────────────────────────────────────────────

func (pp *PropertiesPanel) sectionHeader(title string) fyne.CanvasObject {
	bg := canvas.NewRectangle(color.NRGBA{R: 38, G: 42, B: 80, A: 255})
	bg.SetMinSize(fyne.NewSize(0, 30))
	txt := canvas.NewText(title, color.NRGBA{R: 215, G: 225, B: 255, A: 255})
	txt.TextSize = 12
	txt.TextStyle = fyne.TextStyle{Bold: true}
	return container.NewStack(bg, container.NewPadded(txt))
}

func (pp *PropertiesPanel) notifyModified(dw *model.DesignWidget) {
	if pp.dc != nil {
		pp.dc.Refresh()
	}
	if pp.OnModified != nil {
		pp.OnModified(dw)
	}
}

func boolVal(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func isUniversalEvent(name string) bool {
	return map[string]bool{
		"Click": true, "DoubleClick": true,
		"MouseEnter": true, "MouseLeave": true, "MouseMove": true,
		"GotFocus": true, "LostFocus": true,
		"KeyDown": true, "KeyUp": true, "KeyPress": true,
	}[name]
}

func kindLabelStr(k model.EventKind) string {
	switch k {
	case model.EventKindVoid:
		return "func()"
	case model.EventKindString:
		return "func(value string)"
	case model.EventKindBool:
		return "func(checked bool)"
	case model.EventKindFloat:
		return "func(value float64)"
	case model.EventKindInt:
		return "func(index int)"
	case model.EventKindKey:
		return "func(*fyne.KeyEvent)"
	case model.EventKindMouse:
		return "func(*desktop.MouseEvent)"
	case model.EventKindTableCell:
		return "func(row, col int)"
	case model.EventKindTableChange:
		return "func(row, col int, val string)"
	}
	return "func()"
}

// ─────────────────────────────────────────────────────────────────────────────
// doubleTapLabel — 可双击的标签
// ─────────────────────────────────────────────────────────────────────────────

type doubleTapLabel struct {
	widget.Label
	onDoubleTap func()
}

func newDoubleTapLabel(text string, onDoubleTap func()) *doubleTapLabel {
	l := &doubleTapLabel{onDoubleTap: onDoubleTap}
	l.ExtendBaseWidget(l)
	l.SetText(text)
	return l
}

func (l *doubleTapLabel) DoubleTapped(_ *fyne.PointEvent) {
	if l.onDoubleTap != nil {
		l.onDoubleTap()
	}
}

func (l *doubleTapLabel) Tapped(_ *fyne.PointEvent) {}
