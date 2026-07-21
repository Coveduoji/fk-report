# 移植到 opencode 方案

## 目标

让 `fk-report` 的"生成周报"能选择用 `opencode` 而不是 `claude` 作为后端，同时不影响
现在已经能用的 claude 路径——参考你现在 `~/.claude/skills/weekly-report/` 下
`weekly-auto-generate.sh`（claude 版）和 `weekly-auto-generate-opencode.sh`（opencode 版）
并存、opencode 版本先放着等你装好再启用的做法，这次也做成**两个后端都在、默认 claude、
可以切换**，而不是硬替换掉 claude。

这台机器目前没装 `opencode`，所以下面第 3 节列的都是必须先实测才能定的东西，不能直接照抄
claude 那一套的假设去写代码——写死了很可能是错的。

## 现状：只有一个耦合点

`claude` 这个词现在只出现在 `report_pane.go` 的 `runReportCmd()` 里一行：

```go
cmd := exec.Command("claude", "-p", "生成周报", "--permission-mode", "acceptEdits")
```

以及判断是否"看起来成功但其实被拒绝"的兜底逻辑：

```go
if err == nil && strings.Contains(string(out), "权限") {
    err = errors.New(strings.TrimSpace(string(out)))
}
```

`app.go`、`daily_pane.go`、`fragment_pane.go`、`todo_pane.go`、`welcome_pane.go`、
`fsutil.go` 完全不知道"claude"这个词的存在。这意味着改动面天然很小，但也意味着上面这
两行里藏的所有假设（命令行参数、权限行为、失败时的输出文案）都是 claude 专属的，换后端
时要逐条重新验证，不能只改函数名了事。

## 必须先手动验证的未知项（阻塞项）

在写任何代码之前，先装好 opencode，在终端手动跑，把下面几件事逐一搞清楚：

1. **非交互调用方式**：`weekly-auto-generate-opencode.sh` 里写的是
   `opencode run "生成周报"`，先确认这个调用方式现在还成立（`opencode --help` 核对一下
   子命令和参数名有没有变过）。
2. **权限行为**：claude -p 在非交互模式下，遇到写文件操作如果没有
   `--permission-mode acceptEdits` 会**静默拒绝写入但 exit code 仍是 0**（这台机器上
   实测过，见 `README.md`"移植到 opencode 方不方便"一节）。opencode 有没有等价的坑？
   有没有等价的"自动放行"参数？如果没有对应参数，是不是默认就允许写文件（不同工具的
   权限模型可能完全不同，不能假设 opencode 也需要一个"acceptEdits"）。
   验证方法：手动跑 `opencode run "生成周报"`，检查 `reports/` 目录有没有真的多出文件。
3. **失败时的输出长什么样**：现在的兜底是"输出里出现'权限'两个字就当作失败"，这是针对
   claude 拒绝写入时中文回复原文调出来的，换成 opencode 大概率完全对不上。需要故意制造
   一次失败（比如没网、或者故意不给权限），看 opencode 输出的失败文案，再决定用什么关键词
   或者退出码来判断。
4. **能不能读到 `SKILL.md` 里的整合逻辑**：现在 claude 之所以知道"生成周报"要做什么
   （去哪个目录读 daily 文件、按什么格式整合润色、存到哪），全靠 Claude Code 启动时自动
   扫描 `~/.claude/skills/` 发现 `weekly-report/SKILL.md` 这个 skill。opencode 有没有
   等价的技能/指令发现机制？如果没有，`opencode run "生成周报"` 这句话本身没有任何上下文，
   大概率只会得到一段不知所云的回复，而不是真正读 daily 文件生成周报。这是最可能"看起来
   能跑但结果是错的"的坑，必须优先验证。
   - 如果 opencode 不支持技能发现：需要把 `SKILL.md` 里的步骤内容直接拼进 prompt 里传给
     `opencode run`，比如 `opencode run "$(cat SKILL.md 的步骤部分) 现在开始生成周报"`，
     或者在 opencode 的配置目录里放一份等价的指令文件（具体放哪、格式是什么，取决于
     opencode 自己的机制，需要查它的文档）。

在这四条都有明确答案之前，下面第 4 节的代码方案只是"占位设计"，实际实现时函数体内容
要按验证结果调整。

## 代码方案：抽出一个可切换的 backend

```go
// report_backend.go（新文件）
package main

import "os/exec"

type reportBackend struct {
	name        string
	buildCmd    func(prompt string) *exec.Cmd
	isDenied    func(output string) bool // 输出里有没有"看起来成功但其实被拒绝"的信号
}

var backendClaude = reportBackend{
	name: "claude",
	buildCmd: func(prompt string) *exec.Cmd {
		return exec.Command("claude", "-p", prompt, "--permission-mode", "acceptEdits")
	},
	isDenied: func(out string) bool {
		return strings.Contains(out, "权限")
	},
}

// backendOpencode 的 buildCmd/isDenied 内容要等第 3 节验证完再填，
// 现在先占位成和 claude 一样，多半是错的。
var backendOpencode = reportBackend{
	name: "opencode",
	buildCmd: func(prompt string) *exec.Cmd {
		return exec.Command("opencode", "run", prompt)
	},
	isDenied: func(out string) bool {
		return false // TODO: 验证 opencode 拒绝写入时的真实输出后再实现
	},
}
```

`report_pane.go` 里的 `runReportCmd` 改成接受一个 `reportBackend` 参数（或者读
`reportModel` 上的一个字段），不再硬编码 `claude`：

```go
func runReportCmd(backend reportBackend) tea.Cmd {
	return func() tea.Msg {
		home, _ := os.UserHomeDir()
		cmd := backend.buildCmd("生成周报")
		cmd.Dir = home
		out, err := cmd.CombinedOutput()
		if err == nil && backend.isDenied(string(out)) {
			err = errors.New(strings.TrimSpace(string(out)))
		}
		return reportDoneMsg{err: err}
	}
}
```

### 切换方式：跟哪个体验对齐

两种都合理，选哪个是你的偏好问题，不是技术门槛问题：

- **运行时按键切换**（推荐）：在生成周报面板里加一个按键（比如 `b`）在 claude/opencode
  之间切换当前 backend，画面上显示"当前后端: claude"，改一次当次会话内一直生效。适合
  你自己两个都装、看心情切换的场景，不用记环境变量。
- **环境变量/配置文件**：比如 `FK_REPORT_BACKEND=opencode reportfk`，或者读
  `~/.config/fk-report/config.toml` 里的一行配置。适合你固定只用其中一个、想"设置一次
  以后不用管"的场景。

默认值都建议是 `claude`——因为它是唯一已经实测跑通的路径，opencode 装好之前不该是默认。

## 分阶段实施

1. 装 opencode，手动跑通第 3 节的四项验证，尤其是"能不能读到 SKILL.md 的整合逻辑"这条——
   如果这条不行，说明真正的工作量不在 `fk-report` 这个 Go 项目里，而在"怎么给 opencode
   喂等价的指令"，可能需要先在 `skill-backup/weekly-report/` 或者 opencode 自己的配置里
   补一份东西。
2. 验证通过后再动代码：加 `report_backend.go`，把 `backendOpencode` 的 `buildCmd`/
   `isDenied` 按实测结果填对，`report_pane.go` 改成走 backend 抽象。
3. 加切换方式（按键或配置，二选一，见上一节）。
4. 用真实的 `reports/` 目录跑一遍两个 backend 各自生成一次周报，手动对比输出格式是否
   一致、`isDenied` 判断在两边是否都准确（尤其要故意制造一次拒绝场景，确认不会把失败
   误报成"已生成"，这是当初 claude 那条路径踩过的坑，opencode 这边同样要测一遍反例）。
5. 可选：`skill-backup/` 里如果以后要保留 opencode 专属的指令文件，一并纳入备份、
   在 README 里补一句说明两个后端各自依赖什么。

## 风险与回滚

这个方案是"新增一个可选 backend"，不是替换，默认行为不变，claude 路径完全不动，所以
风险很低：opencode 那条路径验证或实现过程中出问题，随时可以只用 claude、把 opencode 的
代码和切换入口留着不启用即可，不影响现在已经跑通的功能。
