package designer

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mico/golay/model"
)

const stateFileName = ".golay"
const recentMaxCount = 10

// ─────────────────────────────────────────────────────────────────────────────
// XML 项目文件格式
// ─────────────────────────────────────────────────────────────────────────────

// xmlProject .golay XML 根元素
type xmlProject struct {
	XMLName     xml.Name    `xml:"GolayProject"`
	Version     int         `xml:"version,attr"`
	SavedAt     string      `xml:"savedAt,attr"`
	ProjectPath string      `xml:"projectPath,attr,omitempty"`
	Opts        xmlOpts     `xml:"Options"`
	Forms       []xmlForm   `xml:"Forms>Form"`
	Dialogs     []xmlDialog `xml:"Dialogs>Dialog,omitempty"`
}

type xmlOpts struct {
	PackageName string  `xml:"packageName,attr"`
	AppTitle    string  `xml:"appTitle,attr"`
	WindowW     float32 `xml:"windowW,attr"`
	WindowH     float32 `xml:"windowH,attr"`
	ModulePath  string  `xml:"modulePath,attr"`
}

type xmlForm struct {
	ID         int         `xml:"id,attr"`
	Name       string      `xml:"name,attr"`
	Title      string      `xml:"title,attr"`
	Width      float32     `xml:"width,attr"`
	Height     float32     `xml:"height,attr"`
	Resizable  bool        `xml:"resizable,attr"`
	FixedSize  bool        `xml:"fixedSize,attr"`
	FullScreen bool        `xml:"fullScreen,attr"`
	IsModal    bool        `xml:"isModal,attr"`
	Border     string      `xml:"border,attr"`
	IconPath   string      `xml:"iconPath,attr,omitempty"`
	Events     []xmlEvent  `xml:"Events>Event,omitempty"`
	Widgets    []xmlWidget `xml:"Widgets>Widget,omitempty"`
}

type xmlWidget struct {
	ID         int           `xml:"id,attr"`
	Type       int           `xml:"type,attr"`
	Name       string        `xml:"name,attr"`
	X          float32       `xml:"x,attr"`
	Y          float32       `xml:"y,attr"`
	W          float32       `xml:"w,attr"`
	H          float32       `xml:"h,attr"`
	Properties []xmlProperty `xml:"Properties>Property,omitempty"`
	Events     []xmlEvent    `xml:"Events>Event,omitempty"`
}

type xmlProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type xmlEvent struct {
	Name      string `xml:"name,attr"`
	Kind      string `xml:"kind,attr"`
	HandlerFn string `xml:"handlerFn,attr"`
	Enabled   bool   `xml:"enabled,attr"`
}

type xmlDialog struct {
	ID      int    `xml:"id,attr"`
	Name    string `xml:"name,attr"`
	Title   string `xml:"title,attr"`
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Buttons string `xml:"buttons,attr"`
}

// ─────────────────────────────────────────────────────────────────────────────
// 保存 / 读取
// ─────────────────────────────────────────────────────────────────────────────

// SaveState 将设计状态保存为 XML 格式的 .golay 文件
func SaveState(dir string, forms []*model.FormDef, dialogs []*model.MessageBoxDef, opts CodeGenOptions, projectPath string) error {
	proj := buildXMLProject(forms, dialogs, opts, projectPath)
	data, err := xml.MarshalIndent(proj, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 XML 失败: %w", err)
	}
	content := append([]byte(xml.Header), data...)
	return os.WriteFile(filepath.Join(dir, stateFileName), content, 0644)
}

// LoadState 从 .golay XML 文件读取设计状态
func LoadState(statePath string) (forms []*model.FormDef, dialogs []*model.MessageBoxDef, opts CodeGenOptions, projectPath string, err error) {
	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, nil, opts, "", fmt.Errorf("读取文件失败: %w", err)
	}
	var proj xmlProject
	if err = xml.Unmarshal(data, &proj); err != nil {
		return nil, nil, opts, "", fmt.Errorf("解析 XML 失败: %w", err)
	}
	forms, dialogs, opts = parseXMLProject(proj)
	projectPath = proj.ProjectPath
	return
}

// ─────────────────────────────────────────────────────────────────────────────
// model → XML 转换
// ─────────────────────────────────────────────────────────────────────────────

func buildXMLProject(forms []*model.FormDef, dialogs []*model.MessageBoxDef, opts CodeGenOptions, projectPath string) xmlProject {
	proj := xmlProject{
		Version:     1,
		SavedAt:     time.Now().Format(time.RFC3339),
		ProjectPath: projectPath,
		Opts: xmlOpts{
			PackageName: opts.PackageName,
			AppTitle:    opts.AppTitle,
			WindowW:     opts.WindowW,
			WindowH:     opts.WindowH,
			ModulePath:  opts.ModulePath,
		},
	}
	for _, f := range forms {
		proj.Forms = append(proj.Forms, formToXML(f))
	}
	for _, d := range dialogs {
		proj.Dialogs = append(proj.Dialogs, dialogToXML(d))
	}
	return proj
}

func formToXML(f *model.FormDef) xmlForm {
	xf := xmlForm{
		ID:         f.ID,
		Name:       f.Name,
		Title:      f.Title,
		Width:      f.Width,
		Height:     f.Height,
		Resizable:  f.Resizable,
		FixedSize:  f.FixedSize,
		FullScreen: f.FullScreen,
		IsModal:    f.IsModal,
		Border:     string(f.Border),
		IconPath:   f.IconPath,
	}
	for _, ev := range f.Events {
		xf.Events = append(xf.Events, xmlEvent{
			Name:      ev.EventName,
			Kind:      string(ev.Kind),
			HandlerFn: ev.HandlerFn,
			Enabled:   ev.Enabled,
		})
	}
	for _, dw := range f.Widgets {
		xf.Widgets = append(xf.Widgets, widgetToXML(dw))
	}
	return xf
}

func widgetToXML(dw *model.DesignWidget) xmlWidget {
	xw := xmlWidget{
		ID:   dw.ID,
		Type: int(dw.Type),
		Name: dw.Name,
		X:    dw.X,
		Y:    dw.Y,
		W:    dw.W,
		H:    dw.H,
	}
	for _, p := range dw.Properties {
		xw.Properties = append(xw.Properties, xmlProperty{
			Name:  p.Name,
			Value: p.Value,
		})
	}
	for _, ev := range dw.Events {
		xw.Events = append(xw.Events, xmlEvent{
			Name:      ev.EventName,
			Kind:      string(ev.Kind),
			HandlerFn: ev.HandlerFn,
			Enabled:   ev.Enabled,
		})
	}
	return xw
}

func dialogToXML(d *model.MessageBoxDef) xmlDialog {
	return xmlDialog{
		ID:      d.ID,
		Name:    d.Name,
		Title:   d.Title,
		Message: d.Message,
		Type:    string(d.Type),
		Buttons: string(d.Buttons),
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// XML → model 转换
// ─────────────────────────────────────────────────────────────────────────────

func parseXMLProject(proj xmlProject) ([]*model.FormDef, []*model.MessageBoxDef, CodeGenOptions) {
	opts := CodeGenOptions{
		PackageName: proj.Opts.PackageName,
		AppTitle:    proj.Opts.AppTitle,
		WindowW:     proj.Opts.WindowW,
		WindowH:     proj.Opts.WindowH,
		ModulePath:  proj.Opts.ModulePath,
	}

	var forms []*model.FormDef
	for _, xf := range proj.Forms {
		forms = append(forms, xmlToForm(xf))
	}

	var dialogs []*model.MessageBoxDef
	for _, xd := range proj.Dialogs {
		dialogs = append(dialogs, xmlToDialog(xd))
	}

	return forms, dialogs, opts
}

func xmlToForm(xf xmlForm) *model.FormDef {
	form := &model.FormDef{
		ID:         xf.ID,
		Name:       xf.Name,
		Title:      xf.Title,
		Width:      xf.Width,
		Height:     xf.Height,
		Resizable:  xf.Resizable,
		FixedSize:  xf.FixedSize,
		FullScreen: xf.FullScreen,
		IsModal:    xf.IsModal,
		Border:     model.BorderStyle(xf.Border),
		IconPath:   xf.IconPath,
		Widgets:    make([]*model.DesignWidget, 0),
	}
	// 重建事件列表：从默认列表出发，覆盖已保存的 Enabled/HandlerFn
	form.Events = mergeFormEvents(model.DefaultFormEvents(xf.Name), xf.Events)

	for _, xw := range xf.Widgets {
		form.Widgets = append(form.Widgets, xmlToWidget(xw))
	}
	return form
}

func xmlToWidget(xw xmlWidget) *model.DesignWidget {
	t := model.WidgetType(xw.Type)
	dw := &model.DesignWidget{
		ID:   xw.ID,
		Type: t,
		Name: xw.Name,
		X:    xw.X,
		Y:    xw.Y,
		W:    xw.W,
		H:    xw.H,
	}
	// 重建属性：从默认属性出发，覆盖已保存的 Value
	defaults := model.DefaultProperties(t, xw.Name)
	savedVals := make(map[string]string, len(xw.Properties))
	for _, p := range xw.Properties {
		savedVals[p.Name] = p.Value
	}
	for i, p := range defaults {
		if v, ok := savedVals[p.Name]; ok {
			defaults[i].Value = v
		}
	}
	dw.Properties = defaults

	// 重建事件列表
	dw.Events = mergeWidgetEvents(model.DefaultWidgetEvents(t, xw.Name), xw.Events)
	return dw
}

func xmlToDialog(xd xmlDialog) *model.MessageBoxDef {
	return &model.MessageBoxDef{
		ID:      xd.ID,
		Name:    xd.Name,
		Title:   xd.Title,
		Message: xd.Message,
		Type:    model.MsgBoxType(xd.Type),
		Buttons: model.MsgBoxButtons(xd.Buttons),
	}
}

// mergeFormEvents 用已保存的 XML 事件状态覆盖默认事件
func mergeFormEvents(defaults []model.EventBinding, saved []xmlEvent) []model.EventBinding {
	savedMap := make(map[string]xmlEvent, len(saved))
	for _, e := range saved {
		savedMap[e.Name] = e
	}
	for i, def := range defaults {
		if s, ok := savedMap[def.EventName]; ok {
			defaults[i].Enabled = s.Enabled
			defaults[i].HandlerFn = s.HandlerFn
		}
	}
	return defaults
}

// mergeWidgetEvents 用已保存的 XML 事件状态覆盖默认事件
func mergeWidgetEvents(defaults []model.EventBinding, saved []xmlEvent) []model.EventBinding {
	savedMap := make(map[string]xmlEvent, len(saved))
	for _, e := range saved {
		savedMap[e.Name] = e
	}
	for i, def := range defaults {
		if s, ok := savedMap[def.EventName]; ok {
			defaults[i].Enabled = s.Enabled
			defaults[i].HandlerFn = s.HandlerFn
		}
	}
	return defaults
}

// ─────────────────────────────────────────────────────────────────────────────
// 最近项目管理（仍使用 JSON，作为用户配置文件）
// ─────────────────────────────────────────────────────────────────────────────

// RecentProject 最近项目条目
type RecentProject struct {
	StatePath string    `json:"statePath"` // .golay 文件绝对路径
	Name      string    `json:"name"`      // 项目目录名
	Title     string    `json:"title"`     // 主窗体标题
	SavedAt   time.Time `json:"savedAt"`
}

// recentConfigPath 返回最近项目列表文件路径（~/.golay/recent.json）
func recentConfigPath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".golay")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "recent.json")
}

// LoadRecentProjects 加载最近项目列表（自动过滤已不存在的条目）
func LoadRecentProjects() []RecentProject {
	data, err := os.ReadFile(recentConfigPath())
	if err != nil {
		return nil
	}
	var list []RecentProject
	if err := json.Unmarshal(data, &list); err != nil {
		return nil
	}
	var valid []RecentProject
	for _, p := range list {
		if _, err := os.Stat(p.StatePath); err == nil {
			valid = append(valid, p)
		}
	}
	return valid
}

// AddRecentProject 添加或更新最近项目条目
func AddRecentProject(rp RecentProject) {
	list := LoadRecentProjects()
	var filtered []RecentProject
	for _, p := range list {
		if p.StatePath != rp.StatePath {
			filtered = append(filtered, p)
		}
	}
	filtered = append([]RecentProject{rp}, filtered...)
	if len(filtered) > recentMaxCount {
		filtered = filtered[:recentMaxCount]
	}
	data, _ := json.MarshalIndent(filtered, "", "  ")
	_ = os.WriteFile(recentConfigPath(), data, 0644)
}
