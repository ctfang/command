package command

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Console struct {
	MapCommand map[string]MapCommand
	configPath string
	config     ini
	baseOption []ArgParam
	baseHas    []ArgParam
	run        MapCommand
}

// 创建一个命令应用
func New() Console {
	return Console{
		MapCommand: map[string]MapCommand{},
		baseOption: []ArgParam{
			{
				Name:        "h",
				Description: "显示帮助信息",
				Default:     "false",
				Call:        helpHandle,
			},
		},
		baseHas: []ArgParam{
			{Name: "-d", Description: "守护进程启动"},
			{Name: "--d", Description: "守护进程启动（等价 -d）"},
		},
	}
}

func helpHandle(val string, c *Console) (string, bool) {
	if val != "false" {
		help := Help{c}
		help.HelpExecute(c.run.CommandConfig)
		return val, false
	}
	return val, true
}

// 添加通用参数
func (c *Console) AddBaseOption(param ArgParam) {
	c.baseOption = append([]ArgParam{param}, c.baseOption...)
}

// 添加通用 Has 匹配参数（与命令行 token 完全相等即视为命中）
func (c *Console) AddBaseHas(param ArgParam) {
	c.baseHas = append([]ArgParam{param}, c.baseHas...)
}

type Command interface {
	Configure() Configure
	Execute(input Input)
}

type Configure struct {
	// 命令名称
	Name string
	// 说明
	Description string
	// 输入定义
	Input Argument
}

type MapCommand struct {
	Command       Command
	CommandConfig Configure
}

// 参数操作
type Input struct {
	console *Console
	// 是否有参数 【名称string】默认值bool
	Has map[string]bool
	// 必须输入参数 【命令位置】【赋值名称】默认值
	Argument map[string]string
	// 可选输入参数 【Configure 里 Name 不含 -；命令行 -name / -name=v，另兼容 --name=v，不改变原「位置参数」切分规则】默认值
	Option map[string][]string
	// 启动文件
	FilePath string
}

// 参数存储
type ArgParam struct {
	Name        string                                      // 名称
	Description string                                      // 说明
	Default     string                                      // 默认值
	Call        func(val string, c *Console) (string, bool) // 获取值的时候执行, return false中断
}

// 参数设置结构
type Argument struct {
	// 是否有参数 【名称string】
	Has []ArgParam
	// 必须输入参数 【命令位置】【赋值名称】默认值
	Argument []ArgParam
	// 可选输入参数 【Name 为 flag 名不含前导 -；命令行 -Name、-Name=v；兼容 --Name=v；见 readme】默认值
	Option []ArgParam
}

func (c *Console) IniConfig() error {
	path := c.getConfig()
	c.config = ini{}
	return c.config.Load(path)
}

// 载入命令
func (c *Console) AddCommand(Command Command) {
	var SaveCom MapCommand

	CmdConfig := Command.Configure()
	for key, ArgParam := range CmdConfig.Input.Option {
		if c.config.Has(ArgParam.Name) {
			CmdConfig.Input.Option[key].Default = c.config.GetString(ArgParam.Name, "")
		}
	}

	SaveCom.CommandConfig = CmdConfig
	SaveCom.Command = Command
	c.MapCommand[CmdConfig.Name] = SaveCom
}

func (c *Console) getConfig() string {
	return c.configPath
}

func (c *Console) SetConfig(path string) {
	c.configPath = path
}

// 载入命令
func (c *Console) Run() error {
	defaultCmdName := "help"
	_, ok := c.MapCommand[defaultCmdName]
	if !ok {
		// 注册帮助命令
		c.AddCommand(Help{c})
	}

	argsLen := len(os.Args)
	var args []string
	var cmdName string
	if argsLen < 2 {
		cmdName = defaultCmdName
	} else {
		cmdName = os.Args[1]
		args = os.Args[2:]
		_, ok1 := c.MapCommand[cmdName]
		if !ok1 {
			fmt.Println("不存在的命令:" + cmdName)
			cmdName = defaultCmdName
		}
	}

	// 执行到这里，必须有命令
	c.run = c.MapCommand[cmdName]
	input := Input{
		console:  c,
		Has:      map[string]bool{},
		Argument: map[string]string{},
		Option:   map[string][]string{},
		FilePath: os.Args[0],
	}
	err := input.Parsed(c.run.CommandConfig.Input, args)
	if err != nil {
		return err
	}

	c.run.Command.Execute(input)
	return nil
}

// splitArgsAtDoubleDash 在首个 "--" 处切开；与 POSIX/flag 一致，其后 argv 全部视为位置参数（含以 - 开头的 token）。
func splitArgsAtDoubleDash(args []string) (before, after []string) {
	for i, a := range args {
		if a == "--" {
			return append([]string(nil), args[:i]...), append([]string(nil), args[i+1:]...)
		}
	}
	return args, nil
}

// parseFlagNameValue 仅从**单个** argv token 解析：-a / --a / -a=b / --a=b → 选项名 key（无横线）、值。
// 不消费下一个 token，保持本库原有位置参数规则。
func parseFlagNameValue(arg string) (name, value string, ok bool) {
	rest := arg
	switch {
	case strings.HasPrefix(rest, "--"):
		rest = rest[2:]
	case strings.HasPrefix(rest, "-"):
		rest = rest[1:]
	default:
		return "", "", false
	}
	if rest == "" {
		return "", "", false
	}
	if idx := strings.Index(rest, "="); idx >= 0 {
		return rest[:idx], rest[idx+1:], true
	}
	return rest, "", true
}

// optionsFromArgs 仅扫 before 段中带 - 的 token；每个 token 独立解析，与历史行为一致（-age 与 18 分两 token 时 18 仍是位置参数）。
func optionsFromArgs(before []string) map[string][]string {
	opts := make(map[string][]string)
	for _, strArg := range before {
		if !strings.HasPrefix(strArg, "-") {
			continue
		}
		name, value, ok := parseFlagNameValue(strArg)
		if !ok || name == "" {
			continue
		}
		opts[name] = append(opts[name], value)
	}
	return opts
}

// filterPositionalArgs：沿用原规则——before 段中非 "-" 开头的 token 为位置参数；"--" 之后整段追加为位置参数（兼容常见 CLI 终止符）。
func filterPositionalArgs(args []string) []string {
	before, after := splitArgsAtDoubleDash(args)
	pos := make([]string, 0, len(args))
	for _, strArg := range before {
		if strings.HasPrefix(strArg, "-") {
			continue
		}
		pos = append(pos, strArg)
	}
	pos = append(pos, after...)
	return pos
}

// argTokenSet 用于 Has 匹配：O(|args|) 建表，避免对每个 Has 定义扫整段 argv。
func argTokenSet(args []string) map[string]struct{} {
	m := make(map[string]struct{}, len(args))
	for _, t := range args {
		m[t] = struct{}{}
	}
	return m
}

// 参数解析
func (i *Input) Parsed(Config Argument, args []string) error {
	i.ParsedOptions(Config, args)

	hasDefs := append(append([]ArgParam{}, i.console.baseHas...), Config.Has...)
	tokens := argTokenSet(args)
	for _, ArgParam := range hasDefs {
		if _, ok := tokens[ArgParam.Name]; ok {
			i.Has[ArgParam.Name] = true
		} else {
			i.Has[ArgParam.Name] = false
		}
	}

	// 必须位置参数：原规则 + 首个 "--" 之后全部计入位置参数
	positional := filterPositionalArgs(args)
	lenArgument := len(positional)
	for mustInt, kv := range Config.Argument {
		if lenArgument <= mustInt {
			// 不存在，报错,并且输出帮助命令
			fmt.Println("必须输入参数:" + kv.Name)
			return errors.New("必须输入参数:" + kv.Name)
		} else {
			i.Argument[kv.Name] = positional[mustInt]
		}
	}
	return nil
}

// 解析选项值（只解析首个 "--" 之前的 flag token；每个 token 自成一项，不把下一 token 当作值）
func (i *Input) ParsedOptions(Config Argument, args []string) {
	before, _ := splitArgsAtDoubleDash(args)
	optVals := optionsFromArgs(before)
	for _, kv := range i.console.baseOption {
		Config.Option = append(Config.Option, kv)
	}
	defaultsByName := make(map[string][]string)
	for _, kv := range Config.Option {
		defaultsByName[kv.Name] = append(defaultsByName[kv.Name], kv.Default)
	}
	for _, kv := range Config.Option {
		i.Option[kv.Name] = make([]string, 0)
	}
	for name, vals := range optVals {
		i.Option[name] = append(i.Option[name], vals...)
	}
	// 添加默认值（按名称合并各 ArgParam 的 Default，与原先 O(n²) 双层循环等价）
	for _, kv := range Config.Option {
		if len(i.Option[kv.Name]) == 0 {
			i.Option[kv.Name] = append(i.Option[kv.Name], defaultsByName[kv.Name]...)
		}
		// 执行回调, 使用回调赋值
		if kv.Call != nil {
			var stop bool
			i.Option[kv.Name][0], stop = kv.Call(i.Option[kv.Name][0], i.console)
			if stop == false {
				os.Exit(0)
			}
		}
	}
}

// 参数
func (i *Input) GetHas(key string) bool {
	value, ok := i.Has[key]
	if !ok {
		return false
	}
	return value
}

// 参数
func (i *Input) GetArgument(key string) string {
	value, ok := i.Argument[key]
	if !ok {
		return ""
	}
	return value
}

// 参数
func (i *Input) GetOption(key string) string {
	value, ok := i.Option[key]
	if !ok {
		return ""
	}
	return value[0]
}

func (i *Input) GetOptions(key string) []string {
	value, ok := i.Option[key]
	if !ok {
		return []string{}
	}
	return value
}

// 是否后台启动（兼容 -d 与 --d，与 flag 单双横线习惯一致）
func (i *Input) IsDaemon() bool {
	return i.GetHas("-d") || i.GetHas("--d")
}

func (i *Input) GetFilePath() string {
	return i.FilePath
}
