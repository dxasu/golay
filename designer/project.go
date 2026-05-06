package designer

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)


const projectBaseName = "project"

// ProjectManager 管理生成的工程文件
type ProjectManager struct {
	BaseDir     string // e.g. <cwd>/project
	ProjectPath string // e.g. <cwd>/project/project1
	MainFile    string // e.g. <cwd>/project/project1/main.go
}

// NewProjectManager 以当前工作目录下的 project/ 为根
func NewProjectManager() *ProjectManager {
	cwd, _ := os.Getwd()
	return &ProjectManager{
		BaseDir: filepath.Join(cwd, projectBaseName),
	}
}

// nextDir 查找下一个可用的 project1、project2… 目录
func (pm *ProjectManager) nextDir() string {
	for i := 1; ; i++ {
		dir := filepath.Join(pm.BaseDir, fmt.Sprintf("%s%d", projectBaseName, i))
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return dir
		}
	}
}

// SaveProject 创建工程目录并写入 main.go + go.mod，首次保存后自动 go mod tidy
func (pm *ProjectManager) SaveProject(code string, opts CodeGenOptions) (string, error) {
	dir := pm.nextDir()
	pm.ProjectPath = dir

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("无法创建目录 %s: %w", dir, err)
	}

	mainFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(mainFile, []byte(code), 0644); err != nil {
		return "", fmt.Errorf("无法写入 main.go: %w", err)
	}
	pm.MainFile = mainFile

	modName := opts.ModulePath
	if modName == "" {
		modName = filepath.Base(dir)
	}
	gomod := fmt.Sprintf("module %s\n\ngo 1.21\n\nrequire fyne.io/fyne/v2 v2.7.3\n", modName)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0644); err != nil {
		return "", fmt.Errorf("无法写入 go.mod: %w", err)
	}

	// 首次保存后自动执行 go mod tidy（后台运行，不阻塞 UI）
	go pm.runGoModTidy(dir)

	return mainFile, nil
}

// runGoModTidy 在指定目录执行 go mod tidy
func (pm *ProjectManager) runGoModTidy(dir string) {
	goExe, err := exec.LookPath("go")
	if err != nil {
		// 尝试常见路径
		for _, p := range []string{"/usr/local/go/bin/go", "/usr/bin/go"} {
			if _, e := os.Stat(p); e == nil {
				goExe = p
				break
			}
		}
	}
	if goExe == "" {
		return
	}
	cmd := exec.Command(goExe, "mod", "tidy")
	cmd.Dir = dir
	_ = cmd.Run()
}

// UpdateProject 重新写入已有工程的 main.go
func (pm *ProjectManager) UpdateProject(code string) error {
	if pm.MainFile == "" {
		return fmt.Errorf("项目尚未保存")
	}
	return os.WriteFile(pm.MainFile, []byte(code), 0644)
}

// HasProject 是否已保存过工程
func (pm *ProjectManager) HasProject() bool {
	return pm.MainFile != ""
}

// EnsureProject 若尚未保存则先保存，返回 main.go 路径
func (pm *ProjectManager) EnsureProject(code string, opts CodeGenOptions) (string, error) {
	if pm.HasProject() {
		if err := pm.UpdateProject(code); err != nil {
			return "", err
		}
		return pm.MainFile, nil
	}
	return pm.SaveProject(code, opts)
}

// ─────────────────────────────────────────────────────────────────────────────
// 编辑器相关
// ─────────────────────────────────────────────────────────────────────────────

// FindFunctionLine 返回文件中 "func funcName(" 所在的 1-based 行号，未找到返回 1
func FindFunctionLine(filePath, funcName string) int {
	f, err := os.Open(filePath)
	if err != nil {
		return 1
	}
	defer f.Close()

	needle := "func " + funcName + "("
	line := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line++
		if strings.Contains(scanner.Text(), needle) {
			return line
		}
	}
	return 1
}

// OpenFileInEditor 用最佳可用编辑器打开文件，line>0 时尝试导航到该行
func OpenFileInEditor(filePath string, line int) error {
	// VS Code（支持 --goto 直接跳行）
	if codePath, err := exec.LookPath("code"); err == nil {
		var cmd *exec.Cmd
		if line > 0 {
			cmd = exec.Command(codePath, "--goto", fmt.Sprintf("%s:%d", filePath, line))
		} else {
			cmd = exec.Command(codePath, filePath)
		}
		return cmd.Start()
	}

	// Zed
	if zedPath, err := exec.LookPath("zed"); err == nil {
		return exec.Command(zedPath, filePath).Start()
	}

	// Sublime Text
	if sublPath, err := exec.LookPath("subl"); err == nil {
		if line > 0 {
			return exec.Command(sublPath, fmt.Sprintf("%s:%d", filePath, line)).Start()
		}
		return exec.Command(sublPath, filePath).Start()
	}

	// 系统默认程序
	return openWithSystem(filePath)
}

// OpenDirInEditor 用最佳可用编辑器打开目录
func OpenDirInEditor(dirPath string) error {
	if codePath, err := exec.LookPath("code"); err == nil {
		return exec.Command(codePath, dirPath).Start()
	}
	if zedPath, err := exec.LookPath("zed"); err == nil {
		return exec.Command(zedPath, dirPath).Start()
	}
	return openWithSystem(dirPath)
}

func openWithSystem(path string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", path).Start()
	case "windows":
		return exec.Command("explorer", path).Start()
	default:
		return exec.Command("xdg-open", path).Start()
	}
}
