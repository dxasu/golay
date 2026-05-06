package designer

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/mico/golay/model"
)

// App 设计器主应用
type App struct {
	window     fyne.Window
	dc         *DesignCanvas
	toolbox    *Toolbox
	props      *PropertiesPanel
	statusBar  *widget.Label
	statusDot  *canvas.Circle
	opts       CodeGenOptions
	project    *ProjectManager

	forms   []*model.FormDef
	dialogs []*model.MessageBoxDef
	formIdx int

	// UI 组件
	formTabs      *container.AppTabs
	dialogListBox *widget.List
	centerPanel   *fyne.Container
}

// NewApp 创建主应用
func NewApp(win fyne.Window) *App {
	a := &App{
		window:  win,
		opts:    DefaultCodeGenOptions(),
		project: NewProjectManager(),
	}
	// 创建默认主窗体
	mainForm := model.NewFormDef("mainForm", "My App", 800, 600)
	a.forms = []*model.FormDef{mainForm}
	return a
}

// Build 构建完整 UI
func (a *App) Build() {
	a.dc = NewDesignCanvas()
	a.props = NewPropertiesPanel(a.dc)
	a.toolbox = NewToolbox(a.onAddWidget)
	a.statusBar = widget.NewLabel("就绪 — 点击工具箱添加控件 | Delete 删除 | 双击控件打开编辑器")
	a.statusDot = canvas.NewCircle(colorIdle)
	a.statusDot.Resize(fyne.NewSize(8, 8))

	// 绑定当前主窗体到画布
	a.dc.SetForm(a.forms[0])

	// ── 画布事件 ──────────────────────────────────────────────────────────────
	a.dc.OnSelect = func(dw *model.DesignWidget) {
		a.props.ShowWidget(dw)
		info := model.GetWidgetTypeInfo(dw.Type)
		a.setStatus(fmt.Sprintf("选中  %s [%s]   (%.0f, %.0f)  %.0f×%.0f",
			dw.Name, info.DisplayName, dw.X, dw.Y, dw.W, dw.H), colorSelected)
	}
	a.dc.OnDeselect = func() {
		a.props.ShowFormProps(a.currentForm())
		a.setStatus("就绪 — 点击工具箱添加控件 | Delete 删除 | 双击控件打开编辑器", colorIdle)
	}
	a.dc.OnModified = func() {
		if sel := a.dc.GetSelected(); sel != nil {
			a.setStatus(fmt.Sprintf("修改  %s  (%.0f,%.0f) %.0f×%.0f",
				sel.Name, sel.X, sel.Y, sel.W, sel.H), colorEdit)
		}
	}
	a.dc.OnDoubleClick = func(dw *model.DesignWidget) {
		a.openProjectInEditor(dw, "")
	}

	// ── 属性面板事件 ──────────────────────────────────────────────────────────
	a.props.OnModified = func(_ *model.DesignWidget) { a.dc.Refresh() }
	a.props.OnFormModified = func(form *model.FormDef) {
		// 同步 opts 以便代码生成
		if form == a.forms[0] {
			a.opts.AppTitle = form.Title
			a.opts.WindowW = form.Width
			a.opts.WindowH = form.Height
		}
		// 更新 Tab 标题
		a.rebuildFormTabs()
	}
	a.props.OnOpenInEditor = func(dw *model.DesignWidget, eventName string) {
		a.openProjectInEditor(dw, eventName)
	}

	// ── Delete 键 ─────────────────────────────────────────────────────────────
	a.window.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		if ev.Name == fyne.KeyDelete {
			if a.window.Canvas().Focused() == nil && a.dc.GetSelected() != nil {
				name := a.dc.GetSelected().Name
				a.dc.DeleteSelected()
				a.setStatus(fmt.Sprintf("已删除  %s", name), colorWarn)
			}
		}
	})

	// ── 初始显示主窗体属性 ───────────────────────────────────────────────────
	a.props.ShowFormProps(a.forms[0])

	// ── 布局 ──────────────────────────────────────────────────────────────────
	toolbar := a.buildToolbar()
	leftPanel := a.buildLeftPanel()
	a.centerPanel = a.buildCenterPanel()
	rightPanel := a.buildRightPanel()

	mainSplit := container.NewHSplit(
		leftPanel,
		container.NewHSplit(a.centerPanel, rightPanel),
	)
	mainSplit.SetOffset(0.17)
	if right, ok := mainSplit.Trailing.(*container.Split); ok {
		right.SetOffset(0.68)
	}

	root := container.NewBorder(toolbar, a.buildStatusBar(), nil, nil, mainSplit)
	a.window.SetContent(root)
	a.window.SetMainMenu(a.buildMenu())
}

// ─────────────────────────────────────────────────────────────────────────────
// 工具栏
// ─────────────────────────────────────────────────────────────────────────────

func (a *App) buildToolbar() fyne.CanvasObject {
	tbBg := canvas.NewRectangle(color.NRGBA{R: 32, G: 36, B: 62, A: 255})
	tbBg.SetMinSize(fyne.NewSize(0, 46))

	mkBtn := func(label string, icon fyne.Resource, fn func()) *widget.Button {
		b := widget.NewButtonWithIcon(label, icon, fn)
		b.Importance = widget.LowImportance
		return b
	}

	vsep := func() fyne.CanvasObject {
		r := canvas.NewRectangle(color.NRGBA{R: 80, G: 85, B: 120, A: 200})
		r.SetMinSize(fyne.NewSize(1, 28))
		return r
	}

	row := container.NewHBox(
		mkBtn("新建", theme.DocumentCreateIcon(), a.confirmNewProject),
		vsep(),
		mkBtn("保存工程", theme.DocumentSaveIcon(), a.saveAndPreview),
		vsep(),
		mkBtn("新建Form", theme.ContentAddIcon(), a.addNewForm),
		mkBtn("MessageBox", theme.InfoIcon(), a.addMessageBox),
		vsep(),
		mkBtn("删除", theme.DeleteIcon(), func() { a.dc.DeleteSelected() }),
		mkBtn("清空", theme.ContentClearIcon(), a.confirmClearCanvas),
		vsep(),
		mkBtn("设置", theme.SettingsIcon(), a.showSettings),
		mkBtn("帮助", theme.HelpIcon(), a.showHelp),
	)

	// 工具栏右侧：当前 Form 信息
	formInfo := canvas.NewText("", color.NRGBA{R: 160, G: 170, B: 220, A: 200})
	formInfo.TextSize = 10
	if len(a.forms) > 0 {
		f := a.forms[0]
		formInfo.Text = fmt.Sprintf("Form: %s  %s  %.0f×%.0f", f.Name, f.Title, f.Width, f.Height)
	}

	inner := container.NewBorder(nil, nil, container.NewPadded(row), container.NewPadded(formInfo))
	return container.NewStack(tbBg, inner)
}

// ─────────────────────────────────────────────────────────────────────────────
// 左侧面板（工具箱 + MessageBox 列表）
// ─────────────────────────────────────────────────────────────────────────────

func (a *App) buildLeftPanel() fyne.CanvasObject {
	header := a.panelHeader("🔧 控件箱")
	toolboxArea := container.NewVScroll(a.toolbox)

	// MessageBox 区域
	msgHeader := a.miniSectionHeader("💬 MessageBox")
	a.dialogListBox = widget.NewList(
		func() int { return len(a.dialogs) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.InfoIcon()),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(a.dialogs) {
				return
			}
			hbox := obj.(*fyne.Container)
			lbl := hbox.Objects[1].(*widget.Label)
			lbl.SetText(a.dialogs[id].Name)
		},
	)
	a.dialogListBox.OnSelected = func(id widget.ListItemID) {
		if id < len(a.dialogs) {
			a.editMessageBox(a.dialogs[id])
		}
	}

	// 固定高度消息框区域
	msgArea := container.NewVBox(
		msgHeader,
		container.NewGridWrap(fyne.NewSize(0, 80), a.dialogListBox),
	)

	bg := canvas.NewRectangle(color.NRGBA{R: 242, G: 244, B: 252, A: 255})
	inner := container.NewBorder(header, msgArea, nil, nil, toolboxArea)
	return container.NewStack(bg, inner)
}

// ─────────────────────────────────────────────────────────────────────────────
// 中间面板（Form 标签页 + 画布）
// ─────────────────────────────────────────────────────────────────────────────

func (a *App) buildCenterPanel() *fyne.Container {
	a.rebuildFormTabsInner()
	bg := canvas.NewRectangle(color.NRGBA{R: 248, G: 250, B: 255, A: 255})
	canvasHeader := a.panelHeader("🖼 设计画布")
	inner := container.NewBorder(
		container.NewVBox(canvasHeader, a.formTabs),
		nil, nil, nil,
		container.NewScroll(a.dc),
	)
	return container.NewStack(bg, inner)
}

func (a *App) rebuildFormTabsInner() {
	tabs := make([]*container.TabItem, len(a.forms))
	for i, form := range a.forms {
		form := form
		icon := theme.HomeIcon()
		if i > 0 {
			if form.IsModal {
				icon = theme.ConfirmIcon()
			} else {
				icon = theme.DocumentIcon()
			}
		}
		label := form.Name
		if form.Title != "" {
			label = form.Name + " — " + form.Title
		}
		tabs[i] = container.NewTabItemWithIcon(label, icon, widget.NewLabel(""))
	}
	if a.formTabs == nil {
		a.formTabs = container.NewAppTabs(tabs...)
		a.formTabs.SetTabLocation(container.TabLocationTop)
		a.formTabs.OnChanged = func(tab *container.TabItem) {
			a.switchFormByTab(tab)
		}
	} else {
		// 重置 tabs
		for len(a.formTabs.Items) > 0 {
			a.formTabs.Remove(a.formTabs.Items[0])
		}
		for _, t := range tabs {
			a.formTabs.Append(t)
		}
		a.formTabs.Refresh()
	}
}

func (a *App) rebuildFormTabs() {
	if a.formTabs == nil {
		return
	}
	// 更新 tab 标题
	for i, form := range a.forms {
		if i < len(a.formTabs.Items) {
			label := form.Name
			if form.Title != "" {
				label = form.Name + " — " + form.Title
			}
			a.formTabs.Items[i].Text = label
		}
	}
	a.formTabs.Refresh()
}

func (a *App) switchFormByTab(tab *container.TabItem) {
	for i, t := range a.formTabs.Items {
		if t == tab && i < len(a.forms) {
			a.formIdx = i
			a.dc.SetForm(a.forms[i])
			a.props.ShowFormProps(a.forms[i])
			a.setStatus(fmt.Sprintf("切换到  %s [%s]", a.forms[i].Name, a.forms[i].Title), colorIdle)
			return
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 右侧属性面板
// ─────────────────────────────────────────────────────────────────────────────

func (a *App) buildRightPanel() fyne.CanvasObject {
	header := a.panelHeader("⚙ 属性 / 事件")
	bg := canvas.NewRectangle(color.NRGBA{R: 244, G: 245, B: 252, A: 255})
	inner := container.NewBorder(header, nil, nil, nil, container.NewVScroll(a.props))
	return container.NewStack(bg, inner)
}

// ─────────────────────────────────────────────────────────────────────────────
// 状态栏
// ─────────────────────────────────────────────────────────────────────────────

var (
	colorIdle     = color.NRGBA{R: 60, G: 190, B: 80, A: 255}
	colorSelected = color.NRGBA{R: 50, G: 130, B: 230, A: 255}
	colorEdit     = color.NRGBA{R: 220, G: 150, B: 30, A: 255}
	colorWarn     = color.NRGBA{R: 210, G: 60, B: 60, A: 255}
	colorOK       = color.NRGBA{R: 40, G: 180, B: 100, A: 255}
)

func (a *App) buildStatusBar() fyne.CanvasObject {
	barBg := canvas.NewRectangle(color.NRGBA{R: 24, G: 27, B: 50, A: 255})
	barBg.SetMinSize(fyne.NewSize(0, 24))
	sep := canvas.NewRectangle(color.NRGBA{R: 55, G: 60, B: 95, A: 255})
	sep.SetMinSize(fyne.NewSize(0, 1))
	a.statusBar.TextStyle = fyne.TextStyle{}
	row := container.NewHBox(container.NewPadded(a.statusDot), a.statusBar)
	return container.NewVBox(sep, container.NewStack(barBg, container.NewPadded(row)))
}

func (a *App) setStatus(msg string, dotColor color.Color) {
	a.statusBar.SetText(msg)
	a.statusDot.FillColor = dotColor
	a.statusDot.Refresh()
}

// ─────────────────────────────────────────────────────────────────────────────
// 面板辅助
// ─────────────────────────────────────────────────────────────────────────────

func (a *App) panelHeader(title string) fyne.CanvasObject {
	bg := canvas.NewRectangle(color.NRGBA{R: 225, G: 228, B: 248, A: 255})
	bg.SetMinSize(fyne.NewSize(0, 26))
	accent := canvas.NewRectangle(color.NRGBA{R: 60, G: 90, B: 200, A: 255})
	accent.SetMinSize(fyne.NewSize(3, 26))
	txt := canvas.NewText(title, color.NRGBA{R: 28, G: 32, B: 80, A: 255})
	txt.TextSize = 11
	txt.TextStyle = fyne.TextStyle{Bold: true}
	inner := container.NewBorder(nil, nil, accent, nil, container.NewPadded(txt))
	sep := canvas.NewRectangle(color.NRGBA{R: 185, G: 190, B: 225, A: 255})
	sep.SetMinSize(fyne.NewSize(0, 1))
	return container.NewVBox(container.NewStack(bg, inner), sep)
}

func (a *App) miniSectionHeader(title string) fyne.CanvasObject {
	bg := canvas.NewRectangle(color.NRGBA{R: 210, G: 215, B: 240, A: 255})
	bg.SetMinSize(fyne.NewSize(0, 20))
	txt := canvas.NewText(title, color.NRGBA{R: 50, G: 60, B: 140, A: 255})
	txt.TextSize = 10
	txt.TextStyle = fyne.TextStyle{Bold: true}
	return container.NewStack(bg, container.NewPadded(txt))
}

// ─────────────────────────────────────────────────────────────────────────────
// 菜单
// ─────────────────────────────────────────────────────────────────────────────

func (a *App) buildMenu() *fyne.MainMenu {
	fileMenu := fyne.NewMenu("文件",
		fyne.NewMenuItem("新建项目", a.confirmNewProject),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("生成代码并保存", a.saveAndPreview),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("退出", func() { a.window.Close() }),
	)
	editMenu := fyne.NewMenu("编辑",
		fyne.NewMenuItem("删除选中控件  [Delete]", func() { a.dc.DeleteSelected() }),
		fyne.NewMenuItem("清空当前画布", a.confirmClearCanvas),
	)
	formMenu := fyne.NewMenu("窗体",
		fyne.NewMenuItem("新建 Form（普通）", func() { a.addNewForm() }),
		fyne.NewMenuItem("新建 Form（模态）", func() { a.addNewFormModal() }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("添加 MessageBox", a.addMessageBox),
	)
	addMenu := fyne.NewMenu("添加控件")
	for _, info := range model.AllWidgetTypes {
		info := info
		addMenu.Items = append(addMenu.Items,
			fyne.NewMenuItem(info.DisplayName+" ("+info.Name+")", func() { a.onAddWidget(info.Type) }),
		)
	}
	helpMenu := fyne.NewMenu("帮助",
		fyne.NewMenuItem("使用说明", a.showHelp),
	)
	return fyne.NewMainMenu(fileMenu, editMenu, formMenu, addMenu, helpMenu)
}

// ─────────────────────────────────────────────────────────────────────────────
// 操作
// ─────────────────────────────────────────────────────────────────────────────

func (a *App) currentForm() *model.FormDef {
	if a.formIdx < len(a.forms) {
		return a.forms[a.formIdx]
	}
	return a.forms[0]
}

func (a *App) onAddWidget(t model.WidgetType) {
	offset := float32(len(a.dc.GetWidgets())*22) + 30
	dw := model.NewDesignWidget(t, 50+offset, 50+offset)
	a.dc.AddWidget(dw)
	a.dc.SelectWidget(dw)
	a.props.ShowWidget(dw)
	info := model.GetWidgetTypeInfo(t)
	a.setStatus(fmt.Sprintf("已添加  %s [%s]", dw.Name, info.DisplayName), colorOK)
}

func (a *App) confirmNewProject() {
	if len(a.dc.GetWidgets()) == 0 && len(a.forms) <= 1 {
		return
	}
	dialog.ShowConfirm("新建项目", "当前所有设计将被清空，确定继续吗？",
		func(ok bool) {
			if ok {
				// 重置
				a.forms = []*model.FormDef{model.NewFormDef("mainForm", "My App", 800, 600)}
				a.dialogs = nil
				a.formIdx = 0
				a.project = NewProjectManager()
				a.dc.SetForm(a.forms[0])
				a.props.ShowFormProps(a.forms[0])
				a.rebuildFormTabsInner()
				a.dialogListBox.Refresh()
				a.setStatus("已新建空白项目", colorIdle)
			}
		}, a.window)
}

func (a *App) confirmClearCanvas() {
	if len(a.dc.GetWidgets()) == 0 {
		return
	}
	dialog.ShowConfirm("清空画布",
		fmt.Sprintf("确定清空 [%s] 上的所有控件吗？", a.currentForm().Name),
		func(ok bool) {
			if ok {
				a.dc.ClearAll()
				a.setStatus("画布已清空", colorIdle)
			}
		}, a.window)
}

// ─────────────────────────────────────────────────────────────────────────────
// Form 管理
// ─────────────────────────────────────────────────────────────────────────────

func (a *App) addNewForm() {
	a.createForm(false)
}

func (a *App) addNewFormModal() {
	a.createForm(true)
}

func (a *App) createForm(modal bool) {
	n := len(a.forms) + 1
	form := model.NewFormDef(fmt.Sprintf("form%d", n), fmt.Sprintf("新窗体 %d", n), 600, 400)
	form.IsModal = modal
	a.forms = append(a.forms, form)
	a.rebuildFormTabsInner()

	// 切换到新 Form
	a.formIdx = len(a.forms) - 1
	a.dc.SetForm(form)
	a.props.ShowFormProps(form)
	if a.formTabs != nil && len(a.formTabs.Items) > 0 {
		a.formTabs.Select(a.formTabs.Items[len(a.formTabs.Items)-1])
	}
	mode := "普通"
	if modal {
		mode = "模态"
	}
	a.setStatus(fmt.Sprintf("已新建%s窗体  %s", mode, form.Name), colorOK)
}

// ─────────────────────────────────────────────────────────────────────────────
// MessageBox
// ─────────────────────────────────────────────────────────────────────────────

func (a *App) addMessageBox() {
	mb := model.NewMessageBoxDef()
	a.dialogs = append(a.dialogs, mb)
	a.dialogListBox.Refresh()
	a.editMessageBox(mb)
}

func (a *App) editMessageBox(mb *model.MessageBoxDef) {
	nameEntry := widget.NewEntry()
	nameEntry.SetText(mb.Name)

	titleEntry := widget.NewEntry()
	titleEntry.SetText(mb.Title)

	msgEntry := widget.NewMultiLineEntry()
	msgEntry.SetText(mb.Message)
	msgEntry.SetMinRowsVisible(3)

	typeSelect := widget.NewSelect([]string{"Info", "Warning", "Error", "Confirm"}, nil)
	typeSelect.Selected = string(mb.Type)

	btnSelect := widget.NewSelect([]string{"OK", "OKCancel", "YesNo"}, nil)
	btnSelect.Selected = string(mb.Buttons)

	form := widget.NewForm(
		widget.NewFormItem("变量名", nameEntry),
		widget.NewFormItem("标题", titleEntry),
		widget.NewFormItem("内容", msgEntry),
		widget.NewFormItem("类型", typeSelect),
		widget.NewFormItem("按钮", btnSelect),
	)

	// 生成函数名预览
	preview := widget.NewLabel("生成: " + mb.ShowFuncName() + "(parent fyne.Window)")
	preview.TextStyle = fyne.TextStyle{Monospace: true}
	nameEntry.OnChanged = func(v string) {
		mb.Name = v
		preview.SetText("生成: Show" + capitalize(v) + "(parent fyne.Window)")
	}

	del := widget.NewButtonWithIcon("删除此 MessageBox", theme.DeleteIcon(), func() {
		for i, d := range a.dialogs {
			if d == mb {
				a.dialogs = append(a.dialogs[:i], a.dialogs[i+1:]...)
				break
			}
		}
		a.dialogListBox.Refresh()
	})
	del.Importance = widget.DangerImportance

	content := container.NewVBox(form, preview, widget.NewSeparator(), del)
	d := dialog.NewCustomConfirm("配置 MessageBox", "保存", "取消", content,
		func(ok bool) {
			if ok {
				mb.Name = nameEntry.Text
				mb.Title = titleEntry.Text
				mb.Message = msgEntry.Text
				mb.Type = model.MsgBoxType(typeSelect.Selected)
				mb.Buttons = model.MsgBoxButtons(btnSelect.Selected)
				a.dialogListBox.Refresh()
				a.setStatus(fmt.Sprintf("已保存 MessageBox  %s", mb.Name), colorOK)
			}
		}, a.window)
	d.Resize(fyne.NewSize(420, 380))
	d.Show()
}

// ─────────────────────────────────────────────────────────────────────────────
// 保存 & 打开编辑器
// ─────────────────────────────────────────────────────────────────────────────

func (a *App) saveAndPreview() {
	hasContent := false
	for _, f := range a.forms {
		if len(f.Widgets) > 0 {
			hasContent = true
			break
		}
	}
	if !hasContent && len(a.dialogs) == 0 {
		dialog.ShowInformation("提示", "画布上没有控件，请先添加控件。", a.window)
		return
	}

	a.syncOpts()
	code := GenerateCode(a.forms, a.dialogs, a.opts)
	filePath, err := a.project.EnsureProject(code, a.opts)
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}
	a.setStatus(fmt.Sprintf("已保存 → %s", filePath), colorOK)

	codeEntry := widget.NewMultiLineEntry()
	codeEntry.SetText(code)
	codeEntry.TextStyle = fyne.TextStyle{Monospace: true}
	codeEntry.Disable()

	pathLabel := canvas.NewText("📁 "+filePath, color.NRGBA{R: 80, G: 100, B: 160, A: 255})
	pathLabel.TextSize = 10

	openFileBtn := widget.NewButtonWithIcon("在编辑器中打开", theme.DocumentIcon(), func() {
		if err := OpenFileInEditor(filePath, 0); err != nil {
			dialog.ShowError(err, a.window)
		}
	})
	openFileBtn.Importance = widget.HighImportance

	openDirBtn := widget.NewButtonWithIcon("打开工程目录", theme.FolderOpenIcon(), func() {
		if err := OpenDirInEditor(a.project.ProjectPath); err != nil {
			dialog.ShowError(err, a.window)
		}
	})

	content := container.NewBorder(
		container.NewVBox(pathLabel, widget.NewSeparator()),
		container.NewHBox(openFileBtn, openDirBtn),
		nil, nil,
		codeEntry,
	)
	d := dialog.NewCustom("生成代码预览", "关闭", content, a.window)
	d.Resize(fyne.NewSize(800, 560))
	d.Show()
}

func (a *App) openProjectInEditor(dw *model.DesignWidget, eventName string) {
	a.syncOpts()
	code := GenerateCode(a.forms, a.dialogs, a.opts)
	filePath, err := a.project.EnsureProject(code, a.opts)
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	line := 0
	if dw != nil && eventName != "" {
		for _, ev := range dw.Events {
			if ev.EventName == eventName && ev.HandlerFn != "" {
				line = FindFunctionLine(filePath, ev.HandlerFn)
				break
			}
		}
	} else if dw == nil && eventName != "" {
		// Form 事件
		for _, ev := range a.currentForm().Events {
			if ev.EventName == eventName && ev.HandlerFn != "" {
				line = FindFunctionLine(filePath, ev.HandlerFn)
				break
			}
		}
	}

	if err := OpenFileInEditor(filePath, line); err != nil {
		dialog.ShowError(err, a.window)
		return
	}
	if eventName != "" {
		name := "Form"
		if dw != nil {
			name = dw.Name
		}
		a.setStatus(fmt.Sprintf("编辑器: %s → %s (行 %d)", name, eventName, line), colorOK)
	} else {
		a.setStatus("编辑器: "+a.project.ProjectPath, colorOK)
	}
}

func (a *App) syncOpts() {
	if len(a.forms) > 0 {
		f := a.forms[0]
		a.opts.AppTitle = f.Title
		a.opts.WindowW = f.Width
		a.opts.WindowH = f.Height
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 设置 & 帮助
// ─────────────────────────────────────────────────────────────────────────────

func (a *App) showSettings() {
	f := a.forms[0]

	titleEntry := widget.NewEntry()
	titleEntry.SetText(f.Title)

	wEntry := widget.NewEntry()
	wEntry.SetText(fmt.Sprintf("%.0f", f.Width))

	hEntry := widget.NewEntry()
	hEntry.SetText(fmt.Sprintf("%.0f", f.Height))

	moduleEntry := widget.NewEntry()
	moduleEntry.SetText(a.opts.ModulePath)

	projDir := widget.NewLabel(a.project.BaseDir)
	projDir.Wrapping = fyne.TextWrapBreak

	form := widget.NewForm(
		widget.NewFormItem("主窗口标题", titleEntry),
		widget.NewFormItem("主窗口宽", wEntry),
		widget.NewFormItem("主窗口高", hEntry),
		widget.NewFormItem("Go 模块名", moduleEntry),
		widget.NewFormItem("工程目录", projDir),
	)

	dialog.ShowCustomConfirm("项目设置", "确定", "取消", form, func(ok bool) {
		if !ok {
			return
		}
		f.Title = titleEntry.Text
		var fw, fh float64
		fmt.Sscanf(wEntry.Text, "%f", &fw)
		fmt.Sscanf(hEntry.Text, "%f", &fh)
		if fw > 0 {
			f.Width = float32(fw)
		}
		if fh > 0 {
			f.Height = float32(fh)
		}
		a.opts.ModulePath = moduleEntry.Text
		a.syncOpts()
		a.rebuildFormTabs()
		a.props.ShowFormProps(f)
	}, a.window)
}

func (a *App) showHelp() {
	md := strings.Join([]string{
		"## golay 使用说明",
		"",
		"### 基本操作",
		"- 点击**左侧控件箱**中的控件 → 添加到当前画布",
		"- **单击**控件 → 选中，右侧显示属性",
		"- **拖拽**控件 → 移动位置（10px 网格吸附）",
		"- 拖拽**右下角蓝色手柄** → 调整尺寸",
		"- **Delete 键** → 删除选中控件",
		"- 点击画布空白处 → 显示当前 Form 属性",
		"",
		"### 窗体管理",
		"- 工具栏「**新建 Form**」→ 创建普通子窗体（show/showModal）",
		"- 工具栏「**MessageBox**」→ 添加弹窗模板，生成 Show 函数",
		"- 顶部标签切换不同窗体",
		"",
		"### 双击功能",
		"- **双击画布控件** → 保存工程并在编辑器中打开",
		"- **双击事件名** → 跳转到对应处理函数行",
		"",
		"### 工程保存",
		"- 自动保存到 `project/project1`、`project2`… 目录",
		"- 首次保存后自动执行 `go mod tidy`",
		"- 支持 VS Code、Zed、Sublime Text，否则使用系统默认程序",
	}, "\n")
	content := widget.NewRichTextFromMarkdown(md)
	d := dialog.NewCustom("使用说明", "关闭", container.NewVScroll(content), a.window)
	d.Resize(fyne.NewSize(540, 450))
	d.Show()
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	b := s[0]
	if b >= 'a' && b <= 'z' {
		b -= 32
	}
	return string(b) + s[1:]
}
