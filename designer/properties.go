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
	pages := []fyne.CanvasObject{
		pp.buildFormPropsTab(form),
		pp.buildEventsTab(form.Events, func(ev model.EventBinding) {
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
		}),
	}
	tabBar := pp.buildTabBar([]string{"属性", "事件"}, pages)
	return container.NewBorder(title, nil, nil, nil, tabBar)
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
	pages := []fyne.CanvasObject{
		pp.buildPropsTab(dw),
		pp.buildEventsTab(dw.Events,
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
		),
		pp.buildLayoutTab(dw),
	}
	tabBar := pp.buildTabBar([]string{"属性", "事件", "布局"}, pages)
	return container.NewBorder(title, nil, nil, nil, tabBar)
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
// 自绘 Tab 栏（不受系统主题影响，颜色固定）
// ─────────────────────────────────────────────────────────────────────────────

// buildTabBar 返回带有自绘 tab 头的页面切换容器
func (pp *PropertiesPanel) buildTabBar(labels []string, pages []fyne.CanvasObject) fyne.CanvasObject {
	if len(labels) == 0 || len(labels) != len(pages) {
		return container.NewStack()
	}

	// 颜色常量（固定，不跟系统主题走）
	const (
		tabH     = float32(28)
		tabTextS = float32(11)
	)
	colBg        := color.NRGBA{R: 235, G: 237, B: 248, A: 255} // tab 栏背景
	colSel       := color.NRGBA{R: 55, G: 80, B: 180, A: 255}   // 选中 tab 背景
	colUnsel     := color.NRGBA{R: 0, G: 0, B: 0, A: 0}         // 未选中 transparent
	colTextSel   := color.NRGBA{R: 255, G: 255, B: 255, A: 255} // 选中文字白
	colTextUnsel := color.NRGBA{R: 40, G: 50, B: 110, A: 255}   // 未选文字深蓝
	colUnderline := color.NRGBA{R: 55, G: 80, B: 180, A: 255}   // 底部指示线

	currentPage := 0
	pageHolder := container.NewStack(pages[0])

	var tabObjs []*fyne.Container // 每个 tab 的容器（可 refresh）
	var tabBgs   []*canvas.Rectangle
	var tabTexts []*canvas.Text
	var tabLines []*canvas.Rectangle

	for i := range labels {
		i := i
		bg := canvas.NewRectangle(colUnsel)
		bg.SetMinSize(fyne.NewSize(0, tabH))
		if i == 0 {
			bg.FillColor = colSel
		}
		bg.CornerRadius = 3

		txt := canvas.NewText(labels[i], colTextUnsel)
		txt.TextSize = tabTextS
		txt.TextStyle = fyne.TextStyle{Bold: true}
		if i == 0 {
			txt.Color = colTextSel
		}
		txt.Alignment = fyne.TextAlignCenter

		line := canvas.NewRectangle(colUnsel)
		line.SetMinSize(fyne.NewSize(0, 2))
		if i == 0 {
			line.FillColor = colUnderline
		}

		cell := container.NewStack(bg, container.NewCenter(txt))

		tabBgs = append(tabBgs, bg)
		tabTexts = append(tabTexts, txt)
		tabLines = append(tabLines, line)

		col := container.NewBorder(nil, line, nil, nil, cell)
		tabObjs = append(tabObjs, col)

		// 点击切换
		tap := newTappable(col, func() {
			if currentPage == i {
				return
			}
			// 还原旧 tab
			tabBgs[currentPage].FillColor = colUnsel
			tabTexts[currentPage].Color = colTextUnsel
			tabTexts[currentPage].TextStyle.Bold = false
			tabLines[currentPage].FillColor = colUnsel
			tabObjs[currentPage].Refresh()

			// 激活新 tab
			currentPage = i
			tabBgs[i].FillColor = colSel
			tabTexts[i].Color = colTextSel
			tabTexts[i].TextStyle.Bold = true
			tabLines[i].FillColor = colUnderline
			tabObjs[i].Refresh()

			pageHolder.Objects = []fyne.CanvasObject{pages[i]}
			pageHolder.Refresh()
		})
		_ = tap
	}

	// tab 行：每个 tab 等宽
	tabRow := container.NewGridWithColumns(len(labels))
	for _, tc := range tabObjs {
		tabRow.Add(tc)
	}
	tabRowBg := canvas.NewRectangle(colBg)
	tabRowBg.SetMinSize(fyne.NewSize(0, tabH+4))

	header := container.NewStack(tabRowBg, tabRow)
	return container.NewBorder(header, nil, nil, nil, pageHolder)
}

// tappable 包装任意 fyne.CanvasObject 使其可接收点击
type tappable struct {
	widget.BaseWidget
	inner   fyne.CanvasObject
	onTap   func()
}

func newTappable(inner fyne.CanvasObject, onTap func()) *tappable {
	t := &tappable{inner: inner, onTap: onTap}
	t.ExtendBaseWidget(t)
	return t
}
func (t *tappable) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.inner)
}
func (t *tappable) Tapped(_ *fyne.PointEvent) {
	if t.onTap != nil {
		t.onTap()
	}
}
func (t *tappable) TappedSecondary(_ *fyne.PointEvent) {}

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
	hBg := canvas.NewRectangle(color.NRGBA{R: 55, G: 68, B: 145, A: 255})
	hBg.SetMinSize(fyne.NewSize(0, 20))

	mkHdr := func(t string) *canvas.Text {
		txt := canvas.NewText(t, color.NRGBA{R: 220, G: 228, B: 255, A: 255})
		txt.TextSize = 9
		txt.TextStyle = fyne.TextStyle{Bold: true}
		return txt
	}
	row := container.NewBorder(nil, nil,
		mkHdr("启用"),
		nil,
		mkHdr("事件名  /  处理函数"),
	)
	return container.NewStack(hBg, container.NewPadded(row))
}

// eventRow 单个事件行（两行紧凑布局，避免函数名被截断）
func (pp *PropertiesPanel) eventRow(
	idx int,
	ev model.EventBinding,
	onOpen func(ev model.EventBinding),
	onToggle func(i int, v bool),
	onRename func(i int, v string),
) fyne.CanvasObject {
	// 行背景色
	bgCol := color.NRGBA{R: 250, G: 251, B: 255, A: 255}
	if ev.Enabled {
		bgCol = color.NRGBA{R: 232, G: 240, B: 255, A: 255}
	}
	rowBg := canvas.NewRectangle(bgCol)

	// 左色条
	var sideCol color.Color = color.NRGBA{A: 0}
	if ev.Enabled {
		sideCol = color.NRGBA{R: 60, G: 120, B: 240, A: 255}
	}
	sideBar := canvas.NewRectangle(sideCol)
	sideBar.SetMinSize(fyne.NewSize(3, 0))

	// 启用 checkbox（紧凑）
	check := widget.NewCheck("", func(v bool) {
		onToggle(idx, v)
		pp.content.Refresh()
	})
	check.Checked = ev.Enabled

	// 第一行：事件名（可双击）+ 签名小字
	nameLabel := newDoubleTapLabel(ev.EventName, func() {
		if onOpen != nil {
			onOpen(ev)
		}
	})
	nameLabel.TextStyle = fyne.TextStyle{Bold: ev.Enabled}

	sigTxt := canvas.NewText(kindLabelStr(ev.Kind), color.NRGBA{R: 130, G: 140, B: 175, A: 200})
	sigTxt.TextSize = 9

	topRow := container.NewBorder(nil, nil, nil,
		container.NewGridWrap(fyne.NewSize(28, 22), check),
		nameLabel,
	)

	// 第二行：函数名输入框（撑满宽度）
	fnEntry := widget.NewEntry()
	fnEntry.SetText(ev.HandlerFn)
	fnEntry.PlaceHolder = "处理函数名"
	fnEntry.OnChanged = func(v string) { onRename(idx, v) }
	fnEntry.TextStyle = fyne.TextStyle{Monospace: true}

	bottomRow := container.NewBorder(nil, nil,
		container.NewGridWrap(fyne.NewSize(10, 14), sigTxt),
		nil,
		fnEntry,
	)

	inner := container.NewVBox(topRow, bottomRow)

	sep := canvas.NewRectangle(color.NRGBA{R: 210, G: 215, B: 235, A: 255})
	sep.SetMinSize(fyne.NewSize(0, 1))

	content := container.NewBorder(nil, nil, sideBar, nil, container.NewPadded(inner))

	return container.NewVBox(container.NewStack(rowBg, content), sep)
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
