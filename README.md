# wr

一个安全运营周报的终端 TUI 工具（Go + [Bubble Tea](https://github.com/charmbracelet/bubbletea)）。
左侧菜单 + 右侧内容区，日报模式和碎片模式共用同一份 daily 文件；生成周报复用现有
`~/.claude/skills/weekly-report/` skill 的整合润色逻辑。

与 `~/.claude/skills/weekly-report/`（bash 脚本 + cron + skill）是完全独立的两套系统，
互不修改、互不接管，写的是同一批 daily 文件所以数据天然共通。

## 安装

```bash
go build -o ~/.local/bin/wr .
```

`~/.local/bin` 需要在 `PATH` 里。之后直接运行：

```bash
wr
```

## 功能

启动后是全屏界面，`↑`/`↓` 选菜单，`enter` 进入，`esc` 返回菜单，`q`（在菜单里）退出。

- **日报模式**：右侧直接显示/编辑当天 `~/Desktop/weekly_reports/daily/YYYYMMDD.md` 的
  全文内容，`ctrl+s` 保存（整体覆盖写回）。
- **碎片模式**：上方是当天已记录的碎片历史，下方是单行输入框，输入文字回车后立即以
  `- HH:MM 内容` 追加到同一个 daily 文件末尾。
- **生成周报**：进入后自动调用

  ```bash
  claude -p "生成周报" --permission-mode acceptEdits
  ```

  触发现有 `weekly-report` skill 读取本周全部 daily 文件并整合润色，`wr` 本身不做任何
  AI 调用或文本润色。`--permission-mode acceptEdits` 是必须的——非交互模式下没有人能点
  权限确认弹窗，缺了这个参数 `claude -p` 会静默拒绝写文件，但进程仍然 exit code 0，
  界面会误报"已生成"。

## 项目结构

```
main.go           入口，启动 bubbletea.Program
app.go            根 Model：菜单导航、焦点切换、布局拼接
daily_pane.go     日报模式子 Model（textarea 组件）
fragment_pane.go  碎片模式子 Model（viewport 历史 + textinput 输入框）
report_pane.go    生成周报子 Model（spinner + 调用 claude 子进程）
fsutil.go         纯文件系统操作（路径规则、读写 daily 文件），不含 UI 逻辑
styles.go         lipgloss 样式集中定义
```

架构上是标准的 Bubble Tea 组合模式：根 Model 持有三个子 Model，按 `focus`
（菜单/内容）和 `active`（当前面板）路由按键消息；非按键消息（如 spinner 的
`TickMsg`、异步读写文件的结果）无条件广播给全部子面板，这样切走面板后周报生成
等后台任务依然能继续推进并正确收尾。

## 实现原理

`wr` 基于 Bubble Tea 的 Elm 架构：状态是一个不可变结构体（Model），所有变化都通过"消息"
（Msg）驱动 `Update` 产生新状态，`View` 只是把当前状态渲染成字符串的纯函数。

### 组合结构

根 Model（`rootModel`，见 `app.go`）持有三个子 Model 作为字段：`daily`、`fragment`、
`report`，分别对应三个模式。每个子 Model 各自实现自己的 `Update(msg) (T, tea.Cmd)` 和
`View() string`，根 Model 负责路由消息、拼接布局——这是 Bubble Tea 里管理多面板 UI 的
标准组合模式。

### 状态机：焦点 + 激活面板

两个枚举决定了全部交互：

```go
type focusArea int  // focusMenu / focusContent —— 键盘输入现在归谁管
type paneKind  int  // paneDaily / paneFragment / paneReport —— 右侧显示哪个面板
```

`focus == focusMenu` 时，`↑`/`↓`/`enter`/`q` 由根 Model 直接处理（菜单导航、退出、
切换 `active`）；`focus == focusContent` 时，根 Model 只拦截 `esc`（返回菜单），其余
按键转发给 `active` 对应的子 Model——所以在日报模式里打字不会误触发全局的 `q` 退出。

### 消息驱动的异步操作

`tea.Cmd` 本质是 `func() tea.Msg`：一个返回消息的函数，框架在独立 goroutine 里执行它，
执行完把返回值重新塞回 `Update`。所有耗时操作都走这条路径：

- **读文件**（`daily_pane.go` 的 `Load()`）：包装成 `tea.Cmd`，完成后产生
  `dailyLoadedMsg{content, err}`
- **写文件**（`ctrl+s` 分支）：同理产生 `dailySavedMsg`
- **生成周报**（`report_pane.go` 的 `runReportCmd`）：真正调用 `exec.Command("claude", ...)`
  的地方，这是个阻塞几十秒的操作，必须放进 `tea.Cmd` 异步执行，否则整个 TUI 会卡死

根 Model 的 `Update` 里有个关键设计：非按键消息（`default` 分支）会**无条件广播给全部
三个子面板**，而不只是当前 `active` 的那个。这就是为什么按 `esc` 切回菜单、甚至切去看
碎片模式之后，周报生成的 spinner 还能在后台继续转、完成后状态还能正确更新——异步命令
返回的消息不看你当前站在哪个面板，谁都能收到。

### 三个面板各自的实现

- **日报模式**：套用 bubbles 的 `textarea.Model`（多行可编辑文本框），`Load()` 把文件
  整体读入 `SetValue()`，`ctrl+s` 把 `Value()` 整体写回文件——是覆盖式保存，不是增量的。
- **碎片模式**：`viewport.Model`（只读可滚动展示区）+ `textinput.Model`（单行输入框）
  组合。回车时先本地更新内存里的 `lines` 切片让界面立刻有反馈，同时发一个 `tea.Cmd`
  真正调用 `appendFragment()` 追加写文件，写完的结果通过 `fragmentAppendedMsg` 回填
  成功/失败提示。
- **生成周报**：`spinner.Model` 负责转圈动画（内部靠不断自我重发 `spinner.TickMsg`
  实现"心跳"），配合 `exec.Command` 调用 `claude -p`。

### 文件系统层完全独立

`fsutil.go` 里的函数（`dailyDir`、`ensureDailyFile`、`appendFragment`、`fragmentLines`）
不依赖任何 bubbletea 类型，纯粹是路径拼接和文件读写。日报模式和碎片模式操作的是**同一个
物理文件**——日报模式整体读写，碎片模式只做行追加和"筛出碎片行"展示，两边共享这一层。

## 移植到 opencode 方不方便

**方便，改动面很小，但有一处未经验证的风险点。**

`claude` 是唯一硬编码的外部依赖，且只出现在一个地方：`report_pane.go` 的 `runReportCmd()`
函数里那一行 `exec.Command("claude", "-p", "生成周报", "--permission-mode", "acceptEdits")`。
`app.go`、`daily_pane.go`、`fragment_pane.go`、`fsutil.go` 完全不知道"claude"这个词的存在，
纯粹是文件读写和 TUI 渲染。所以理论上移植只需要改这一个函数，把命令换成 opencode 的等价
调用（参考你已有的 `weekly-auto-generate-opencode.sh` 里的 `opencode run "生成周报"`）。

**风险点**：`claude -p` 这条路径踩过一个坑——非交互模式下没有人能点权限确认弹窗，缺了
`--permission-mode acceptEdits` 会静默拒绝写文件但 exit code 仍是 0（细节见"已知限制"上一次
调试记录）。opencode 有没有等价的"非交互模式自动放行文件写入"参数、以及被拒绝时的输出长
什么样，我这台机器上没装 opencode，没法实测确认。改造时至少要重新验证一遍这条链路，不能
假设行为和 claude 一致。

另外这只是 `wr` 这个壳的移植成本；"生成周报"背后读取 daily 文件、整合润色的具体步骤，
是写在 `~/.claude/skills/weekly-report/SKILL.md` 里、由 Claude Code 的 skill 机制加载的，
opencode 能不能读到同一份指令、要不要单独给 opencode 准备一份等价的 prompt/配置，是另一个
不在 `wr-cli` 这个项目范围内、需要单独确认的问题。

## 已知限制

- 仅在 Linux 下验证过；`daily`/`reports` 目录路径写死为
  `~/Desktop/weekly_reports/`，与旧 skill 系统保持一致。
- 日报模式的保存是整篇覆盖，不是差异合并——如果你在外部编辑器和 `wr` 里同时改同一个
  文件，后保存的会覆盖先保存的。
