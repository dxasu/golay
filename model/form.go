package model

import "fmt"

// ─────────────────────────────────────────────────────────────────────────────
// FormDef  窗体定义
// ─────────────────────────────────────────────────────────────────────────────

// BorderStyle 窗体边框样式
type BorderStyle string

const (
	BorderNone        BorderStyle = "None"
	BorderSizable     BorderStyle = "Sizable"
	BorderFixed       BorderStyle = "FixedSingle"
	BorderFixedDialog BorderStyle = "FixedDialog"
	BorderFixedTool   BorderStyle = "FixedToolWindow"
)

// FormDef 代表设计器中的一个窗体（主窗体或子窗体）
type FormDef struct {
	ID          int
	Name        string      // 变量名，如 mainForm / form2
	Title       string      // 窗口标题
	Width       float32     // 宽
	Height      float32     // 高
	Resizable   bool        // 可拖拽调整大小
	FixedSize   bool        // 固定尺寸（不可调整）
	FullScreen  bool        // 全屏启动
	IconPath    string      // 图标路径
	Border      BorderStyle // 边框样式
	IsModal     bool        // 以模态方式显示（用于子窗体）
	Widgets     []*DesignWidget
	Events      []EventBinding
}

var formIDCounter int

// NewFormDef 创建一个新的窗体定义
func NewFormDef(name, title string, w, h float32) *FormDef {
	formIDCounter++
	f := &FormDef{
		ID:        formIDCounter,
		Name:      name,
		Title:     title,
		Width:     w,
		Height:    h,
		Resizable: true,
		Border:    BorderSizable,
		Widgets:   make([]*DesignWidget, 0),
	}
	f.Events = buildFormEvents(name)
	return f
}

// FormEventDefs Form 级别的事件定义
var FormEventDefs = []EventDef{
	{"Load", EventKindVoid, "窗体加载完成（初始化逻辑放这里）", false},
	{"FormClosing", EventKindVoid, "窗体正在关闭（可在此取消关闭）", false},
	{"FormClosed", EventKindVoid, "窗体已关闭", false},
	{"Resize", EventKindVoid, "窗体尺寸改变", false},
	{"Activated", EventKindVoid, "窗体获得焦点/激活", false},
	{"Deactivate", EventKindVoid, "窗体失去焦点", false},
}

func buildFormEvents(name string) []EventBinding {
	var result []EventBinding
	for _, def := range FormEventDefs {
		result = append(result, EventBinding{
			EventName: def.Name,
			Kind:      def.Kind,
			HandlerFn: name + "_" + def.Name,
			Enabled:   false,
		})
	}
	return result
}

// GetProperty 返回 FormDef 的属性值（供属性面板使用）
func (f *FormDef) GetFormProperties() []Property {
	return []Property{
		{Name: "Title", Value: f.Title, Type: "string", Description: "窗口标题"},
		{Name: "Width", Value: fmt.Sprintf("%.0f", f.Width), Type: "float", Description: "宽度（像素）"},
		{Name: "Height", Value: fmt.Sprintf("%.0f", f.Height), Type: "float", Description: "高度（像素）"},
		{Name: "Resizable", Value: boolStr(f.Resizable), Type: "bool", Description: "允许拖拽调整大小"},
		{Name: "FixedSize", Value: boolStr(f.FixedSize), Type: "bool", Description: "固定尺寸"},
		{Name: "FullScreen", Value: boolStr(f.FullScreen), Type: "bool", Description: "全屏启动"},
		{Name: "IconPath", Value: f.IconPath, Type: "file", Description: "窗口图标路径"},
		{
			Name: "Border", Value: string(f.Border), Type: "options",
			Options:     []string{string(BorderSizable), string(BorderFixed), string(BorderFixedDialog), string(BorderFixedTool), string(BorderNone)},
			Description: "边框样式",
		},
		{Name: "IsModal", Value: boolStr(f.IsModal), Type: "bool", Description: "以模态方式显示（子窗体有效）"},
	}
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// SetFormProperty 从属性面板回写属性值
func (f *FormDef) SetFormProperty(name, value string) {
	switch name {
	case "Title":
		f.Title = value
	case "Width":
		var v float32
		fmt.Sscanf(value, "%f", &v)
		if v > 0 {
			f.Width = v
		}
	case "Height":
		var v float32
		fmt.Sscanf(value, "%f", &v)
		if v > 0 {
			f.Height = v
		}
	case "Resizable":
		f.Resizable = value == "true"
	case "FixedSize":
		f.FixedSize = value == "true"
	case "FullScreen":
		f.FullScreen = value == "true"
	case "IconPath":
		f.IconPath = value
	case "Border":
		f.Border = BorderStyle(value)
	case "IsModal":
		f.IsModal = value == "true"
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// MessageBoxDef  弹窗定义
// ─────────────────────────────────────────────────────────────────────────────

// MsgBoxType 消息框类型
type MsgBoxType string

const (
	MsgBoxInfo     MsgBoxType = "Info"
	MsgBoxWarning  MsgBoxType = "Warning"
	MsgBoxError    MsgBoxType = "Error"
	MsgBoxConfirm  MsgBoxType = "Confirm"
)

// MsgBoxButtons 消息框按钮
type MsgBoxButtons string

const (
	MsgBoxBtnOK       MsgBoxButtons = "OK"
	MsgBoxBtnOKCancel MsgBoxButtons = "OKCancel"
	MsgBoxBtnYesNo    MsgBoxButtons = "YesNo"
)

// MessageBoxDef 弹窗模板定义
type MessageBoxDef struct {
	ID      int
	Name    string
	Title   string
	Message string
	Type    MsgBoxType
	Buttons MsgBoxButtons
}

var msgBoxIDCounter int

// NewMessageBoxDef 创建一个新的消息框定义
func NewMessageBoxDef() *MessageBoxDef {
	msgBoxIDCounter++
	return &MessageBoxDef{
		ID:      msgBoxIDCounter,
		Name:    fmt.Sprintf("msgBox%d", msgBoxIDCounter),
		Title:   "提示",
		Message: "这里输入消息内容",
		Type:    MsgBoxInfo,
		Buttons: MsgBoxBtnOK,
	}
}

// FuncName 返回 show/modalShow 的生成函数名
func (m *MessageBoxDef) ShowFuncName() string {
	return fmt.Sprintf("Show%s", capitalize(m.Name))
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
