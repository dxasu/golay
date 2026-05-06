package designer

import (
	"fmt"
	"strings"

	"github.com/mico/golay/model"
)

// CodeGenOptions 代码生成选项
type CodeGenOptions struct {
	PackageName string
	AppTitle    string
	WindowW     float32
	WindowH     float32
	ModulePath  string
}

// DefaultCodeGenOptions 默认选项
func DefaultCodeGenOptions() CodeGenOptions {
	return CodeGenOptions{
		PackageName: "main",
		AppTitle:    "My App",
		WindowW:     800,
		WindowH:     600,
		ModulePath:  "myapp",
	}
}

// GenerateCode 根据多个 Form 和 MessageBox 列表生成完整的 Go 程序代码
func GenerateCode(forms []*model.FormDef, dialogs []*model.MessageBoxDef, opts CodeGenOptions) string {
	if len(forms) == 0 {
		return ""
	}

	// 收集所有控件用于 import 分析
	var allWidgets []*model.DesignWidget
	for _, f := range forms {
		allWidgets = append(allWidgets, f.Widgets...)
	}

	var b strings.Builder
	imports := collectImportsAll(allWidgets, dialogs)

	b.WriteString(fmt.Sprintf("package %s\n\n", opts.PackageName))
	b.WriteString("import (\n")
	for _, imp := range imports {
		b.WriteString(fmt.Sprintf("\t%q\n", imp))
	}
	b.WriteString(")\n\n")

	// mustParseURL helper（如有 LinkLabel）
	for _, dw := range allWidgets {
		if dw.Type == model.WidgetLinkLabel {
			b.WriteString("import \"net/url\"\n\n")
			b.WriteString("func mustParseURL(raw string) *url.URL {\n")
			b.WriteString("\tu, _ := url.Parse(raw)\n\treturn u\n}\n\n")
			break
		}
	}

	// Form 级事件处理函数
	for _, form := range forms {
		for _, ev := range form.Events {
			if ev.Enabled && ev.HandlerFn != "" {
				b.WriteString(fmt.Sprintf("// %s - Form 事件 %s\n", form.Name, ev.EventName))
				b.WriteString(fmt.Sprintf("func %s() {\n\t// TODO: 实现逻辑\n}\n\n", ev.HandlerFn))
			}
		}
	}

	// 控件事件处理函数
	for _, dw := range allWidgets {
		for _, h := range buildEventHandlers([]*model.DesignWidget{dw}) {
			b.WriteString(h)
			b.WriteString("\n")
		}
	}

	// MessageBox 辅助函数
	for _, mb := range dialogs {
		b.WriteString(genMessageBoxFunc(mb))
		b.WriteString("\n")
	}

	// main()
	mainForm := forms[0]
	b.WriteString("func main() {\n")
	b.WriteString("\ta := app.New()\n")
	b.WriteString(genWindowInit("w", mainForm, true))

	// 主窗体 Load 事件
	for _, ev := range mainForm.Events {
		if ev.EventName == "Load" && ev.Enabled {
			b.WriteString(fmt.Sprintf("\t%s()\n", ev.HandlerFn))
		}
	}

	// 主窗体 FormClosing 事件
	for _, ev := range mainForm.Events {
		if ev.EventName == "FormClosing" && ev.Enabled {
			b.WriteString(fmt.Sprintf("\tw.SetCloseIntercept(func() { %s() })\n", ev.HandlerFn))
		}
	}

	b.WriteString("\tw.ShowAndRun()\n}\n\n")

	// 子窗体函数
	for i, form := range forms[1:] {
		_ = i
		b.WriteString(genSubFormFunc(form))
		b.WriteString("\n")
	}

	return b.String()
}

// genWindowInit 生成窗口初始化代码（包含控件布局）
func genWindowInit(winVar string, form *model.FormDef, isMaster bool) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("\t%s := a.NewWindow(%q)\n", winVar, form.Title))
	b.WriteString(fmt.Sprintf("\t%s.Resize(fyne.NewSize(%.0f, %.0f))\n", winVar, form.Width, form.Height))

	if form.FixedSize {
		b.WriteString(fmt.Sprintf("\t%s.SetFixedSize(true)\n", winVar))
	}
	if form.FullScreen {
		b.WriteString(fmt.Sprintf("\t%s.SetFullScreen(true)\n", winVar))
	}
	if isMaster {
		b.WriteString(fmt.Sprintf("\t%s.SetMaster()\n", winVar))
	}
	if form.IconPath != "" {
		b.WriteString(fmt.Sprintf("\t// 设置图标: %s.SetIcon(loadIcon(%q))\n", winVar, form.IconPath))
	}

	b.WriteString("\n")

	for _, dw := range form.Widgets {
		code := generateWidgetInit(dw)
		if code != "" {
			b.WriteString(code)
			b.WriteString("\n")
		}
	}

	if len(form.Widgets) > 0 {
		b.WriteString(fmt.Sprintf("\tcontent := container.NewWithoutLayout(\n"))
		for _, dw := range form.Widgets {
			b.WriteString(fmt.Sprintf("\t\t%s,\n", dw.Name))
		}
		b.WriteString("\t)\n\n")

		for _, dw := range form.Widgets {
			b.WriteString(fmt.Sprintf("\t%s.Move(fyne.NewPos(%.0f, %.0f))\n", dw.Name, dw.X, dw.Y))
			b.WriteString(fmt.Sprintf("\t%s.Resize(fyne.NewSize(%.0f, %.0f))\n", dw.Name, dw.W, dw.H))
		}
		b.WriteString(fmt.Sprintf("\n\t%s.SetContent(content)\n", winVar))
	} else {
		b.WriteString(fmt.Sprintf("\t%s.SetContent(container.NewWithoutLayout())\n", winVar))
	}
	return b.String()
}

// genSubFormFunc 生成子窗体的 Show/ShowModal 函数
func genSubFormFunc(form *model.FormDef) string {
	var b strings.Builder
	funcName := "Show" + capitalize(form.Name)

	if form.IsModal {
		b.WriteString(fmt.Sprintf("// %s 以模态方式显示子窗体 %s\n", funcName+"Modal", form.Title))
		b.WriteString(fmt.Sprintf("func %sModal(a fyne.App, parent fyne.Window) {\n", funcName))
	} else {
		b.WriteString(fmt.Sprintf("// %s 显示子窗体 %s\n", funcName, form.Title))
		b.WriteString(fmt.Sprintf("func %s(a fyne.App) {\n", funcName))
	}

	b.WriteString(genWindowInit("w", form, false))

	if form.IsModal {
		b.WriteString("\tw.CenterOnScreen()\n")
		b.WriteString("\tw.Show()\n")
		// Fyne 本身没有真正的 modal block，用 Canvas 方式模拟
		b.WriteString("\t// 模态显示：阻止父窗体操作直到此窗口关闭\n")
		b.WriteString("\tparent.Canvas().SetOnTypedKey(nil)\n")
		b.WriteString("\tdefer func() { /* 恢复父窗体 */ }()\n")
	} else {
		b.WriteString("\tw.Show()\n")
	}
	b.WriteString("}\n")
	return b.String()
}

// genMessageBoxFunc 生成 MessageBox 辅助函数
func genMessageBoxFunc(mb *model.MessageBoxDef) string {
	var b strings.Builder
	fnName := mb.ShowFuncName()
	b.WriteString(fmt.Sprintf("// %s 显示消息框\n", fnName))
	b.WriteString(fmt.Sprintf("func %s(parent fyne.Window) {\n", fnName))

	switch mb.Type {
	case model.MsgBoxInfo:
		b.WriteString(fmt.Sprintf("\tdialog.ShowInformation(%q, %q, parent)\n", mb.Title, mb.Message))
	case model.MsgBoxWarning:
		b.WriteString(fmt.Sprintf("\terr := fmt.Errorf(%q)\n", mb.Message))
		b.WriteString("\tdialog.ShowError(err, parent)\n")
	case model.MsgBoxError:
		b.WriteString(fmt.Sprintf("\terr := fmt.Errorf(%q)\n", mb.Message))
		b.WriteString("\tdialog.ShowError(err, parent)\n")
	case model.MsgBoxConfirm:
		b.WriteString(fmt.Sprintf("\tdialog.ShowConfirm(%q, %q, func(ok bool) {\n", mb.Title, mb.Message))
		b.WriteString("\t\t// TODO: 处理用户选择\n")
		b.WriteString("\t\t_ = ok\n\t}, parent)\n")
	}
	b.WriteString("}\n")
	return b.String()
}

// ─────────────────────────────────────────────────────────────────────────────
// import 收集
// ─────────────────────────────────────────────────────────────────────────────

func collectImports(widgets []*model.DesignWidget) []string {
	return collectImportsAll(widgets, nil)
}

func collectImportsAll(widgets []*model.DesignWidget, dialogs []*model.MessageBoxDef) []string {
	need := map[string]bool{
		"fyne.io/fyne/v2":           true,
		"fyne.io/fyne/v2/app":       true,
		"fyne.io/fyne/v2/container": true,
	}
	if len(dialogs) > 0 {
		need["fyne.io/fyne/v2/dialog"] = true
		need["fmt"] = true
	}
	for _, dw := range widgets {
		switch dw.Type {
		case model.WidgetButton, model.WidgetLabel, model.WidgetTextBox,
			model.WidgetMultiLineEntry, model.WidgetComboBox, model.WidgetCheckBox,
			model.WidgetRadioButton, model.WidgetSlider, model.WidgetProgressBar,
			model.WidgetSeparator, model.WidgetRichTextBox, model.WidgetLinkLabel,
			model.WidgetListBox, model.WidgetDataGridView, model.WidgetTreeView,
			model.WidgetToolStrip, model.WidgetStatusStrip:
			need["fyne.io/fyne/v2/widget"] = true
		case model.WidgetPictureBox:
			need["fyne.io/fyne/v2/canvas"] = true
		case model.WidgetGroupBox, model.WidgetTabControl:
			need["fyne.io/fyne/v2/widget"] = true
		}
		for _, ev := range dw.EnabledEvents() {
			if ev.Kind == model.EventKindMouse || ev.Kind == model.EventKindKey {
				need["fyne.io/fyne/v2/driver/desktop"] = true
			}
		}
	}
	order := []string{
		"fmt",
		"fyne.io/fyne/v2",
		"fyne.io/fyne/v2/app",
		"fyne.io/fyne/v2/canvas",
		"fyne.io/fyne/v2/container",
		"fyne.io/fyne/v2/dialog",
		"fyne.io/fyne/v2/driver/desktop",
		"fyne.io/fyne/v2/widget",
	}
	var result []string
	for _, imp := range order {
		if need[imp] {
			result = append(result, imp)
		}
	}
	return result
}

// ─────────────────────────────────────────────────────────────────────────────
// 事件处理函数生成
// ─────────────────────────────────────────────────────────────────────────────

func buildEventHandlers(widgets []*model.DesignWidget) []string {
	var result []string
	for _, dw := range widgets {
		for _, ev := range dw.EnabledEvents() {
			result = append(result, buildHandlerFunc(dw, ev))
		}
	}
	return result
}

func buildHandlerFunc(dw *model.DesignWidget, ev model.EventBinding) string {
	comment := fmt.Sprintf("// %s - %s 事件", dw.Name, ev.EventName)
	var sig string
	switch ev.Kind {
	case model.EventKindVoid:
		sig = fmt.Sprintf("func %s() {", ev.HandlerFn)
	case model.EventKindString:
		sig = fmt.Sprintf("func %s(value string) {", ev.HandlerFn)
	case model.EventKindBool:
		sig = fmt.Sprintf("func %s(checked bool) {", ev.HandlerFn)
	case model.EventKindFloat:
		sig = fmt.Sprintf("func %s(value float64) {", ev.HandlerFn)
	case model.EventKindInt:
		sig = fmt.Sprintf("func %s(index int) {", ev.HandlerFn)
	case model.EventKindKey:
		sig = fmt.Sprintf("func %s(key *fyne.KeyEvent) {", ev.HandlerFn)
	case model.EventKindMouse:
		sig = fmt.Sprintf("func %s(ev *desktop.MouseEvent) {", ev.HandlerFn)
	case model.EventKindTableCell:
		sig = fmt.Sprintf("func %s(row, col int) {", ev.HandlerFn)
	case model.EventKindTableChange:
		sig = fmt.Sprintf("func %s(row, col int, value string) {", ev.HandlerFn)
	default:
		sig = fmt.Sprintf("func %s() {", ev.HandlerFn)
	}
	return fmt.Sprintf("%s\n%s\n\t// TODO: 实现逻辑\n}\n", comment, sig)
}

// ─────────────────────────────────────────────────────────────────────────────
// 控件初始化代码生成
// ─────────────────────────────────────────────────────────────────────────────

func generateWidgetInit(dw *model.DesignWidget) string {
	switch dw.Type {
	case model.WidgetButton:
		return genButton(dw)
	case model.WidgetLabel:
		return genLabel(dw)
	case model.WidgetLinkLabel:
		return genLinkLabel(dw)
	case model.WidgetTextBox:
		return genTextBox(dw)
	case model.WidgetMultiLineEntry:
		return genMultiLineEntry(dw)
	case model.WidgetRichTextBox:
		return genRichTextBox(dw)
	case model.WidgetComboBox:
		return genComboBox(dw)
	case model.WidgetCheckBox:
		return genCheckBox(dw)
	case model.WidgetRadioButton:
		return genRadioButton(dw)
	case model.WidgetDateTimePicker:
		return genDateTimePicker(dw)
	case model.WidgetSlider:
		return genSlider(dw)
	case model.WidgetProgressBar:
		return genProgressBar(dw)
	case model.WidgetSeparator:
		return fmt.Sprintf("\t%s := widget.NewSeparator()\n", dw.Name)
	case model.WidgetPanel:
		return genPanel(dw)
	case model.WidgetGroupBox:
		return genGroupBox(dw)
	case model.WidgetTabControl:
		return genTabControl(dw)
	case model.WidgetToolStrip:
		return genToolStrip(dw)
	case model.WidgetDataGridView:
		return genDataGridView(dw)
	case model.WidgetListBox:
		return genListBox(dw)
	case model.WidgetListView:
		return genListView(dw)
	case model.WidgetTreeView:
		return genTreeView(dw)
	case model.WidgetStatusStrip:
		return genStatusStrip(dw)
	case model.WidgetPictureBox:
		return genPictureBox(dw)
	}
	return fmt.Sprintf("\t// TODO: %s (%s) 控件待实现\n", dw.Name, model.GetWidgetTypeInfo(dw.Type).Name)
}

// ── 各控件代码生成 ────────────────────────────────────────────────────────────

func genButton(dw *model.DesignWidget) string {
	text := dw.GetProperty("Text")
	style := dw.GetProperty("Style")
	enabled := dw.GetProperty("Enabled") != "false"

	// 找 Click 回调
	cb := findCallback(dw, "Click")

	var b strings.Builder
	b.WriteString(fmt.Sprintf("\t%s := widget.NewButton(%q, %s)\n", dw.Name, text, cb))
	switch style {
	case "Primary":
		b.WriteString(fmt.Sprintf("\t%s.Importance = widget.HighImportance\n", dw.Name))
	case "Warning":
		b.WriteString(fmt.Sprintf("\t%s.Importance = widget.MediumImportance\n", dw.Name))
	case "Danger":
		b.WriteString(fmt.Sprintf("\t%s.Importance = widget.DangerImportance\n", dw.Name))
	case "Low":
		b.WriteString(fmt.Sprintf("\t%s.Importance = widget.LowImportance\n", dw.Name))
	}
	if !enabled {
		b.WriteString(fmt.Sprintf("\t%s.Disable()\n", dw.Name))
	}
	return b.String()
}

func genLabel(dw *model.DesignWidget) string {
	text := dw.GetProperty("Text")
	bold := dw.GetProperty("Bold") == "true"
	align := dw.GetProperty("TextAlign")

	var b strings.Builder
	b.WriteString(fmt.Sprintf("\t%s := widget.NewLabel(%q)\n", dw.Name, text))
	if bold {
		b.WriteString(fmt.Sprintf("\t%s.TextStyle = fyne.TextStyle{Bold: true}\n", dw.Name))
	}
	switch align {
	case "Center":
		b.WriteString(fmt.Sprintf("\t%s.Alignment = fyne.TextAlignCenter\n", dw.Name))
	case "Right":
		b.WriteString(fmt.Sprintf("\t%s.Alignment = fyne.TextAlignTrailing\n", dw.Name))
	}
	return b.String()
}

func genLinkLabel(dw *model.DesignWidget) string {
	text := dw.GetProperty("Text")
	url := dw.GetProperty("URL")
	return fmt.Sprintf("\t%s := widget.NewHyperlink(%q, mustParseURL(%q))\n", dw.Name, text, url)
}

func genTextBox(dw *model.DesignWidget) string {
	password := dw.GetProperty("Password") == "true"
	placeholder := dw.GetProperty("PlaceHolder")
	text := dw.GetProperty("Text")
	readonly := dw.GetProperty("ReadOnly") == "true"
	enabled := dw.GetProperty("Enabled") != "false"

	var b strings.Builder
	if password {
		b.WriteString(fmt.Sprintf("\t%s := widget.NewPasswordEntry()\n", dw.Name))
	} else {
		b.WriteString(fmt.Sprintf("\t%s := widget.NewEntry()\n", dw.Name))
	}
	if placeholder != "" {
		b.WriteString(fmt.Sprintf("\t%s.PlaceHolder = %q\n", dw.Name, placeholder))
	}
	if text != "" {
		b.WriteString(fmt.Sprintf("\t%s.SetText(%q)\n", dw.Name, text))
	}
	for _, ev := range dw.EnabledEvents() {
		switch ev.EventName {
		case "TextChanged":
			b.WriteString(fmt.Sprintf("\t%s.OnChanged = %s\n", dw.Name, ev.HandlerFn))
		case "KeyPress":
			b.WriteString(fmt.Sprintf("\t// %s: KeyPress 事件需在 desktop.Keyable 实现中处理\n", dw.Name))
		}
	}
	if readonly {
		b.WriteString(fmt.Sprintf("\t%s.Disable()\n", dw.Name))
	} else if !enabled {
		b.WriteString(fmt.Sprintf("\t%s.Disable()\n", dw.Name))
	}
	return b.String()
}

func genMultiLineEntry(dw *model.DesignWidget) string {
	placeholder := dw.GetProperty("PlaceHolder")
	text := dw.GetProperty("Text")
	readonly := dw.GetProperty("ReadOnly") == "true"

	var b strings.Builder
	b.WriteString(fmt.Sprintf("\t%s := widget.NewMultiLineEntry()\n", dw.Name))
	if placeholder != "" {
		b.WriteString(fmt.Sprintf("\t%s.PlaceHolder = %q\n", dw.Name, placeholder))
	}
	if text != "" {
		b.WriteString(fmt.Sprintf("\t%s.SetText(%q)\n", dw.Name, text))
	}
	for _, ev := range dw.EnabledEvents() {
		if ev.EventName == "TextChanged" {
			b.WriteString(fmt.Sprintf("\t%s.OnChanged = %s\n", dw.Name, ev.HandlerFn))
		}
	}
	if readonly {
		b.WriteString(fmt.Sprintf("\t%s.Disable()\n", dw.Name))
	}
	return b.String()
}

func genRichTextBox(dw *model.DesignWidget) string {
	text := dw.GetProperty("Text")
	readonly := dw.GetProperty("ReadOnly") == "true"

	var b strings.Builder
	if readonly {
		b.WriteString(fmt.Sprintf("\t%s := widget.NewRichTextFromMarkdown(%q)\n", dw.Name, text))
	} else {
		b.WriteString(fmt.Sprintf("\t%s := widget.NewRichTextFromMarkdown(%q)\n", dw.Name, text))
		b.WriteString(fmt.Sprintf("\t%s.Wrapping = fyne.TextWrapWord\n", dw.Name))
	}
	return b.String()
}

func genComboBox(dw *model.DesignWidget) string {
	items := splitCSV(dw.GetProperty("Items"))
	selIdx := parseInt(dw.GetProperty("SelectedIndex"), 0)

	cb := "nil"
	for _, ev := range dw.EnabledEvents() {
		if ev.EventName == "SelectedIndexChanged" || ev.EventName == "TextChanged" {
			cb = ev.HandlerFn
			break
		}
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("\t%s := widget.NewSelect(%s, %s)\n", dw.Name, formatSlice(items), cb))
	if selIdx >= 0 && selIdx < len(items) {
		b.WriteString(fmt.Sprintf("\t%s.Selected = %q\n", dw.Name, items[selIdx]))
	}
	return b.String()
}

func genCheckBox(dw *model.DesignWidget) string {
	text := dw.GetProperty("Text")
	checked := dw.GetProperty("Checked") == "true"

	cb := "nil"
	for _, ev := range dw.EnabledEvents() {
		if ev.EventName == "CheckedChanged" {
			cb = ev.HandlerFn
		}
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("\t%s := widget.NewCheck(%q, %s)\n", dw.Name, text, cb))
	if checked {
		b.WriteString(fmt.Sprintf("\t%s.Checked = true\n", dw.Name))
	}
	if dw.GetProperty("Enabled") == "false" {
		b.WriteString(fmt.Sprintf("\t%s.Disable()\n", dw.Name))
	}
	return b.String()
}

func genRadioButton(dw *model.DesignWidget) string {
	opts := splitCSV(dw.GetProperty("Options"))
	selected := dw.GetProperty("Selected")
	horizontal := dw.GetProperty("Horizontal") == "true"

	cb := "nil"
	for _, ev := range dw.EnabledEvents() {
		if ev.EventName == "CheckedChanged" {
			cb = ev.HandlerFn
		}
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("\t%s := widget.NewRadioGroup(%s, %s)\n", dw.Name, formatSlice(opts), cb))
	if selected != "" {
		b.WriteString(fmt.Sprintf("\t%s.Selected = %q\n", dw.Name, selected))
	}
	if horizontal {
		b.WriteString(fmt.Sprintf("\t%s.Horizontal = true\n", dw.Name))
	}
	return b.String()
}

func genDateTimePicker(dw *model.DesignWidget) string {
	format := dw.GetProperty("Format")
	value := dw.GetProperty("Value")
	showTime := dw.GetProperty("ShowTime") == "true"

	var b strings.Builder
	b.WriteString(fmt.Sprintf("\t// DateTimePicker: %s\n", dw.Name))
	b.WriteString(fmt.Sprintf("\t%s := widget.NewEntry()\n", dw.Name))
	if value != "" {
		b.WriteString(fmt.Sprintf("\t%s.SetText(%q)\n", dw.Name, value))
	} else {
		if showTime {
			b.WriteString(fmt.Sprintf("\t%s.PlaceHolder = %q\n", dw.Name, format+" 15:04:05"))
		} else {
			b.WriteString(fmt.Sprintf("\t%s.PlaceHolder = %q\n", dw.Name, format))
		}
	}
	for _, ev := range dw.EnabledEvents() {
		if ev.EventName == "ValueChanged" {
			b.WriteString(fmt.Sprintf("\t%s.OnChanged = %s\n", dw.Name, ev.HandlerFn))
		}
	}
	return b.String()
}

func genSlider(dw *model.DesignWidget) string {
	min := parseFloat(dw.GetProperty("Min"), 0)
	max := parseFloat(dw.GetProperty("Max"), 100)
	val := parseFloat(dw.GetProperty("Value"), 0)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("\t%s := widget.NewSlider(%.0f, %.0f)\n", dw.Name, min, max))
	b.WriteString(fmt.Sprintf("\t%s.Value = %.0f\n", dw.Name, val))
	for _, ev := range dw.EnabledEvents() {
		if ev.EventName == "ValueChanged" {
			b.WriteString(fmt.Sprintf("\t%s.OnChanged = %s\n", dw.Name, ev.HandlerFn))
		}
	}
	return b.String()
}

func genProgressBar(dw *model.DesignWidget) string {
	min := parseFloat(dw.GetProperty("Min"), 0)
	max := parseFloat(dw.GetProperty("Max"), 100)
	val := parseFloat(dw.GetProperty("Value"), 50)
	infinite := dw.GetProperty("Infinite") == "true"

	var b strings.Builder
	if infinite {
		b.WriteString(fmt.Sprintf("\t%s := widget.NewProgressBarInfinite()\n", dw.Name))
	} else {
		b.WriteString(fmt.Sprintf("\t%s := widget.NewProgressBar()\n", dw.Name))
		b.WriteString(fmt.Sprintf("\t%s.Min = %.2f\n", dw.Name, min))
		b.WriteString(fmt.Sprintf("\t%s.Max = %.2f\n", dw.Name, max))
		b.WriteString(fmt.Sprintf("\t%s.Value = %.2f\n", dw.Name, val))
	}
	return b.String()
}

func genPanel(dw *model.DesignWidget) string {
	return fmt.Sprintf("\t%s := container.NewWithoutLayout() // Panel: 可向其中添加子控件\n", dw.Name)
}

func genGroupBox(dw *model.DesignWidget) string {
	title := dw.GetProperty("Text")
	return fmt.Sprintf("\t%s := widget.NewCard(%q, \"\", container.NewWithoutLayout())\n", dw.Name, title)
}

func genTabControl(dw *model.DesignWidget) string {
	tabsStr := dw.GetProperty("Tabs")
	tabs := splitCSV(tabsStr)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("\t%s := container.NewAppTabs(\n", dw.Name))
	for _, tab := range tabs {
		tab = strings.TrimSpace(tab)
		b.WriteString(fmt.Sprintf("\t\tcontainer.NewTabItem(%q, container.NewWithoutLayout()),\n", tab))
	}
	b.WriteString("\t)\n")

	selIdx := parseInt(dw.GetProperty("SelectedIndex"), 0)
	if selIdx > 0 {
		b.WriteString(fmt.Sprintf("\t%s.SelectIndex(%d)\n", dw.Name, selIdx))
	}
	for _, ev := range dw.EnabledEvents() {
		if ev.EventName == "SelectedIndexChanged" {
			b.WriteString(fmt.Sprintf("\t%s.OnChanged = func(tab *container.TabItem) {\n", dw.Name))
			b.WriteString(fmt.Sprintf("\t\t%s(%s.SelectedIndex())\n\t}\n", ev.HandlerFn, dw.Name))
		}
	}
	return b.String()
}

func genToolStrip(dw *model.DesignWidget) string {
	items := strings.Split(dw.GetProperty("Items"), ",")
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\t%s := widget.NewToolbar(\n", dw.Name))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "|" {
			b.WriteString("\t\twidget.NewToolbarSeparator(),\n")
		} else {
			b.WriteString(fmt.Sprintf("\t\twidget.NewToolbarAction(theme.DocumentIcon(), func() {\n"))
			b.WriteString(fmt.Sprintf("\t\t\t// TODO: %s 点击\n\t\t}),\n", item))
		}
	}
	b.WriteString("\t)\n")
	return b.String()
}

func genDataGridView(dw *model.DesignWidget) string {
	cols := splitCSV(dw.GetProperty("Columns"))
	colLen := len(cols)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("\t// DataGridView: %s\n", dw.Name))
	b.WriteString(fmt.Sprintf("\t%sData := [][]string{}\n", dw.Name))
	b.WriteString(fmt.Sprintf("\t%s := widget.NewTable(\n", dw.Name))
	b.WriteString(fmt.Sprintf("\t\tfunc() (int, int) { return len(%sData), %d },\n", dw.Name, colLen))
	b.WriteString("\t\tfunc() fyne.CanvasObject { return widget.NewLabel(\"\") },\n")
	b.WriteString(fmt.Sprintf("\t\tfunc(id widget.TableCellID, obj fyne.CanvasObject) {\n"))
	b.WriteString(fmt.Sprintf("\t\t\tif id.Row < len(%sData) && id.Col < len(%sData[id.Row]) {\n", dw.Name, dw.Name))
	b.WriteString(fmt.Sprintf("\t\t\t\tobj.(*widget.Label).SetText(%sData[id.Row][id.Col])\n\t\t\t}\n", dw.Name))
	b.WriteString("\t\t},\n\t)\n")

	// 表头
	b.WriteString(fmt.Sprintf("\t%s.ShowHeaderRow = true\n", dw.Name))
	b.WriteString(fmt.Sprintf("\t%s.CreateHeader = func() fyne.CanvasObject { return widget.NewLabel(\"\") }\n", dw.Name))
	b.WriteString(fmt.Sprintf("\t%s.UpdateHeader = func(id widget.TableCellID, obj fyne.CanvasObject) {\n", dw.Name))
	b.WriteString(fmt.Sprintf("\t\theaders := %s\n", formatSlice(cols)))
	b.WriteString("\t\tif id.Col < len(headers) { obj.(*widget.Label).SetText(headers[id.Col]) }\n\t}\n")

	// 事件
	for _, ev := range dw.EnabledEvents() {
		switch ev.EventName {
		case "CellClick":
			b.WriteString(fmt.Sprintf("\t%s.OnSelected = func(id widget.TableCellID) { %s(id.Row, id.Col) }\n", dw.Name, ev.HandlerFn))
		case "SelectionChanged":
			b.WriteString(fmt.Sprintf("\t%s.OnSelected = func(id widget.TableCellID) { %s(id.Row) }\n", dw.Name, ev.HandlerFn))
		}
	}
	return b.String()
}

func genListBox(dw *model.DesignWidget) string {
	items := splitCSV(dw.GetProperty("Items"))
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\t%sItems := %s\n", dw.Name, formatSlice(items)))
	cb := "nil"
	for _, ev := range dw.EnabledEvents() {
		if ev.EventName == "SelectedIndexChanged" {
			cb = ev.HandlerFn
		}
	}
	b.WriteString(fmt.Sprintf("\t%s := widget.NewList(\n", dw.Name))
	b.WriteString(fmt.Sprintf("\t\tfunc() int { return len(%sItems) },\n", dw.Name))
	b.WriteString("\t\tfunc() fyne.CanvasObject { return widget.NewLabel(\"\") },\n")
	b.WriteString(fmt.Sprintf("\t\tfunc(id widget.ListItemID, obj fyne.CanvasObject) {\n\t\t\tobj.(*widget.Label).SetText(%sItems[id])\n\t\t},\n", dw.Name))
	b.WriteString("\t)\n")
	if cb != "nil" {
		b.WriteString(fmt.Sprintf("\t%s.OnSelected = func(id widget.ListItemID) { %s(int(id)) }\n", dw.Name, cb))
	}
	return b.String()
}

func genListView(dw *model.DesignWidget) string {
	cols := splitCSV(dw.GetProperty("Columns"))
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\t// ListView: %s (类 DataGridView 多列列表)\n", dw.Name))
	b.WriteString(fmt.Sprintf("\t%sData := [][]string{}\n", dw.Name))
	b.WriteString(fmt.Sprintf("\t%s := widget.NewTable(\n", dw.Name))
	b.WriteString(fmt.Sprintf("\t\tfunc() (int, int) { return len(%sData), %d },\n", dw.Name, len(cols)))
	b.WriteString("\t\tfunc() fyne.CanvasObject { return widget.NewLabel(\"\") },\n")
	b.WriteString(fmt.Sprintf("\t\tfunc(id widget.TableCellID, obj fyne.CanvasObject) {\n\t\t\tif id.Row < len(%sData) { obj.(*widget.Label).SetText(%sData[id.Row][id.Col]) }\n\t\t},\n", dw.Name, dw.Name))
	b.WriteString("\t)\n")
	return b.String()
}

func genTreeView(dw *model.DesignWidget) string {
	roots := splitCSV(dw.GetProperty("RootNodes"))
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\t// TreeView: %s\n", dw.Name))
	b.WriteString(fmt.Sprintf("\t%sData := map[string][]string{\n", dw.Name))
	b.WriteString(fmt.Sprintf("\t\t\"\": {"))
	rootStrs := make([]string, len(roots))
	for i, r := range roots {
		rootStrs[i] = fmt.Sprintf("%q", strings.TrimSpace(r))
	}
	b.WriteString(strings.Join(rootStrs, ", "))
	b.WriteString("},\n")
	for _, r := range roots {
		r = strings.TrimSpace(r)
		b.WriteString(fmt.Sprintf("\t\t%q: {\"子节点1\", \"子节点2\"},\n", r))
	}
	b.WriteString("\t}\n")
	b.WriteString(fmt.Sprintf("\t%s := widget.NewTree(\n", dw.Name))
	b.WriteString(fmt.Sprintf("\t\tfunc(uid widget.TreeNodeID) []widget.TreeNodeID { return %sData[uid] },\n", dw.Name))
	b.WriteString(fmt.Sprintf("\t\tfunc(uid widget.TreeNodeID) bool { _, ok := %sData[uid]; return ok },\n", dw.Name))
	b.WriteString("\t\tfunc(branch bool) fyne.CanvasObject { return widget.NewLabel(\"\") },\n")
	b.WriteString("\t\tfunc(uid widget.TreeNodeID, branch bool, obj fyne.CanvasObject) { obj.(*widget.Label).SetText(uid) },\n")
	b.WriteString("\t)\n")
	for _, ev := range dw.EnabledEvents() {
		switch ev.EventName {
		case "NodeClick":
			b.WriteString(fmt.Sprintf("\t%s.OnSelected = func(uid widget.TreeNodeID) { %s(string(uid)) }\n", dw.Name, ev.HandlerFn))
		}
	}
	return b.String()
}

func genStatusStrip(dw *model.DesignWidget) string {
	text := dw.GetProperty("Text")
	return fmt.Sprintf("\t%s := widget.NewLabel(%q) // StatusStrip\n", dw.Name, text)
}

func genPictureBox(dw *model.DesignWidget) string {
	path := dw.GetProperty("ImagePath")
	sizeMode := dw.GetProperty("SizeMode")
	fillMode := "canvas.ImageFillContain"
	switch sizeMode {
	case "Stretch":
		fillMode = "canvas.ImageFillStretch"
	case "Normal":
		fillMode = "canvas.ImageFillOriginal"
	}
	if path != "" {
		return fmt.Sprintf("\t%s := canvas.NewImageFromFile(%q)\n\t%s.FillMode = %s\n", dw.Name, path, dw.Name, fillMode)
	}
	return fmt.Sprintf("\t%s := canvas.NewImageFromResource(nil) // PictureBox: 请设置图片资源\n\t%s.FillMode = %s\n", dw.Name, dw.Name, fillMode)
}

// ─────────────────────────────────────────────────────────────────────────────
// 辅助函数
// ─────────────────────────────────────────────────────────────────────────────

// findCallback 找到控件上已启用的指定事件回调，否则返回 "nil"
func findCallback(dw *model.DesignWidget, eventName string) string {
	for _, ev := range dw.EnabledEvents() {
		if ev.EventName == eventName {
			return ev.HandlerFn
		}
	}
	return "nil"
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func formatSlice(ss []string) string {
	parts := make([]string, len(ss))
	for i, s := range ss {
		parts[i] = fmt.Sprintf("%q", s)
	}
	return "[]string{" + strings.Join(parts, ", ") + "}"
}

func parseFloat(s string, def float64) float64 {
	var f float64
	if _, err := fmt.Sscanf(strings.TrimSpace(s), "%f", &f); err != nil {
		return def
	}
	return f
}

func parseInt(s string, def int) int {
	var n int
	if _, err := fmt.Sscanf(strings.TrimSpace(s), "%d", &n); err != nil {
		return def
	}
	return n
}
