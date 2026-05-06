
## 一、常用控件清单（按功能分类）

WinForms 控件库庞大，但日常开发主要围绕以下核心组件：

| 控件类别 | 核心控件 | 主要用途 |
| :--- | :--- | :--- |
| **容器控件** | `Panel`, `GroupBox`, `TabControl` | 界面布局与内容分组 |
| **文本显示** | `Label`, `LinkLabel`, `RichTextBox` | 静态文本、超链接及富文本编辑 |
| **数据输入** | `TextBox`, `ComboBox`, `CheckBox`, `RadioButton`, `DateTimePicker` | 用户输入与选择 |
| **动作交互** | `Button`, `MenuStrip`, `ToolStrip`, `ContextMenuStrip` | 触发命令、菜单栏、工具栏 |
| **数据展示** | `DataGridView`, `ListBox`, `ListView`, `TreeView` | 表格、列表、树形结构数据展示 |
| **进度反馈** | `ProgressBar`, `StatusStrip` | 显示操作进度和状态信息 |
| **图形媒体** | `PictureBox` | 显示图片 |

## 二、通用核心事件（几乎适用于所有控件）

WinForms 采用**事件驱动模型**，以下事件是绝大多数控件都支持的“通用语”：

| 事件名称 | 触发时机 | 典型应用场景 |
| :--- | :--- | :--- |
| **Click** | 鼠标单击控件时 | `Button` 点击确认、`Label` 点击跳转 |
| **MouseClick** | 鼠标单击（提供更详细的鼠标按键信息） | 区分左键、右键点击 |
| **DoubleClick** | 鼠标双击控件时 | 快速打开项目、确认操作 |
| **MouseEnter** | 鼠标指针进入控件区域时 | 悬浮高亮效果（`Button` 变色） |
| **MouseLeave** | 鼠标指针离开控件区域时 | 取消悬浮状态 |
| **MouseMove** | 鼠标在控件上移动时 | 实时跟踪鼠标位置（绘图工具） |
| **KeyDown** | 控件获得焦点时按下键盘键 | 快捷键处理、输入验证 |
| **KeyUp** | 控件获得焦点时释放键盘键 | 配合 `KeyDown` 完成按键动作 |
| **KeyPress** | 控件获得焦点时按下字符键 | 主要用于字符输入处理 |
| **GotFocus** | 控件获得输入焦点时 | 输入框被选中时改变边框颜色 |
| **LostFocus** | 控件失去输入焦点时 | 输入完成后立即验证数据（失焦验证） |
| **TextChanged** | 控件的文本内容发生改变时 | 实时搜索（输入框内容变化即触发） |

## 三、高频控件的专属事件

除了通用事件，特定控件有其独特的“高光”事件：

| 控件 | 关键事件 | 说明 |
| :--- | :--- | :--- |
| **Form（窗体）** | `Load`, `FormClosing`, `FormClosed` | 窗体生命周期管理（初始化、关闭前确认） |
| **Button** | `Click` (最主要) | 99% 的交互都绑定在此事件上 |
| **TextBox** | `TextChanged`, `KeyPress` | 实时搜索、限制输入字符（如仅数字） |
| **ComboBox** | `SelectedIndexChanged` | 下拉选项改变时触发，用于联动其他控件 |
| **CheckBox / RadioButton** | `CheckedChanged` | 选中状态改变时触发 |
| **DataGridView** | `CellClick`, `CellValueChanged` | 点击单元格、编辑单元格内容 |
| **Timer** | `Tick` | 定时器滴答事件，用于周期性任务 |

## 四、Go GUI 设计器的实现启示

如果你正在用 Go 开发类似 WinForms 的设计器，这套机制是极佳的蓝本：

1.  **事件绑定机制**：Go 设计器需要实现类似的**委托（Delegate）模型**。在 WinForms 中，事件处理器（如 `button1_Click`）通过 `+=` 操作符绑定。在 Go 中，你可以用**函数回调（Callback）**或**Channel** 来模拟这一机制。
2.  **属性网格（Property Grid）**：WinForms 设计器的核心是属性窗口。你需要为每个控件维护一个属性集合（如 `Text`, `Size`, `Enabled`），并实现 `OnPropertyChanged` 事件来实时更新界面。
3.  **控件树（Control Tree）**：WinForms 的 `Controls` 集合管理子控件。你的 Go 设计器也需要一个类似的容器结构来维护父子关系，用于渲染和布局计算。
