package model

import "fmt"

// WidgetType 控件类型枚举
type WidgetType int

const (
	// ── 容器控件 ──────────────────────────────────────
	WidgetPanel      WidgetType = iota // Panel 面板
	WidgetGroupBox                     // GroupBox 分组框
	WidgetTabControl                   // TabControl 选项卡容器

	// ── 文本显示 ──────────────────────────────────────
	WidgetLabel       // Label 静态标签
	WidgetLinkLabel   // LinkLabel 超链接标签
	WidgetRichTextBox // RichTextBox 富文本框

	// ── 数据输入 ──────────────────────────────────────
	WidgetTextBox        // TextBox 单行文本框
	WidgetMultiLineEntry // MultiLineEntry 多行文本框
	WidgetComboBox       // ComboBox 下拉组合框
	WidgetCheckBox       // CheckBox 复选框
	WidgetRadioButton    // RadioButton 单选按钮组
	WidgetDateTimePicker // DateTimePicker 日期时间选择器

	// ── 动作交互 ──────────────────────────────────────
	WidgetButton    // Button 按钮
	WidgetToolStrip // ToolStrip 工具栏

	// ── 数据展示 ──────────────────────────────────────
	WidgetDataGridView // DataGridView 数据表格
	WidgetListBox      // ListBox 列表框
	WidgetListView     // ListView 列表视图
	WidgetTreeView     // TreeView 树形控件

	// ── 进度反馈 ──────────────────────────────────────
	WidgetProgressBar // ProgressBar 进度条
	WidgetStatusStrip // StatusStrip 状态栏

	// ── 图形媒体 ──────────────────────────────────────
	WidgetPictureBox // PictureBox 图片框

	// ── 其他通用 ──────────────────────────────────────
	WidgetSeparator // Separator 分隔线
	WidgetSlider    // Slider 滑块
)

// WidgetCategory 控件分类
type WidgetCategory string

const (
	CategoryContainer WidgetCategory = "容器控件"
	CategoryText      WidgetCategory = "文本显示"
	CategoryInput     WidgetCategory = "数据输入"
	CategoryAction    WidgetCategory = "动作交互"
	CategoryData      WidgetCategory = "数据展示"
	CategoryFeedback  WidgetCategory = "进度反馈"
	CategoryMedia     WidgetCategory = "图形媒体"
	CategoryMisc      WidgetCategory = "其他"
)

// WidgetTypeInfo 控件类型元数据
type WidgetTypeInfo struct {
	Type        WidgetType
	Name        string
	DisplayName string
	Category    WidgetCategory
	DefaultW    float32
	DefaultH    float32
	Color       [3]uint8
}

// AllWidgetTypes 所有可用控件（按分类有序）
var AllWidgetTypes = []WidgetTypeInfo{
	// 容器
	{WidgetPanel, "Panel", "面板", CategoryContainer, 200, 150, [3]uint8{200, 210, 230}},
	{WidgetGroupBox, "GroupBox", "分组框", CategoryContainer, 200, 150, [3]uint8{185, 200, 220}},
	{WidgetTabControl, "TabControl", "选项卡", CategoryContainer, 300, 200, [3]uint8{170, 190, 230}},
	// 文本
	{WidgetLabel, "Label", "标签", CategoryText, 100, 28, [3]uint8{80, 80, 80}},
	{WidgetLinkLabel, "LinkLabel", "超链接", CategoryText, 120, 28, [3]uint8{30, 100, 200}},
	{WidgetRichTextBox, "RichTextBox", "富文本框", CategoryText, 200, 120, [3]uint8{250, 250, 240}},
	// 输入
	{WidgetTextBox, "TextBox", "文本框", CategoryInput, 150, 32, [3]uint8{255, 255, 255}},
	{WidgetMultiLineEntry, "MultiLineEntry", "多行文本框", CategoryInput, 200, 100, [3]uint8{245, 245, 255}},
	{WidgetComboBox, "ComboBox", "下拉框", CategoryInput, 150, 32, [3]uint8{220, 230, 255}},
	{WidgetCheckBox, "CheckBox", "复选框", CategoryInput, 120, 28, [3]uint8{100, 180, 100}},
	{WidgetRadioButton, "RadioButton", "单选组", CategoryInput, 120, 80, [3]uint8{180, 120, 60}},
	{WidgetDateTimePicker, "DateTimePicker", "日期时间", CategoryInput, 180, 32, [3]uint8{255, 220, 180}},
	// 交互
	{WidgetButton, "Button", "按钮", CategoryAction, 120, 36, [3]uint8{70, 130, 180}},
	{WidgetToolStrip, "ToolStrip", "工具栏", CategoryAction, 300, 36, [3]uint8{210, 215, 230}},
	// 数据展示
	{WidgetDataGridView, "DataGridView", "数据表格", CategoryData, 300, 180, [3]uint8{240, 245, 255}},
	{WidgetListBox, "ListBox", "列表框", CategoryData, 150, 140, [3]uint8{248, 248, 255}},
	{WidgetListView, "ListView", "列表视图", CategoryData, 200, 160, [3]uint8{245, 250, 255}},
	{WidgetTreeView, "TreeView", "树形控件", CategoryData, 180, 200, [3]uint8{245, 255, 245}},
	// 进度
	{WidgetProgressBar, "ProgressBar", "进度条", CategoryFeedback, 200, 20, [3]uint8{60, 160, 200}},
	{WidgetStatusStrip, "StatusStrip", "状态栏", CategoryFeedback, 300, 24, [3]uint8{210, 220, 210}},
	// 媒体
	{WidgetPictureBox, "PictureBox", "图片框", CategoryMedia, 150, 120, [3]uint8{200, 220, 240}},
	// 其他
	{WidgetSeparator, "Separator", "分隔线", CategoryMisc, 150, 8, [3]uint8{200, 200, 200}},
	{WidgetSlider, "Slider", "滑块", CategoryMisc, 150, 36, [3]uint8{200, 140, 60}},
}

// GetWidgetTypeInfo 按类型查找元数据
func GetWidgetTypeInfo(t WidgetType) WidgetTypeInfo {
	for _, info := range AllWidgetTypes {
		if info.Type == t {
			return info
		}
	}
	return AllWidgetTypes[0]
}

// WidgetsByCategory 按分类分组返回控件列表
func WidgetsByCategory() map[WidgetCategory][]WidgetTypeInfo {
	result := make(map[WidgetCategory][]WidgetTypeInfo)
	for _, w := range AllWidgetTypes {
		result[w.Category] = append(result[w.Category], w)
	}
	return result
}

// CategoryOrder 分类展示顺序
var CategoryOrder = []WidgetCategory{
	CategoryContainer,
	CategoryText,
	CategoryInput,
	CategoryAction,
	CategoryData,
	CategoryFeedback,
	CategoryMedia,
	CategoryMisc,
}

// ─────────────────────────────────────────────────────────────────────────────
// 属性定义
// ─────────────────────────────────────────────────────────────────────────────

// Property 控件属性
type Property struct {
	Name        string
	Value       string
	Type        string   // "string" | "bool" | "int" | "float" | "options" | "color" | "file"
	Options     []string // 仅 type=="options" 时使用
	Description string
}

// ─────────────────────────────────────────────────────────────────────────────
// 事件定义
// ─────────────────────────────────────────────────────────────────────────────

// EventKind 事件种类，决定生成的函数签名
type EventKind string

const (
	EventKindVoid        EventKind = "void"         // func()
	EventKindString      EventKind = "string"        // func(value string)
	EventKindBool        EventKind = "bool"          // func(checked bool)
	EventKindFloat       EventKind = "float"         // func(value float64)
	EventKindInt         EventKind = "int"           // func(index int)
	EventKindKey         EventKind = "key"           // func(key *fyne.KeyEvent)
	EventKindMouse       EventKind = "mouse"         // func(ev *desktop.MouseEvent)
	EventKindTableCell   EventKind = "tableCell"     // func(row, col int)
	EventKindTableChange EventKind = "tableChange"   // func(row, col int, value string)
)

// EventDef 事件定义
type EventDef struct {
	Name        string    // 事件名，如 "Click"
	Kind        EventKind // 签名类型
	Description string    // 说明
	Universal   bool      // 是否为通用事件（几乎所有控件都有）
}

// universalEvents 通用事件（所有控件共享）
var universalEvents = []EventDef{
	{"Click", EventKindVoid, "鼠标单击", true},
	{"DoubleClick", EventKindVoid, "鼠标双击", true},
	{"MouseEnter", EventKindVoid, "鼠标进入控件区域", true},
	{"MouseLeave", EventKindVoid, "鼠标离开控件区域", true},
	{"MouseMove", EventKindMouse, "鼠标在控件上移动", true},
	{"GotFocus", EventKindVoid, "控件获得焦点", true},
	{"LostFocus", EventKindVoid, "控件失去焦点", true},
	{"KeyDown", EventKindKey, "按下键盘键", true},
	{"KeyUp", EventKindKey, "释放键盘键", true},
	{"KeyPress", EventKindKey, "按下字符键", true},
}

// widgetSpecificEvents 各控件专属事件
var widgetSpecificEvents = map[WidgetType][]EventDef{
	WidgetButton: {
		{"Click", EventKindVoid, "按钮被点击（最主要事件）", false},
	},
	WidgetTextBox: {
		{"TextChanged", EventKindString, "文本内容改变时", false},
		{"KeyPress", EventKindKey, "按下字符键（可用于限制输入）", false},
	},
	WidgetMultiLineEntry: {
		{"TextChanged", EventKindString, "文本内容改变时", false},
	},
	WidgetComboBox: {
		{"SelectedIndexChanged", EventKindInt, "下拉选项索引改变", false},
		{"TextChanged", EventKindString, "文本内容改变时", false},
	},
	WidgetCheckBox: {
		{"CheckedChanged", EventKindBool, "选中状态改变", false},
	},
	WidgetRadioButton: {
		{"CheckedChanged", EventKindString, "选中项改变（返回选中值）", false},
	},
	WidgetDateTimePicker: {
		{"ValueChanged", EventKindString, "日期时间值改变", false},
	},
	WidgetSlider: {
		{"ValueChanged", EventKindFloat, "滑块值改变", false},
	},
	WidgetProgressBar: {
		{"ValueChanged", EventKindFloat, "进度值改变", false},
	},
	WidgetDataGridView: {
		{"CellClick", EventKindTableCell, "点击单元格", false},
		{"CellValueChanged", EventKindTableChange, "单元格内容改变", false},
		{"SelectionChanged", EventKindInt, "选中行改变", false},
	},
	WidgetListBox: {
		{"SelectedIndexChanged", EventKindInt, "列表选中项改变", false},
		{"DoubleClick", EventKindVoid, "双击列表项", false},
	},
	WidgetListView: {
		{"ItemActivate", EventKindInt, "激活（双击/回车）列表项", false},
		{"SelectedIndexChanged", EventKindInt, "选中项改变", false},
	},
	WidgetTreeView: {
		{"NodeClick", EventKindString, "点击树节点（返回节点 UID）", false},
		{"NodeExpand", EventKindString, "展开树节点", false},
		{"NodeCollapse", EventKindString, "折叠树节点", false},
	},
	WidgetRichTextBox: {
		{"TextChanged", EventKindString, "文本内容改变", false},
		{"SelectionChanged", EventKindVoid, "选中文本改变", false},
	},
	WidgetLinkLabel: {
		{"Click", EventKindVoid, "超链接被点击", false},
	},
	WidgetPanel: {
		{"Click", EventKindVoid, "面板被点击", false},
	},
	WidgetGroupBox: {
		{"Click", EventKindVoid, "分组框被点击", false},
	},
	WidgetPictureBox: {
		{"Click", EventKindVoid, "图片框被点击", false},
		{"DoubleClick", EventKindVoid, "图片框被双击", false},
	},
	WidgetTabControl: {
		{"SelectedIndexChanged", EventKindInt, "选项卡切换", false},
	},
}

// EventBinding 事件绑定实例（存在设计控件上）
type EventBinding struct {
	EventName string
	Kind      EventKind
	HandlerFn string
	Enabled   bool
}

// ─────────────────────────────────────────────────────────────────────────────
// DesignWidget
// ─────────────────────────────────────────────────────────────────────────────

// DesignWidget 设计器中的控件实例
type DesignWidget struct {
	ID         int
	Type       WidgetType
	Name       string
	X, Y       float32
	W, H       float32
	Properties []Property
	Events     []EventBinding
}

var globalIDCounter int

// NewDesignWidget 创建一个新的设计控件
func NewDesignWidget(t WidgetType, x, y float32) *DesignWidget {
	globalIDCounter++
	info := GetWidgetTypeInfo(t)
	dw := &DesignWidget{
		ID:   globalIDCounter,
		Type: t,
		Name: fmt.Sprintf("%s%d", lowerFirst(info.Name), globalIDCounter),
		X:    x,
		Y:    y,
		W:    info.DefaultW,
		H:    info.DefaultH,
	}
	dw.Properties = defaultProperties(t, dw.Name)
	dw.Events = buildEvents(t, dw.Name)
	return dw
}

func lowerFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	b := s[0]
	if b >= 'A' && b <= 'Z' {
		b += 32
	}
	return string(b) + s[1:]
}

// buildEvents 构建控件的完整事件列表（专属事件 + 通用事件，去重）
func buildEvents(t WidgetType, name string) []EventBinding {
	seen := map[string]bool{}
	var result []EventBinding

	// 专属事件优先
	if specific, ok := widgetSpecificEvents[t]; ok {
		for _, def := range specific {
			seen[def.Name] = true
			result = append(result, EventBinding{
				EventName: def.Name,
				Kind:      def.Kind,
				HandlerFn: name + "_" + def.Name,
				Enabled:   false,
			})
		}
	}

	// 通用事件补充（跳过已有的）
	for _, def := range universalEvents {
		if seen[def.Name] {
			continue
		}
		seen[def.Name] = true
		result = append(result, EventBinding{
			EventName: def.Name,
			Kind:      def.Kind,
			HandlerFn: name + "_" + def.Name,
			Enabled:   false,
		})
	}
	return result
}

// defaultProperties 按控件类型返回默认属性列表
func defaultProperties(t WidgetType, name string) []Property {
	base := []Property{
		{Name: "Name", Value: name, Type: "string", Description: "控件变量名"},
		{Name: "Enabled", Value: "true", Type: "bool", Description: "是否启用"},
		{Name: "Visible", Value: "true", Type: "bool", Description: "是否可见"},
	}

	switch t {
	case WidgetPanel:
		return append(base,
			Property{Name: "BackColor", Value: "#F0F0F0", Type: "color", Description: "背景色"},
			Property{Name: "BorderStyle", Value: "None", Type: "options",
				Options: []string{"None", "FixedSingle", "Fixed3D"}, Description: "边框样式"},
		)

	case WidgetGroupBox:
		return append(base,
			Property{Name: "Text", Value: "GroupBox", Type: "string", Description: "分组标题"},
		)

	case WidgetTabControl:
		return append(base,
			Property{Name: "Tabs", Value: "标签页1,标签页2,标签页3", Type: "string", Description: "选项卡标题（逗号分隔）"},
			Property{Name: "SelectedIndex", Value: "0", Type: "int", Description: "默认选中索引"},
		)

	case WidgetLabel:
		return append(base,
			Property{Name: "Text", Value: "Label", Type: "string", Description: "显示文字"},
			Property{Name: "TextAlign", Value: "Left", Type: "options",
				Options: []string{"Left", "Center", "Right"}, Description: "对齐方式"},
			Property{Name: "Bold", Value: "false", Type: "bool", Description: "粗体"},
		)

	case WidgetLinkLabel:
		return append(base,
			Property{Name: "Text", Value: "点击这里", Type: "string", Description: "显示文字"},
			Property{Name: "URL", Value: "https://example.com", Type: "string", Description: "跳转链接"},
		)

	case WidgetRichTextBox:
		return append(base,
			Property{Name: "Text", Value: "", Type: "string", Description: "初始文字"},
			Property{Name: "ReadOnly", Value: "false", Type: "bool", Description: "只读"},
			Property{Name: "WordWrap", Value: "true", Type: "bool", Description: "自动换行"},
		)

	case WidgetTextBox:
		return append(base,
			Property{Name: "Text", Value: "", Type: "string", Description: "文本内容"},
			Property{Name: "PlaceHolder", Value: "", Type: "string", Description: "占位提示"},
			Property{Name: "MaxLength", Value: "0", Type: "int", Description: "最大长度（0=不限）"},
			Property{Name: "ReadOnly", Value: "false", Type: "bool", Description: "只读"},
			Property{Name: "Password", Value: "false", Type: "bool", Description: "密码模式"},
		)

	case WidgetMultiLineEntry:
		return append(base,
			Property{Name: "Text", Value: "", Type: "string", Description: "文本内容"},
			Property{Name: "PlaceHolder", Value: "", Type: "string", Description: "占位提示"},
			Property{Name: "ReadOnly", Value: "false", Type: "bool", Description: "只读"},
			Property{Name: "WordWrap", Value: "true", Type: "bool", Description: "自动换行"},
		)

	case WidgetComboBox:
		return append(base,
			Property{Name: "Items", Value: "选项1,选项2,选项3", Type: "string", Description: "选项列表（逗号分隔）"},
			Property{Name: "SelectedIndex", Value: "0", Type: "int", Description: "默认选中索引"},
			Property{Name: "DropDownStyle", Value: "DropDownList", Type: "options",
				Options: []string{"DropDown", "DropDownList", "Simple"}, Description: "下拉样式"},
		)

	case WidgetCheckBox:
		return append(base,
			Property{Name: "Text", Value: "CheckBox", Type: "string", Description: "复选框文字"},
			Property{Name: "Checked", Value: "false", Type: "bool", Description: "默认选中"},
			Property{Name: "ThreeState", Value: "false", Type: "bool", Description: "三态模式"},
		)

	case WidgetRadioButton:
		return append(base,
			Property{Name: "Options", Value: "选项1,选项2,选项3", Type: "string", Description: "选项列表（逗号分隔）"},
			Property{Name: "Selected", Value: "", Type: "string", Description: "默认选中项"},
			Property{Name: "Horizontal", Value: "false", Type: "bool", Description: "水平排列"},
		)

	case WidgetDateTimePicker:
		return append(base,
			Property{Name: "Format", Value: "2006-01-02", Type: "string", Description: "日期格式（Go 风格）"},
			Property{Name: "Value", Value: "", Type: "string", Description: "默认值（空=当前时间）"},
			Property{Name: "ShowTime", Value: "false", Type: "bool", Description: "显示时间"},
		)

	case WidgetButton:
		return append(base,
			Property{Name: "Text", Value: "Button", Type: "string", Description: "按钮文字"},
			Property{Name: "Style", Value: "Default", Type: "options",
				Options: []string{"Default", "Primary", "Warning", "Danger", "Low"}, Description: "按钮样式"},
		)

	case WidgetToolStrip:
		return append(base,
			Property{Name: "Items", Value: "新建,打开,保存,|,剪切,复制,粘贴", Type: "string", Description: "工具栏项（| 为分隔符）"},
		)

	case WidgetDataGridView:
		return append(base,
			Property{Name: "Columns", Value: "列1,列2,列3", Type: "string", Description: "列标题（逗号分隔）"},
			Property{Name: "AllowEdit", Value: "true", Type: "bool", Description: "允许编辑"},
			Property{Name: "AllowSort", Value: "true", Type: "bool", Description: "允许排序"},
			Property{Name: "MultiSelect", Value: "false", Type: "bool", Description: "多选模式"},
			Property{Name: "RowHeight", Value: "24", Type: "int", Description: "行高"},
		)

	case WidgetListBox:
		return append(base,
			Property{Name: "Items", Value: "项目1,项目2,项目3", Type: "string", Description: "列表项（逗号分隔）"},
			Property{Name: "MultiSelect", Value: "false", Type: "bool", Description: "多选模式"},
			Property{Name: "SelectedIndex", Value: "-1", Type: "int", Description: "默认选中（-1=无）"},
		)

	case WidgetListView:
		return append(base,
			Property{Name: "Columns", Value: "列1,列2,列3", Type: "string", Description: "列标题（逗号分隔）"},
			Property{Name: "View", Value: "Details", Type: "options",
				Options: []string{"Details", "List", "SmallIcon", "LargeIcon"}, Description: "视图模式"},
			Property{Name: "MultiSelect", Value: "true", Type: "bool", Description: "多选"},
			Property{Name: "FullRowSelect", Value: "true", Type: "bool", Description: "整行选中"},
		)

	case WidgetTreeView:
		return append(base,
			Property{Name: "RootNodes", Value: "根节点1,根节点2", Type: "string", Description: "根节点（逗号分隔）"},
			Property{Name: "ShowRoot", Value: "true", Type: "bool", Description: "显示根节点"},
		)

	case WidgetProgressBar:
		return append(base,
			Property{Name: "Min", Value: "0", Type: "float", Description: "最小值"},
			Property{Name: "Max", Value: "100", Type: "float", Description: "最大值"},
			Property{Name: "Value", Value: "50", Type: "float", Description: "当前值"},
			Property{Name: "Infinite", Value: "false", Type: "bool", Description: "无限循环动画"},
		)

	case WidgetStatusStrip:
		return append(base,
			Property{Name: "Text", Value: "就绪", Type: "string", Description: "状态文字"},
		)

	case WidgetPictureBox:
		return append(base,
			Property{Name: "ImagePath", Value: "", Type: "file", Description: "图片路径"},
			Property{Name: "SizeMode", Value: "Zoom", Type: "options",
				Options: []string{"Normal", "Zoom", "Stretch", "CenterImage"}, Description: "缩放模式"},
		)

	case WidgetSeparator:
		return base

	case WidgetSlider:
		return append(base,
			Property{Name: "Min", Value: "0", Type: "float", Description: "最小值"},
			Property{Name: "Max", Value: "100", Type: "float", Description: "最大值"},
			Property{Name: "Value", Value: "0", Type: "float", Description: "当前值"},
			Property{Name: "Step", Value: "1", Type: "float", Description: "步长"},
			Property{Name: "Orientation", Value: "Horizontal", Type: "options",
				Options: []string{"Horizontal", "Vertical"}, Description: "方向"},
		)
	}
	return base
}

// ─────────────────────────────────────────────────────────────────────────────
// DesignWidget 辅助方法
// ─────────────────────────────────────────────────────────────────────────────

// GetProperty 获取指定属性值
func (dw *DesignWidget) GetProperty(name string) string {
	for _, p := range dw.Properties {
		if p.Name == name {
			return p.Value
		}
	}
	return ""
}

// SetProperty 设置指定属性值
func (dw *DesignWidget) SetProperty(name, value string) {
	for i, p := range dw.Properties {
		if p.Name == name {
			dw.Properties[i].Value = value
			if name == "Name" {
				dw.Name = value
			}
			return
		}
	}
}

// SetEventEnabled 启用/禁用某个事件
func (dw *DesignWidget) SetEventEnabled(eventName string, enabled bool) {
	for i, e := range dw.Events {
		if e.EventName == eventName {
			dw.Events[i].Enabled = enabled
			return
		}
	}
}

// SetEventHandler 设置事件处理函数名
func (dw *DesignWidget) SetEventHandler(eventName, handlerFn string) {
	for i, e := range dw.Events {
		if e.EventName == eventName {
			dw.Events[i].HandlerFn = handlerFn
			return
		}
	}
}

// EnabledEvents 返回所有已启用事件
func (dw *DesignWidget) EnabledEvents() []EventBinding {
	var result []EventBinding
	for _, e := range dw.Events {
		if e.Enabled {
			result = append(result, e)
		}
	}
	return result
}
