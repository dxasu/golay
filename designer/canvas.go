package designer

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	"github.com/mico/golay/model"
)

// DesignCanvas 是设计器的核心画布，支持拖拽放置控件
type DesignCanvas struct {
	widget.BaseWidget

	currentForm       *model.FormDef // 当前绑定的 FormDef，控件存储在其中
	selected          *model.DesignWidget
	dragging          bool
	resizing          bool
	dragOffX          float32
	dragOffY          float32
	resizeOrigW       float32
	resizeOrigH       float32
	resizeStartMouseX float32
	resizeStartMouseY float32

	OnSelect      func(dw *model.DesignWidget)
	OnDeselect    func()
	OnModified    func()
	OnDoubleClick func(dw *model.DesignWidget)

	gridSize float32
	showGrid bool

	CanvasW float32
	CanvasH float32
}

// NewDesignCanvas 创建新的设计画布
func NewDesignCanvas() *DesignCanvas {
	dc := &DesignCanvas{
		gridSize: 10,
		showGrid: true,
		CanvasW:  800,
		CanvasH:  600,
	}
	dc.ExtendBaseWidget(dc)
	return dc
}

// SetForm 切换当前编辑的 Form，并清除选中状态
func (dc *DesignCanvas) SetForm(form *model.FormDef) {
	dc.currentForm = form
	dc.selected = nil
	if dc.CanvasW < form.Width {
		dc.CanvasW = form.Width
	}
	if dc.CanvasH < form.Height {
		dc.CanvasH = form.Height
	}
	dc.Refresh()
}

// GetCurrentForm 返回当前绑定的 FormDef
func (dc *DesignCanvas) GetCurrentForm() *model.FormDef {
	return dc.currentForm
}

func (dc *DesignCanvas) widgets() []*model.DesignWidget {
	if dc.currentForm == nil {
		return nil
	}
	return dc.currentForm.Widgets
}

func (dc *DesignCanvas) AddWidget(dw *model.DesignWidget) {
	if dc.currentForm == nil {
		return
	}
	dc.currentForm.Widgets = append(dc.currentForm.Widgets, dw)
	dc.Refresh()
	if dc.OnModified != nil {
		dc.OnModified()
	}
}

func (dc *DesignCanvas) DeleteSelected() {
	if dc.selected == nil || dc.currentForm == nil {
		return
	}
	ws := dc.currentForm.Widgets
	for i, w := range ws {
		if w == dc.selected {
			dc.currentForm.Widgets = append(ws[:i], ws[i+1:]...)
			break
		}
	}
	dc.selected = nil
	dc.Refresh()
	if dc.OnDeselect != nil {
		dc.OnDeselect()
	}
	if dc.OnModified != nil {
		dc.OnModified()
	}
}

func (dc *DesignCanvas) GetWidgets() []*model.DesignWidget {
	return dc.widgets()
}

func (dc *DesignCanvas) GetSelected() *model.DesignWidget {
	return dc.selected
}

func (dc *DesignCanvas) SelectWidget(dw *model.DesignWidget) {
	dc.selected = dw
	dc.Refresh()
	if dw != nil && dc.OnSelect != nil {
		dc.OnSelect(dw)
	}
}

func (dc *DesignCanvas) ClearAll() {
	if dc.currentForm == nil {
		return
	}
	dc.currentForm.Widgets = dc.currentForm.Widgets[:0]
	dc.selected = nil
	dc.Refresh()
	if dc.OnDeselect != nil {
		dc.OnDeselect()
	}
}

func (dc *DesignCanvas) snapToGrid(v float32) float32 {
	if dc.gridSize <= 1 {
		return v
	}
	return float32(int(v/dc.gridSize)) * dc.gridSize
}

func (dc *DesignCanvas) widgetAt(pos fyne.Position) *model.DesignWidget {
	ws := dc.widgets()
	for i := len(ws) - 1; i >= 0; i-- {
		w := ws[i]
		if pos.X >= w.X && pos.X <= w.X+w.W &&
			pos.Y >= w.Y && pos.Y <= w.Y+w.H {
			return w
		}
	}
	return nil
}

func (dc *DesignCanvas) isResizeHandle(dw *model.DesignWidget, pos fyne.Position) bool {
	const handleSize = float32(12)
	return pos.X >= dw.X+dw.W-handleSize &&
		pos.X <= dw.X+dw.W &&
		pos.Y >= dw.Y+dw.H-handleSize &&
		pos.Y <= dw.Y+dw.H
}

// ── Fyne 接口 ────────────────────────────────────────────────────────────────

func (dc *DesignCanvas) CreateRenderer() fyne.WidgetRenderer {
	return newDesignCanvasRenderer(dc)
}

func (dc *DesignCanvas) Tapped(ev *fyne.PointEvent) {
	hit := dc.widgetAt(ev.Position)
	if hit != nil {
		dc.selected = hit
		dc.Refresh()
		if dc.OnSelect != nil {
			dc.OnSelect(hit)
		}
	} else {
		dc.selected = nil
		dc.Refresh()
		if dc.OnDeselect != nil {
			dc.OnDeselect()
		}
	}
}

// DoubleTapped 双击控件时触发（打开编辑器等操作）
func (dc *DesignCanvas) DoubleTapped(ev *fyne.PointEvent) {
	hit := dc.widgetAt(ev.Position)
	if hit != nil && dc.OnDoubleClick != nil {
		dc.selected = hit
		dc.Refresh()
		dc.OnDoubleClick(hit)
	}
}

func (dc *DesignCanvas) DragEnd() {
	dc.dragging = false
	dc.resizing = false
	if dc.OnModified != nil {
		dc.OnModified()
	}
}

func (dc *DesignCanvas) Dragged(ev *fyne.DragEvent) {
	if dc.selected == nil {
		return
	}
	if dc.resizing {
		newW := dc.resizeOrigW + (ev.Position.X - dc.resizeStartMouseX)
		newH := dc.resizeOrigH + (ev.Position.Y - dc.resizeStartMouseY)
		if newW < 24 {
			newW = 24
		}
		if newH < 12 {
			newH = 12
		}
		dc.selected.W = dc.snapToGrid(newW)
		dc.selected.H = dc.snapToGrid(newH)
		dc.Refresh()
		return
	}
	if !dc.dragging {
		if dc.isResizeHandle(dc.selected, ev.Position) {
			dc.resizing = true
			dc.resizeOrigW = dc.selected.W
			dc.resizeOrigH = dc.selected.H
			dc.resizeStartMouseX = ev.Position.X
			dc.resizeStartMouseY = ev.Position.Y
		} else {
			hit := dc.widgetAt(ev.Position)
			if hit != nil {
				if hit != dc.selected {
					dc.selected = hit
					if dc.OnSelect != nil {
						dc.OnSelect(hit)
					}
				}
				dc.dragging = true
				dc.dragOffX = ev.Position.X - dc.selected.X
				dc.dragOffY = ev.Position.Y - dc.selected.Y
			}
		}
		return
	}
	newX := dc.snapToGrid(ev.Position.X - dc.dragOffX)
	newY := dc.snapToGrid(ev.Position.Y - dc.dragOffY)
	if newX < 0 {
		newX = 0
	}
	if newY < 0 {
		newY = 0
	}
	dc.selected.X = newX
	dc.selected.Y = newY
	dc.Refresh()
}

// ── Renderer ─────────────────────────────────────────────────────────────────

type designCanvasRenderer struct {
	dc      *DesignCanvas
	bg      *canvas.Rectangle
	objects []fyne.CanvasObject
}

func newDesignCanvasRenderer(dc *DesignCanvas) *designCanvasRenderer {
	bg := canvas.NewRectangle(color.NRGBA{R: 250, G: 251, B: 255, A: 255})
	return &designCanvasRenderer{dc: dc, bg: bg}
}

func (r *designCanvasRenderer) Layout(size fyne.Size) { r.bg.Resize(size) }
func (r *designCanvasRenderer) MinSize() fyne.Size    { return fyne.NewSize(600, 480) }
func (r *designCanvasRenderer) Destroy()              {}

func (r *designCanvasRenderer) Refresh() {
	r.buildObjects()
	canvas.Refresh(r.dc)
}

func (r *designCanvasRenderer) Objects() []fyne.CanvasObject {
	r.buildObjects()
	return r.objects
}

func (r *designCanvasRenderer) buildObjects() {
	objs := []fyne.CanvasObject{r.bg}
	if r.dc.showGrid {
		objs = append(objs, r.buildGrid(r.dc.Size())...)
	}
	for _, dw := range r.dc.widgets() {
		objs = append(objs, r.renderWidget(dw)...)
	}
	r.objects = objs
}

// buildGrid 使用点阵网格（比线网格更美观）
func (r *designCanvasRenderer) buildGrid(size fyne.Size) []fyne.CanvasObject {
	objs := []fyne.CanvasObject{}
	gs := r.dc.gridSize * 2 // 点间距为网格的2倍
	dotColor := color.NRGBA{R: 180, G: 185, B: 210, A: 200}
	for x := gs; x < size.Width; x += gs {
		for y := gs; y < size.Height; y += gs {
			dot := canvas.NewRectangle(dotColor)
			dot.Move(fyne.NewPos(x-0.5, y-0.5))
			dot.Resize(fyne.NewSize(1.5, 1.5))
			objs = append(objs, dot)
		}
	}
	return objs
}

// renderWidget 根据控件类型分派到专用渲染函数
func (r *designCanvasRenderer) renderWidget(dw *model.DesignWidget) []fyne.CanvasObject {
	switch dw.Type {
	case model.WidgetPanel:
		return r.renderPanel(dw)
	case model.WidgetGroupBox:
		return r.renderGroupBox(dw)
	case model.WidgetTabControl:
		return r.renderTabControl(dw)
	case model.WidgetLabel:
		return r.renderLabel(dw)
	case model.WidgetLinkLabel:
		return r.renderLinkLabel(dw)
	case model.WidgetRichTextBox:
		return r.renderRichTextBox(dw)
	case model.WidgetTextBox:
		return r.renderTextBox(dw)
	case model.WidgetMultiLineEntry:
		return r.renderMultiLineEntry(dw)
	case model.WidgetComboBox:
		return r.renderComboBox(dw)
	case model.WidgetCheckBox:
		return r.renderCheckBox(dw)
	case model.WidgetRadioButton:
		return r.renderRadioButton(dw)
	case model.WidgetDateTimePicker:
		return r.renderDateTimePicker(dw)
	case model.WidgetButton:
		return r.renderButton(dw)
	case model.WidgetToolStrip:
		return r.renderToolStrip(dw)
	case model.WidgetDataGridView:
		return r.renderDataGridView(dw)
	case model.WidgetListBox:
		return r.renderListBox(dw)
	case model.WidgetListView:
		return r.renderListView(dw)
	case model.WidgetTreeView:
		return r.renderTreeView(dw)
	case model.WidgetProgressBar:
		return r.renderProgressBar(dw)
	case model.WidgetStatusStrip:
		return r.renderStatusStrip(dw)
	case model.WidgetPictureBox:
		return r.renderPictureBox(dw)
	case model.WidgetSeparator:
		return r.renderSeparator(dw)
	case model.WidgetSlider:
		return r.renderSlider(dw)
	}
	return r.renderGeneric(dw)
}

// ── 通用辅助 ─────────────────────────────────────────────────────────────────

func (r *designCanvasRenderer) baseBg(dw *model.DesignWidget, bg color.Color) *canvas.Rectangle {
	rect := canvas.NewRectangle(bg)
	rect.Move(fyne.NewPos(dw.X, dw.Y))
	rect.Resize(fyne.NewSize(dw.W, dw.H))
	return rect
}

func (r *designCanvasRenderer) border(dw *model.DesignWidget, c color.Color, stroke float32) *canvas.Rectangle {
	rect := canvas.NewRectangle(color.Transparent)
	rect.StrokeColor = c
	rect.StrokeWidth = stroke
	rect.Move(fyne.NewPos(dw.X, dw.Y))
	rect.Resize(fyne.NewSize(dw.W, dw.H))
	return rect
}

func (r *designCanvasRenderer) selectionBorder(dw *model.DesignWidget) []fyne.CanvasObject {
	if dw != r.dc.selected {
		return nil
	}
	b := canvas.NewRectangle(color.Transparent)
	b.StrokeColor = color.NRGBA{R: 30, G: 120, B: 255, A: 255}
	b.StrokeWidth = 2
	b.CornerRadius = 3
	b.Move(fyne.NewPos(dw.X-1, dw.Y-1))
	b.Resize(fyne.NewSize(dw.W+2, dw.H+2))

	handle := canvas.NewRectangle(color.NRGBA{R: 30, G: 120, B: 255, A: 255})
	handle.Move(fyne.NewPos(dw.X+dw.W-9, dw.Y+dw.H-9))
	handle.Resize(fyne.NewSize(9, 9))
	return []fyne.CanvasObject{b, handle}
}

func (r *designCanvasRenderer) text(t string, x, y, w, h, size float32, c color.Color, align fyne.TextAlign, bold bool) *canvas.Text {
	ct := canvas.NewText(t, c)
	ct.TextSize = size
	ct.Alignment = align
	ct.TextStyle = fyne.TextStyle{Bold: bold}
	ct.Move(fyne.NewPos(x, y))
	ct.Resize(fyne.NewSize(w, h))
	return ct
}

func (r *designCanvasRenderer) shadow(dw *model.DesignWidget) *canvas.Rectangle {
	s := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 20})
	s.Move(fyne.NewPos(dw.X+2, dw.Y+2))
	s.Resize(fyne.NewSize(dw.W, dw.H))
	return s
}

func clampText(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "…"
}

// ── 各控件渲染 ────────────────────────────────────────────────────────────────

func (r *designCanvasRenderer) renderGeneric(dw *model.DesignWidget) []fyne.CanvasObject {
	info := model.GetWidgetTypeInfo(dw.Type)
	bgC := color.NRGBA{R: info.Color[0], G: info.Color[1], B: info.Color[2], A: 255}
	objs := []fyne.CanvasObject{r.shadow(dw), r.baseBg(dw, bgC)}
	objs = append(objs, r.text(dw.Name, dw.X+4, dw.Y+4, dw.W-8, 12, 10, color.NRGBA{R: 100, G: 100, B: 120, A: 200}, fyne.TextAlignLeading, false))
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderPanel(dw *model.DesignWidget) []fyne.CanvasObject {
	objs := []fyne.CanvasObject{
		r.baseBg(dw, color.NRGBA{R: 235, G: 238, B: 248, A: 200}),
		r.border(dw, color.NRGBA{R: 170, G: 180, B: 210, A: 255}, 1),
	}
	objs = append(objs, r.text("Panel: "+dw.Name, dw.X+4, dw.Y+3, dw.W-8, 13, 10, color.NRGBA{R: 100, G: 110, B: 150, A: 200}, fyne.TextAlignLeading, false))
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderGroupBox(dw *model.DesignWidget) []fyne.CanvasObject {
	title := dw.GetProperty("Text")
	if title == "" {
		title = dw.Name
	}
	// 主体
	body := canvas.NewRectangle(color.NRGBA{R: 240, G: 242, B: 252, A: 180})
	body.StrokeColor = color.NRGBA{R: 150, G: 160, B: 200, A: 255}
	body.StrokeWidth = 1
	body.Move(fyne.NewPos(dw.X, dw.Y+8))
	body.Resize(fyne.NewSize(dw.W, dw.H-8))
	// 标题背景
	titleBg := canvas.NewRectangle(color.NRGBA{R: 240, G: 242, B: 252, A: 255})
	titleBg.Move(fyne.NewPos(dw.X+8, dw.Y))
	titleBg.Resize(fyne.NewSize(float32(len([]rune(title)))*8+8, 16))
	titleText := r.text(title, dw.X+12, dw.Y, float32(len([]rune(title)))*8+4, 16, 11, color.NRGBA{R: 60, G: 80, B: 160, A: 255}, fyne.TextAlignLeading, true)
	objs := []fyne.CanvasObject{body, titleBg, titleText}
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderTabControl(dw *model.DesignWidget) []fyne.CanvasObject {
	tabStr := dw.GetProperty("Tabs")
	tabs := strings.Split(tabStr, ",")
	body := r.baseBg(dw, color.NRGBA{R: 240, G: 242, B: 252, A: 255})
	bodyBorder := r.border(dw, color.NRGBA{R: 160, G: 170, B: 210, A: 255}, 1)
	objs := []fyne.CanvasObject{body, bodyBorder}

	tabW := float32(80)
	tabH := float32(24)
	for i, tab := range tabs {
		tab = strings.TrimSpace(tab)
		isFirst := i == 0
		tabBg := canvas.NewRectangle(color.NRGBA{R: 220, G: 225, B: 245, A: 255})
		if isFirst {
			tabBg.FillColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		}
		tabBg.StrokeColor = color.NRGBA{R: 160, G: 170, B: 210, A: 255}
		tabBg.StrokeWidth = 1
		tabBg.Move(fyne.NewPos(dw.X+float32(i)*tabW, dw.Y))
		tabBg.Resize(fyne.NewSize(tabW-1, tabH))
		tabText := r.text(clampText(tab, 8), dw.X+float32(i)*tabW+4, dw.Y+4, tabW-8, 14, 11,
			color.NRGBA{R: 50, G: 60, B: 120, A: 255}, fyne.TextAlignLeading, isFirst)
		objs = append(objs, tabBg, tabText)
	}
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderLabel(dw *model.DesignWidget) []fyne.CanvasObject {
	txt := dw.GetProperty("Text")
	bold := dw.GetProperty("Bold") == "true"
	alignStr := dw.GetProperty("TextAlign")
	align := fyne.TextAlignLeading
	switch alignStr {
	case "Center":
		align = fyne.TextAlignCenter
	case "Right":
		align = fyne.TextAlignTrailing
	}
	objs := []fyne.CanvasObject{
		r.baseBg(dw, color.NRGBA{R: 245, G: 245, B: 248, A: 0}),
		r.text(txt, dw.X+2, dw.Y+(dw.H-14)/2, dw.W-4, 14, 13, color.NRGBA{R: 30, G: 30, B: 30, A: 255}, align, bold),
	}
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderLinkLabel(dw *model.DesignWidget) []fyne.CanvasObject {
	txt := dw.GetProperty("Text")
	objs := []fyne.CanvasObject{
		r.baseBg(dw, color.Transparent),
		r.text(txt, dw.X+2, dw.Y+(dw.H-14)/2, dw.W-4, 14, 13, color.NRGBA{R: 0, G: 80, B: 200, A: 255}, fyne.TextAlignLeading, false),
	}
	// 下划线模拟
	underline := canvas.NewRectangle(color.NRGBA{R: 0, G: 80, B: 200, A: 200})
	underline.Move(fyne.NewPos(dw.X+2, dw.Y+dw.H-3))
	underline.Resize(fyne.NewSize(dw.W-4, 1))
	objs = append(objs, underline)
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderRichTextBox(dw *model.DesignWidget) []fyne.CanvasObject {
	objs := []fyne.CanvasObject{
		r.shadow(dw),
		r.baseBg(dw, color.NRGBA{R: 255, G: 255, B: 250, A: 255}),
		r.border(dw, color.NRGBA{R: 180, G: 180, B: 200, A: 255}, 1),
	}
	objs = append(objs, r.text("RichText", dw.X+4, dw.Y+4, dw.W-8, 13, 10, color.NRGBA{R: 140, G: 140, B: 160, A: 200}, fyne.TextAlignLeading, false))
	objs = append(objs, r.text(dw.GetProperty("Text"), dw.X+4, dw.Y+18, dw.W-8, dw.H-22, 12, color.NRGBA{R: 30, G: 30, B: 30, A: 255}, fyne.TextAlignLeading, false))
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderTextBox(dw *model.DesignWidget) []fyne.CanvasObject {
	txt := dw.GetProperty("Text")
	ph := dw.GetProperty("PlaceHolder")
	display := txt
	textColor := color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	if display == "" && ph != "" {
		display = ph
		textColor = color.NRGBA{R: 160, G: 160, B: 170, A: 255}
	}
	if dw.GetProperty("Password") == "true" && len(display) > 0 {
		display = strings.Repeat("●", len([]rune(display)))
	}
	objs := []fyne.CanvasObject{
		r.shadow(dw),
		r.baseBg(dw, color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
		r.border(dw, color.NRGBA{R: 170, G: 170, B: 190, A: 255}, 1),
		r.text(display, dw.X+6, dw.Y+(dw.H-13)/2, dw.W-12, 13, 12, textColor, fyne.TextAlignLeading, false),
	}
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderMultiLineEntry(dw *model.DesignWidget) []fyne.CanvasObject {
	txt := dw.GetProperty("Text")
	if txt == "" {
		txt = dw.GetProperty("PlaceHolder")
	}
	objs := []fyne.CanvasObject{
		r.shadow(dw),
		r.baseBg(dw, color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
		r.border(dw, color.NRGBA{R: 170, G: 170, B: 190, A: 255}, 1),
		r.text(txt, dw.X+6, dw.Y+6, dw.W-12, dw.H-12, 12, color.NRGBA{R: 30, G: 30, B: 30, A: 255}, fyne.TextAlignLeading, false),
	}
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderComboBox(dw *model.DesignWidget) []fyne.CanvasObject {
	items := dw.GetProperty("Items")
	first := ""
	if parts := strings.SplitN(items, ",", 2); len(parts) > 0 {
		first = strings.TrimSpace(parts[0])
	}
	objs := []fyne.CanvasObject{
		r.shadow(dw),
		r.baseBg(dw, color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
		r.border(dw, color.NRGBA{R: 170, G: 170, B: 190, A: 255}, 1),
		r.text(first, dw.X+6, dw.Y+(dw.H-13)/2, dw.W-28, 13, 12, color.NRGBA{R: 30, G: 30, B: 30, A: 255}, fyne.TextAlignLeading, false),
	}
	// 下拉箭头
	arrowBg := canvas.NewRectangle(color.NRGBA{R: 200, G: 205, B: 225, A: 255})
	arrowBg.Move(fyne.NewPos(dw.X+dw.W-22, dw.Y+1))
	arrowBg.Resize(fyne.NewSize(21, dw.H-2))
	arrowText := r.text("▼", dw.X+dw.W-20, dw.Y+(dw.H-13)/2, 16, 13, 10, color.NRGBA{R: 80, G: 80, B: 100, A: 255}, fyne.TextAlignCenter, false)
	objs = append(objs, arrowBg, arrowText)
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderCheckBox(dw *model.DesignWidget) []fyne.CanvasObject {
	txt := dw.GetProperty("Text")
	checked := dw.GetProperty("Checked") == "true"
	boxBg := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	boxBorder := color.NRGBA{R: 130, G: 130, B: 150, A: 255}
	box := canvas.NewRectangle(boxBg)
	box.StrokeColor = boxBorder
	box.StrokeWidth = 1
	box.Move(fyne.NewPos(dw.X+4, dw.Y+(dw.H-14)/2))
	box.Resize(fyne.NewSize(14, 14))
	objs := []fyne.CanvasObject{r.baseBg(dw, color.Transparent), box}
	if checked {
		check := r.text("✓", dw.X+5, dw.Y+(dw.H-14)/2, 14, 14, 12, color.NRGBA{R: 30, G: 130, B: 30, A: 255}, fyne.TextAlignCenter, true)
		objs = append(objs, check)
	}
	objs = append(objs, r.text(txt, dw.X+22, dw.Y+(dw.H-13)/2, dw.W-26, 13, 12, color.NRGBA{R: 30, G: 30, B: 30, A: 255}, fyne.TextAlignLeading, false))
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderRadioButton(dw *model.DesignWidget) []fyne.CanvasObject {
	optsStr := dw.GetProperty("Options")
	opts := strings.Split(optsStr, ",")
	selected := dw.GetProperty("Selected")
	objs := []fyne.CanvasObject{r.baseBg(dw, color.Transparent)}
	rowH := dw.H / float32(len(opts)+1)
	if rowH < 20 {
		rowH = 20
	}
	for i, opt := range opts {
		opt = strings.TrimSpace(opt)
		y := dw.Y + float32(i)*rowH + (rowH-14)/2
		// 圆形
		circle := canvas.NewCircle(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		circle.StrokeColor = color.NRGBA{R: 130, G: 130, B: 150, A: 255}
		circle.StrokeWidth = 1
		circle.Move(fyne.NewPos(dw.X+4, y))
		circle.Resize(fyne.NewSize(14, 14))
		objs = append(objs, circle)
		if opt == selected {
			dot := canvas.NewCircle(color.NRGBA{R: 30, G: 120, B: 200, A: 255})
			dot.Move(fyne.NewPos(dw.X+8, y+4))
			dot.Resize(fyne.NewSize(6, 6))
			objs = append(objs, dot)
		}
		objs = append(objs, r.text(opt, dw.X+22, y, dw.W-26, 13, 12, color.NRGBA{R: 30, G: 30, B: 30, A: 255}, fyne.TextAlignLeading, false))
	}
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderDateTimePicker(dw *model.DesignWidget) []fyne.CanvasObject {
	val := dw.GetProperty("Value")
	if val == "" {
		val = "2026-01-01"
	}
	objs := []fyne.CanvasObject{
		r.shadow(dw),
		r.baseBg(dw, color.NRGBA{R: 255, G: 248, B: 235, A: 255}),
		r.border(dw, color.NRGBA{R: 200, G: 170, B: 130, A: 255}, 1),
		r.text(val, dw.X+6, dw.Y+(dw.H-13)/2, dw.W-32, 13, 12, color.NRGBA{R: 60, G: 40, B: 20, A: 255}, fyne.TextAlignLeading, false),
	}
	calBtn := canvas.NewRectangle(color.NRGBA{R: 220, G: 190, B: 150, A: 255})
	calBtn.Move(fyne.NewPos(dw.X+dw.W-26, dw.Y+1))
	calBtn.Resize(fyne.NewSize(25, dw.H-2))
	calText := r.text("📅", dw.X+dw.W-24, dw.Y+(dw.H-13)/2-1, 22, 14, 11, color.NRGBA{R: 80, G: 60, B: 30, A: 255}, fyne.TextAlignCenter, false)
	objs = append(objs, calBtn, calText)
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderButton(dw *model.DesignWidget) []fyne.CanvasObject {
	txt := dw.GetProperty("Text")
	style := dw.GetProperty("Style")
	bgC := color.NRGBA{R: 70, G: 130, B: 200, A: 255}
	txtC := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	switch style {
	case "Primary":
		bgC = color.NRGBA{R: 40, G: 100, B: 220, A: 255}
	case "Warning":
		bgC = color.NRGBA{R: 220, G: 140, B: 20, A: 255}
	case "Danger":
		bgC = color.NRGBA{R: 200, G: 50, B: 50, A: 255}
	case "Low":
		bgC = color.NRGBA{R: 220, G: 225, B: 240, A: 255}
		txtC = color.NRGBA{R: 60, G: 80, B: 140, A: 255}
	}
	bg := r.baseBg(dw, bgC)
	bg.CornerRadius = 4
	objs := []fyne.CanvasObject{r.shadow(dw), bg}
	objs = append(objs, r.text(txt, dw.X+4, dw.Y+(dw.H-13)/2, dw.W-8, 13, 12, txtC, fyne.TextAlignCenter, true))
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderToolStrip(dw *model.DesignWidget) []fyne.CanvasObject {
	items := strings.Split(dw.GetProperty("Items"), ",")
	objs := []fyne.CanvasObject{
		r.baseBg(dw, color.NRGBA{R: 235, G: 237, B: 245, A: 255}),
		r.border(dw, color.NRGBA{R: 190, G: 195, B: 215, A: 255}, 1),
	}
	x := dw.X + 4
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "|" {
			sep := canvas.NewRectangle(color.NRGBA{R: 180, G: 185, B: 205, A: 255})
			sep.Move(fyne.NewPos(x, dw.Y+4))
			sep.Resize(fyne.NewSize(1, dw.H-8))
			objs = append(objs, sep)
			x += 8
			continue
		}
		btnW := float32(len([]rune(item))*8 + 12)
		if btnW < 28 {
			btnW = 28
		}
		btn := canvas.NewRectangle(color.NRGBA{R: 210, G: 215, B: 235, A: 200})
		btn.CornerRadius = 3
		btn.Move(fyne.NewPos(x, dw.Y+3))
		btn.Resize(fyne.NewSize(btnW, dw.H-6))
		btnTxt := r.text(item, x+2, dw.Y+(dw.H-12)/2, btnW-4, 12, 11,
			color.NRGBA{R: 40, G: 50, B: 100, A: 255}, fyne.TextAlignCenter, false)
		objs = append(objs, btn, btnTxt)
		x += btnW + 2
	}
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderDataGridView(dw *model.DesignWidget) []fyne.CanvasObject {
	cols := strings.Split(dw.GetProperty("Columns"), ",")
	objs := []fyne.CanvasObject{
		r.shadow(dw),
		r.baseBg(dw, color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
		r.border(dw, color.NRGBA{R: 160, G: 170, B: 200, A: 255}, 1),
	}
	// 表头
	headerH := float32(24)
	headerBg := canvas.NewRectangle(color.NRGBA{R: 215, G: 220, B: 240, A: 255})
	headerBg.Move(fyne.NewPos(dw.X, dw.Y))
	headerBg.Resize(fyne.NewSize(dw.W, headerH))
	objs = append(objs, headerBg)
	colW := dw.W / float32(max(len(cols), 1))
	for i, col := range cols {
		col = strings.TrimSpace(col)
		x := dw.X + float32(i)*colW
		if i > 0 {
			div := canvas.NewRectangle(color.NRGBA{R: 170, G: 175, B: 205, A: 255})
			div.Move(fyne.NewPos(x, dw.Y))
			div.Resize(fyne.NewSize(1, dw.H))
			objs = append(objs, div)
		}
		objs = append(objs, r.text(col, x+4, dw.Y+(headerH-12)/2, colW-8, 12, 11,
			color.NRGBA{R: 40, G: 50, B: 100, A: 255}, fyne.TextAlignLeading, true))
	}
	// 模拟2行数据
	for row := 1; row <= 2; row++ {
		rowY := dw.Y + headerH + float32(row-1)*22
		if rowY+22 > dw.Y+dw.H {
			break
		}
		rowBg := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		if row%2 == 0 {
			rowBg = color.NRGBA{R: 245, G: 247, B: 255, A: 255}
		}
		rb := canvas.NewRectangle(rowBg)
		rb.Move(fyne.NewPos(dw.X+1, rowY))
		rb.Resize(fyne.NewSize(dw.W-2, 22))
		objs = append(objs, rb)
	}
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderListBox(dw *model.DesignWidget) []fyne.CanvasObject {
	items := strings.Split(dw.GetProperty("Items"), ",")
	objs := []fyne.CanvasObject{
		r.shadow(dw),
		r.baseBg(dw, color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
		r.border(dw, color.NRGBA{R: 160, G: 170, B: 200, A: 255}, 1),
	}
	rowH := float32(22)
	for i, item := range items {
		item = strings.TrimSpace(item)
		y := dw.Y + float32(i)*rowH
		if y+rowH > dw.Y+dw.H {
			break
		}
		bg := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		if i == 0 {
			bg = color.NRGBA{R: 190, G: 215, B: 255, A: 255}
		}
		rb := canvas.NewRectangle(bg)
		rb.Move(fyne.NewPos(dw.X+1, y))
		rb.Resize(fyne.NewSize(dw.W-2, rowH))
		objs = append(objs, rb)
		objs = append(objs, r.text(item, dw.X+6, y+(rowH-12)/2, dw.W-10, 12, 12,
			color.NRGBA{R: 20, G: 20, B: 40, A: 255}, fyne.TextAlignLeading, false))
	}
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderListView(dw *model.DesignWidget) []fyne.CanvasObject {
	cols := strings.Split(dw.GetProperty("Columns"), ",")
	objs := []fyne.CanvasObject{
		r.shadow(dw),
		r.baseBg(dw, color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
		r.border(dw, color.NRGBA{R: 160, G: 170, B: 200, A: 255}, 1),
	}
	headerH := float32(22)
	headerBg := canvas.NewRectangle(color.NRGBA{R: 218, G: 225, B: 245, A: 255})
	headerBg.Move(fyne.NewPos(dw.X, dw.Y))
	headerBg.Resize(fyne.NewSize(dw.W, headerH))
	objs = append(objs, headerBg)
	colW := dw.W / float32(max(len(cols), 1))
	for i, col := range cols {
		col = strings.TrimSpace(col)
		x := dw.X + float32(i)*colW
		if i > 0 {
			div := canvas.NewRectangle(color.NRGBA{R: 170, G: 175, B: 210, A: 255})
			div.Move(fyne.NewPos(x, dw.Y))
			div.Resize(fyne.NewSize(1, headerH))
			objs = append(objs, div)
		}
		objs = append(objs, r.text(col, x+4, dw.Y+4, colW-8, 14, 11,
			color.NRGBA{R: 40, G: 50, B: 100, A: 255}, fyne.TextAlignLeading, true))
	}
	// 模拟行
	for row := 0; row < 2; row++ {
		y := dw.Y + headerH + float32(row)*22
		if y+22 > dw.Y+dw.H {
			break
		}
		bg := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		if row%2 == 1 {
			bg = color.NRGBA{R: 247, G: 249, B: 255, A: 255}
		}
		rb := canvas.NewRectangle(bg)
		rb.Move(fyne.NewPos(dw.X+1, y))
		rb.Resize(fyne.NewSize(dw.W-2, 22))
		objs = append(objs, rb)
	}
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderTreeView(dw *model.DesignWidget) []fyne.CanvasObject {
	rootsStr := dw.GetProperty("RootNodes")
	roots := strings.Split(rootsStr, ",")
	objs := []fyne.CanvasObject{
		r.shadow(dw),
		r.baseBg(dw, color.NRGBA{R: 255, G: 255, B: 255, A: 255}),
		r.border(dw, color.NRGBA{R: 160, G: 170, B: 200, A: 255}, 1),
	}
	rowH := float32(22)
	for i, root := range roots {
		root = strings.TrimSpace(root)
		y := dw.Y + float32(i)*(rowH+2) + 4
		if y+rowH > dw.Y+dw.H {
			break
		}
		arrow := r.text("▶", dw.X+4, y+(rowH-12)/2, 12, 12, 10,
			color.NRGBA{R: 80, G: 90, B: 120, A: 255}, fyne.TextAlignLeading, false)
		folder := r.text("📁", dw.X+16, y+(rowH-12)/2, 14, 12, 10,
			color.NRGBA{R: 200, G: 160, B: 40, A: 255}, fyne.TextAlignLeading, false)
		label := r.text(root, dw.X+30, y+(rowH-12)/2, dw.W-34, 12, 12,
			color.NRGBA{R: 20, G: 20, B: 40, A: 255}, fyne.TextAlignLeading, false)
		objs = append(objs, arrow, folder, label)
		// 子节点示例
		if i == 0 {
			childY := y + rowH
			if childY+rowH <= dw.Y+dw.H {
				childLabel := r.text("└ 子节点", dw.X+34, childY+(rowH-12)/2, dw.W-38, 12, 11,
					color.NRGBA{R: 80, G: 80, B: 100, A: 200}, fyne.TextAlignLeading, false)
				objs = append(objs, childLabel)
			}
		}
	}
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderProgressBar(dw *model.DesignWidget) []fyne.CanvasObject {
	minV := parseF32(dw.GetProperty("Min"), 0)
	maxV := parseF32(dw.GetProperty("Max"), 100)
	val := parseF32(dw.GetProperty("Value"), 50)
	infinite := dw.GetProperty("Infinite") == "true"

	ratio := float32(0)
	if maxV > minV {
		ratio = (val - minV) / (maxV - minV)
		if ratio < 0 {
			ratio = 0
		}
		if ratio > 1 {
			ratio = 1
		}
	}
	if infinite {
		ratio = 0.4
	}

	objs := []fyne.CanvasObject{
		r.baseBg(dw, color.NRGBA{R: 225, G: 230, B: 245, A: 255}),
		r.border(dw, color.NRGBA{R: 170, G: 175, B: 205, A: 255}, 1),
	}
	if ratio > 0 {
		fill := canvas.NewRectangle(color.NRGBA{R: 60, G: 140, B: 220, A: 255})
		fill.CornerRadius = 2
		fill.Move(fyne.NewPos(dw.X+1, dw.Y+1))
		fill.Resize(fyne.NewSize((dw.W-2)*ratio, dw.H-2))
		objs = append(objs, fill)
	}
	label := fmt.Sprintf("%.0f%%", ratio*100)
	if infinite {
		label = "loading…"
	}
	objs = append(objs, r.text(label, dw.X, dw.Y, dw.W, dw.H, 10, color.NRGBA{R: 30, G: 30, B: 60, A: 220}, fyne.TextAlignCenter, false))
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderStatusStrip(dw *model.DesignWidget) []fyne.CanvasObject {
	txt := dw.GetProperty("Text")
	objs := []fyne.CanvasObject{
		r.baseBg(dw, color.NRGBA{R: 220, G: 225, B: 220, A: 255}),
		r.border(dw, color.NRGBA{R: 160, G: 175, B: 160, A: 255}, 1),
		r.text("● "+txt, dw.X+4, dw.Y+(dw.H-12)/2, dw.W-8, 12, 11,
			color.NRGBA{R: 40, G: 80, B: 40, A: 255}, fyne.TextAlignLeading, false),
	}
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderPictureBox(dw *model.DesignWidget) []fyne.CanvasObject {
	objs := []fyne.CanvasObject{
		r.shadow(dw),
		r.baseBg(dw, color.NRGBA{R: 200, G: 215, B: 235, A: 255}),
		r.border(dw, color.NRGBA{R: 140, G: 160, B: 200, A: 255}, 1),
	}
	// 图片图标占位
	iconSize := float32(36)
	if iconSize > dw.W-10 {
		iconSize = dw.W - 10
	}
	if iconSize > dw.H-10 {
		iconSize = dw.H - 10
	}
	iconX := dw.X + (dw.W-iconSize)/2
	iconY := dw.Y + (dw.H-iconSize)/2 - 8
	iconBg := canvas.NewRectangle(color.NRGBA{R: 160, G: 185, B: 220, A: 200})
	iconBg.Move(fyne.NewPos(iconX, iconY))
	iconBg.Resize(fyne.NewSize(iconSize, iconSize))
	objs = append(objs,
		iconBg,
		r.text("🖼", iconX, iconY+iconSize/2-8, iconSize, 18, 20, color.NRGBA{R: 80, G: 110, B: 160, A: 255}, fyne.TextAlignCenter, false),
		r.text("PictureBox", dw.X+2, dw.Y+dw.H-14, dw.W-4, 12, 10, color.NRGBA{R: 80, G: 100, B: 140, A: 200}, fyne.TextAlignCenter, false),
	)
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderSeparator(dw *model.DesignWidget) []fyne.CanvasObject {
	line := canvas.NewRectangle(color.NRGBA{R: 180, G: 185, B: 205, A: 255})
	line.Move(fyne.NewPos(dw.X, dw.Y+dw.H/2))
	line.Resize(fyne.NewSize(dw.W, 1))
	objs := []fyne.CanvasObject{r.baseBg(dw, color.Transparent), line}
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func (r *designCanvasRenderer) renderSlider(dw *model.DesignWidget) []fyne.CanvasObject {
	minV := parseF32(dw.GetProperty("Min"), 0)
	maxV := parseF32(dw.GetProperty("Max"), 100)
	val := parseF32(dw.GetProperty("Value"), 0)
	ratio := float32(0)
	if maxV > minV {
		ratio = (val - minV) / (maxV - minV)
	}
	trackH := float32(4)
	trackY := dw.Y + (dw.H-trackH)/2
	track := canvas.NewRectangle(color.NRGBA{R: 200, G: 205, B: 225, A: 255})
	track.CornerRadius = 2
	track.Move(fyne.NewPos(dw.X+8, trackY))
	track.Resize(fyne.NewSize(dw.W-16, trackH))
	fill := canvas.NewRectangle(color.NRGBA{R: 70, G: 140, B: 220, A: 255})
	fill.CornerRadius = 2
	fill.Move(fyne.NewPos(dw.X+8, trackY))
	fill.Resize(fyne.NewSize((dw.W-16)*ratio, trackH))
	thumbX := dw.X + 8 + (dw.W-16)*ratio - 7
	thumb := canvas.NewCircle(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	thumb.StrokeColor = color.NRGBA{R: 60, G: 130, B: 210, A: 255}
	thumb.StrokeWidth = 2
	thumb.Move(fyne.NewPos(thumbX, dw.Y+(dw.H-14)/2))
	thumb.Resize(fyne.NewSize(14, 14))
	objs := []fyne.CanvasObject{r.baseBg(dw, color.Transparent), track, fill, thumb}
	objs = append(objs, r.selectionBorder(dw)...)
	return objs
}

func parseF32(s string, def float32) float32 {
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err != nil {
		return def
	}
	return float32(f)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
