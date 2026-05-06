package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// golayTheme 仿 IDE 深色工具栏 + 明亮内容区主题
type golayTheme struct{}

func newGolayTheme() fyne.Theme { return &golayTheme{} }

func (t *golayTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	// 整体背景：近白，让内容区域保持整洁
	case theme.ColorNameBackground:
		return color.NRGBA{R: 248, G: 249, B: 254, A: 255}

	// 前景文字
	case theme.ColorNameForeground:
		return color.NRGBA{R: 28, G: 32, B: 60, A: 255}

	// 主色调：蓝紫色
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 70, G: 100, B: 210, A: 255}

	// 按钮默认背景
	case theme.ColorNameButton:
		return color.NRGBA{R: 230, G: 233, B: 248, A: 255}

	// 焦点/选中高亮
	case theme.ColorNameFocus:
		return color.NRGBA{R: 70, G: 120, B: 240, A: 200}

	// 悬浮背景
	case theme.ColorNameHover:
		return color.NRGBA{R: 210, G: 218, B: 250, A: 255}

	// 输入框、卡片背景
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}

	// 分割线
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 195, G: 200, B: 230, A: 255}

	// 选中/高亮
	case theme.ColorNameSelection:
		return color.NRGBA{R: 190, G: 210, B: 255, A: 200}

	// 置灰/禁用文字
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 160, G: 165, B: 195, A: 255}

	// 危险操作（红色）
	case theme.ColorNameError:
		return color.NRGBA{R: 210, G: 50, B: 60, A: 255}
	}

	return theme.DefaultTheme().Color(name, variant)
}

func (t *golayTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *golayTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *golayTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 12
	case theme.SizeNamePadding:
		return 5
	case theme.SizeNameInnerPadding:
		return 8
	case theme.SizeNameScrollBar:
		return 6
	case theme.SizeNameScrollBarSmall:
		return 3
	}
	return theme.DefaultTheme().Size(name)
}
