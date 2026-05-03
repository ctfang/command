## golang cli 应用封装

命令行子命令封装：注册命令、解析位置参数与选项、帮助输出，并可选从 INI 填充 `Option` 默认值。

### 功能一览

- **子命令**：实现 `Command`（`Configure` + `Execute`），`AddCommand` 注册；未传子命令时默认进入 `help`。
- **必选位置参数 `Argument`**：按声明顺序，从「非 `-` 开头」的 token 中依次取值（见下文「兼容说明」）。
- **可选参数 `Option`**：配置里 `Name` 不写横线；命令行 `-name=v` 或 `--name=v`。
- **匹配参数 `Has`**：argv 中出现与 `Name` **完全一致**的某一 token 即为 true（如 `one`、`-t`）。
- **全局**：默认带 `-h`（帮助）、`-d` / `--d`（`IsDaemon()` / `GetHas`）；可用 `AddBaseOption`、`AddBaseHas` 扩展。
- **配置文件**：`SetConfig` + `IniConfig()` 后，仅对已有 `Option.Name` 从 INI 覆盖**默认值**（键名与 `Name` 一致，无 `-`）。
- **错误**：`Run()`、`IniConfig()` 返回 `error`（缺必选参数等会返回错误，需自行处理或退出）。

### API 约定与升级提示

当前实现为 **Go module**，安装示例：`go get github.com/ctfang/command@v1.1.0`（可按需换 tag）。

- `func (c *Console) Run() error`：解析失败或业务前置校验失败时返回 `error`，**不会**默默以 0 退出；`main` 中建议 `if err := app.Run(); err != nil { log.Fatal(err) }`。
- `func (c *Console) IniConfig() error`：打开或读取配置失败时返回 `error`；文件不存在时视为无配置，返回 `nil`。
- 若在 `Option.Callback`（`Call`）里触发帮助，库内仍会 `os.Exit(0)`，与返回 `error` 的路径并存，属既有设计。

以下内容描述**当前版本**行为，不按「历次修改」叙事；旧代码请对照上面对 `Run` / `IniConfig` 的签名做迁移。

### 安装

~~~~gotemplate
go get github.com/ctfang/command
~~~~

完整可运行示例见仓库 [examples/main.go](examples/main.go)。

### 基础使用（可复制）

下列单文件合并了 `main`、子命令类型与 `Hello` 的完整示例（含 `import`，可直接保存为 `main.go` 后编译运行）。

~~~~go
package main

import (
	"fmt"
	"log"

	"github.com/ctfang/command"
)

func main() {
	app := command.New()
	app.AddCommand(Echo{})
	app.AddCommand(Hello{})
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

// Echo 实现 Command 接口
type Echo struct{}

func (Echo) Configure() command.Configure {
	return command.Configure{
		Name:        "echo",
		Description: "示例命令 echo",
	}
}

func (Echo) Execute(input command.Input) {
	log.Println("echo command")
}

type Hello struct{}

func (Hello) Configure() command.Configure {
	return command.Configure{
		Name:        "hello",
		Description: "示例命令 hello",
		Input: command.Argument{
			Argument: []command.ArgParam{
				{Name: "name", Description: "命令后第一个位置参数"},
				{Name: "sex", Description: "命令后第二个位置参数"},
			},
			Has: []command.ArgParam{
				{Name: "one", Description: "argv 中是否出现 token one"},
				{Name: "-t", Description: "argv 中是否出现 token -t"},
			},
			Option: []command.ArgParam{
				{Name: "age", Description: "年龄选项", Default: "18"},
			},
		},
	}
}

func (Hello) Execute(input command.Input) {
	fmt.Println("hello")
	fmt.Println("名称：", input.GetArgument("name"))
	fmt.Println("性别：", input.GetArgument("sex"))
	fmt.Println("年龄：", input.GetOption("age"))
	fmt.Println("是否输入了 one：", input.GetHas("one"))
	fmt.Println("是否输入了 -t：", input.GetHas("-t"))
	fmt.Println("守护进程：", input.IsDaemon())
}
~~~~

### 运行方式与输出说明

在**仓库根目录**：

~~~~
go run ./examples/
~~~~

或进入 `examples` 目录后：

~~~~
go run .
~~~~

以下为**示意输出**（帮助版式以实际为准；日志行时间略去）。

~~~~
# 无子命令 → 帮助
Usage:
  command [options] [arguments] [has]
Base Options:
  -h                       显示帮助信息
Base Has Param:
  -d                       守护进程启动
  --d                     守护进程启动（等价 -d）
Available commands:
  ...

# 子命令 echo
... echo command
~~~~

### 与常见 CLI / flag 的兼容（不改变原有规则）

在**原有**切分规则之上做兼容，**不**改成标准库 `flag` 那种「`-age` 单独出现时把下一 token 吞成值」的行为。

- **位置参数**：仍从「非 `-` 开头」的 token 收集；若写成两个 token `-age` 与 `20`，则 `20` **仍是位置参数**，**不会**自动变成 `age` 的值。要写入选项请用同一 argv 单元：**`-age=20`** 或 **`--age=20`**。
- **`--` 终止符**：第一个 `--` **之后**的整段 argv 一律只算位置参数（即使形如 `-x`）。
- **`--name=v`**：与 `-name=v` 一样，写入 `Option.Name == "name"`。

`Has` 仍为「与某一 argv token **字符串完全相等**」。

#### 命令行示例（对照理解）

- `hello 张 三 -age=20` 与 `hello 张 三 --age=20`：`age` 为 `20`，两个位置参数为 `张`、`三`。
- `hello 张 三 -age 20`：`age` **无** `=`，本库**不会**把 `20` 绑到 `age`；`20` 会被当作第三个位置参数（若 `Argument` 只声明 2 个，可能导致缺参错误）。
- `hello 张 三 -- -- -oops`：`--` 之后全部算位置参数，`-oops` 是第三个位置参数（若只声明 2 个则会报错）。

### 守护进程标记

- `input.IsDaemon()` 等价于「出现了 `-d` 或 `--d`」（与全局 `baseHas` 一致）。
- 也可直接 `input.GetHas("-d")`、`input.GetHas("--d")`。

### 设置参数类型小结

参数分三类：

1. **`Argument`**：必选，按顺序对应位置参数（规则见上）。
2. **`Option`**：可选；声明里的 `Default` 在无命令行值时使用；可被 INI 覆盖默认值。
3. **`Has`**：argv 中是否出现指定 token。

#### 查看子命令帮助

~~~~
go run ./examples/ hello -h
~~~~

示意：

~~~~
Usage:
  hello <name> <sex>
...
Description:
   示例命令 hello
~~~~

#### 一次完整调用示意

~~~~
go run ./examples/ hello 李四 男 -age=18 -t one
~~~~

~~~~
hello
名称： 李四
性别： 男
年龄： 18
是否输入了 one： true
是否输入了 -t： true
守护进程： false
~~~~

### 全局 Option / Has 与子命令合并规则

- **`New()`** 已注册全局 **`baseOption`**（含 `-h`）与 **`baseHas`**（`-d`、`--d`）。解析时会把它们并入当前子命令的 `Option` / `Has` 定义后再解析，因此每个子命令都具备这些入口，无需在每个 `Configure` 里重复写 `Has` 才能识别 `-d`。
- **`AddBaseOption`**、**`AddBaseHas`**：向全局列表**前置**追加（与 `New` 默认项一起参与解析与帮助展示）。
- 子命令若在 `Configure().Input` 里再次声明**同名** `Option` / `Has`，会与全局项**同时存在于解析列表**中（非「覆盖替换」全局项）；帮助中会多次列出同名项时，以你声明的条数为准。若需完全自定义某个全局 flag 的行为，宜在子命令内用不同 `Name` 或调整全局注册逻辑。

示例：增加一个仅供演示的全局可选参数（不写 `Call` 则仅为占位，可按需读 `GetOption("verbose")`）：

~~~~go
app.AddBaseOption(command.ArgParam{
	Name:        "verbose",
	Description: "示例：全局可选开关",
	Default:     "false",
	Call:        nil,
})
~~~~

### ini 配置文件默认值

仍可代码里兜底，例如：

~~~~go
name := input.GetOption("name")
if name == "" {
	name = "李四"
}
~~~~

多环境下更宜使用 **config.ini**：键名与 **`Option` 的 `Name`** 一致（**不要**带 `-`）。仅在 `AddCommand` 时，若 INI 中存在该键，会用 `GetString` 覆盖对应 `ArgParam` 的 **Default**（命令行仍优先）。

~~~~go
app := command.New()
app.SetConfig("config.ini")
if err := app.IniConfig(); err != nil {
	log.Fatal(err)
}
app.AddCommand(MyCmd{}) // 内含 Option Name: "url"
if err := app.Run(); err != nil {
	log.Fatal(err)
}
~~~~

**config.ini 示例**（与下方 `MyCmd` 中 `url` 选项对应）：

~~~~ini
; 注释以分号开头
url="127.0.0.1:8080"
~~~~

~~~~go
type MyCmd struct{}

func (MyCmd) Configure() command.Configure {
	return command.Configure{
		Name:        "mysvc",
		Description: "演示 INI 默认 url",
		Input: command.Argument{
			Option: []command.ArgParam{
				{Name: "url", Description: "服务地址", Default: "localhost:8080"},
			},
		},
	}
}

func (MyCmd) Execute(input command.Input) {
	// 若 ini 有 url，默认已是 ini 中的值；命令行 -url=x 仍覆盖
	fmt.Println(input.GetOption("url"))
}
~~~~

若未配置 ini 或缺键，则 `GetOption("url")` 使用 `Configure` 里的 `Default`。

### 更多示例（仓库 `examples/`）

- **[examples/main.go](examples/main.go)**：含 `Echo`、`Hello`、`ConfigDemo`；通过 `resolveConfigExample` 在**仓库根**（`go run ./examples/ …`）或 **`examples` 目录**（`go run .`）下都能找到 [config.example.ini](examples/config.example.ini)。
- **全局 `-verbose`**：`go run . hello 甲 乙 -verbose=true` 或 `-verbose=1`，在 `Hello` / `ConfigDemo` 输出里可看到 `GetOption("verbose")`。
- **INI 与 `hello` 的 `age`**：`config.example.ini` 中 `age="30"` 会在 `AddCommand` 时写入对应 `Option` 的默认（仍可用命令行 `-age=25` 覆盖）。
- **仅演示 URL**：`go run . configdemo`；覆盖：`go run . configdemo -url=http://override`。
- **守护进程标记**：`go run . hello 甲 乙 --d` 或 `-d`，`IsDaemon()` 为 true。
- **双横线终止**：`go run . hello 甲 乙 -age=20 -- --ignored-flag`（若 `Argument` 仅两个，第三个位置来自 `--` 后，可按需自行试验）。
