package command

import (
	"fmt"
	"os"
	"strings"
)

type Console struct {
	MapCommand map[string]MapCommand
}

type CommandInterface interface {
	Configure() CommandConfig
	Execute(input Input)
}

type Command struct {
}

type CommandConfig struct {
	//命令名称
	Name string
	// 说明
	Description string
	// 输入定义
	Input InputConfig
}

type MapCommand struct {
	Command       CommandInterface
	CommandConfig CommandConfig
}

type Input struct {
	// 是否有参数 【名称string】默认值bool
	Has map[string]bool
	// 必须输入参数 【命令位置】【赋值名称】默认值
	Argument map[string]string
	// 可选输入参数 【赋值名称（开头必须是-）】默认值
	Option map[string]string
	// 启动文件
	FilePath string
}

type InputConfig struct {
	// 是否有参数 【名称string】默认值bool
	Has map[string]bool
	// 必须输入参数 【命令位置】【赋值名称】默认值
	Argument map[int]KeyValue
	// 可选输入参数 【赋值名称（开头必须是-）】默认值
	Option map[string]string
}

type KeyValue struct {
	Key   string
	Value string
}

// 载入命令
func (c *Console) AddCommand(Command CommandInterface) {
	var SaveCom MapCommand
	var CmdConfig CommandConfig

	CmdConfig = Command.Configure()
	SaveCom.CommandConfig = CmdConfig
	SaveCom.Command = Command
	c.MapCommand[CmdConfig.Name] = SaveCom
}

// 载入命令
func (c *Console) Run() {
	defaultCmdName := "help"
	_,ok := c.MapCommand[defaultCmdName]
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
	MapCmd := c.MapCommand[cmdName]
	input := Input{
		Has: map[string]bool{},
		Argument: map[string]string{},
		Option: map[string]string{},
		FilePath:os.Args[0],
	}
	input.Parsed(MapCmd.CommandConfig.Input, args)
	MapCmd.Command.Execute(input)
}

// 参数解析
func (i *Input) Parsed(Config InputConfig, args []string) {
	for name, value := range Config.Has {
		for _, strArg := range args {
			if name == strArg {
				i.Has[name] = true
			}
		}
		_, ok := i.Has[name]
		if !ok {
			i.Has[name] = value
		}
	}

	lenArgument := len(args)
	for mustInt, kv := range Config.Argument {
		if lenArgument < mustInt {
			// 不存在，报错,并且输出帮助命令
			fmt.Println("必须输入参数:" + kv.Key)
			os.Exit(1)
		} else {
			i.Argument[kv.Key] = args[mustInt]
		}
	}
	var strArgKy, strValue string
	for _, strArg := range args {
		startIndex := strings.Index(strArg, "-")
		if startIndex == 0 {
			stopIndex := strings.Index(strArg, "=")
			if stopIndex < 0 {
				// 不存在 = 号
				strArgKy = strArg[startIndex+1:]
				defaultValue, _ := i.Option[strArgKy]
				strValue = defaultValue
			} else {
				strArgKy = strArg[startIndex+1 : stopIndex]
				strValue = strArg[stopIndex+1:]

			}
			if strArgKy != "" {
				_, ok := i.Option[strArgKy]
				if ok {
					i.Option[strArgKy] = strValue;
				}
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
	return value
}


func (i *Input) GetFilePath() string {
	return i.FilePath
}