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

func helpDaemon(val string, c *Console) bool {
	return false
}

// 添加通用参数
func (c *Console) AddBaseOption(param ArgParam) {
	c.baseOption = append(c.baseOption, param)
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
	// 可选输入参数 【赋值名称（开头必须是-）】默认值
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
	// 可选输入参数 【赋值名称（开头必须是-）】默认值
	Option []ArgParam
}

func (c *Console) IniConfig() {
	path := c.getConfig()
	c.config = ini{}
	c.config.Load(path)
}

var cacheInput = make(map[string]map[string]bool)

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
func (c *Console) Run() {
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
		return
	}

	c.run.Command.Execute(input)
}

// 参数解析
func (i *Input) Parsed(Config Argument, args []string) error {
	// 选项值
	i.ParsedOptions(Config, args)

	for _, ArgParam := range Config.Has {
		for _, strArg := range args {
			if ArgParam.Name == strArg {
				i.Has[ArgParam.Name] = true
			}
		}
		_, ok := i.Has[ArgParam.Name]
		if !ok {
			i.Has[ArgParam.Name] = false
		}
	}

	// 必须值
	lenArgument := len(args)
	for mustInt, kv := range Config.Argument {
		if lenArgument <= mustInt {
			// 不存在，报错,并且输出帮助命令
			fmt.Println("必须输入参数:" + kv.Name)
			return errors.New("必须输入参数:" + kv.Name)
		} else {
			i.Argument[kv.Name] = args[mustInt]
		}
	}
	return nil
}

// 解析选项值
func (i *Input) ParsedOptions(Config Argument, args []string) {
	for _, kv := range i.console.baseOption {
		Config.Option = append(Config.Option, kv)
	}
	for _, kv := range Config.Option {
		i.Option[kv.Name] = make([]string, 0)
	}
	var strArgKy, strValue string
	for _, strArg := range args {
		startIndex := strings.Index(strArg, "-")
		if startIndex == 0 {
			stopIndex := strings.Index(strArg, "=")
			if stopIndex < 0 {
				// 不存在 = 号
				strArgKy = strArg[startIndex+1:]
				strValue = ""
			} else {
				strArgKy = strArg[startIndex+1 : stopIndex]
				strValue = strArg[stopIndex+1:]
			}
			if strArgKy != "" {
				if _, ok := i.Option[strArgKy]; !ok {
					i.Option[strArgKy] = make([]string, 0)
				}
				i.Option[strArgKy] = append(i.Option[strArgKy], strValue)
			}
		}
	}
	for _, kv := range Config.Option {
		if len(i.Option[kv.Name]) == 0 {
			i.Option[kv.Name] = append(i.Option[kv.Name], kv.Default)
		}
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

// 是否后台启动
func (i *Input) IsDaemon() bool {
	value, ok := i.Has["-d"]
	if !ok {
		return false
	}
	return value
}

func (i *Input) GetFilePath() string {
	return i.FilePath
}
