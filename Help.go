package command

import (
	"fmt"
	"sort"
	"strconv"
)

type Help struct {
	console *Console
}

func (Help) Configure() Configure {
	return Configure{
		Name:        "help",
		Description: "帮助命令",
		Input:       Argument{},
	}
}

func (h Help) Execute(input Input) {
	fmt.Println("Usage:")
	fmt.Println("  command [options] [arguments] [has]")
	fmt.Println("Base Options:")
	for _, ArgParam := range h.console.baseOption {
		h.EchoSpace("  -"+ArgParam.Name, 25)
		fmt.Println(ArgParam.Description)
	}
	fmt.Println("Available commands:")
	// 命令排序
	var keys []string
	var macLen int
	for cmdName, _ := range h.console.MapCommand {
		keys = append(keys, cmdName)
		tempLen := len(cmdName)
		if tempLen > macLen {
			macLen = tempLen
		}
	}
	sort.Strings(keys)
	macLen += 4
	for _, cmdName := range keys {
		h.EchoSpace("  "+cmdName, macLen)
		kv := h.console.MapCommand[cmdName]
		fmt.Println(kv.CommandConfig.Description)
	}
}

// 字符串不足，空格补充输出
func (Help) EchoSpace(str string, mac int) {
	strCon := strconv.Itoa(mac)
	fmt.Printf("%-"+strCon+"s", str)
}

// 某个命令需要帮助时
func (h Help) HelpExecute(con Configure) {
	fmt.Println("Usage:")
	fmt.Print("  ", con.Name)
	for _, ArgParam := range con.Input.Argument {
		fmt.Print(" <", ArgParam.Name, ">")
	}
	fmt.Println()
	for _, ArgParam := range con.Input.Option {
		if len(ArgParam.Default) >= 1 {
			h.EchoSpace("    -"+ArgParam.Name, 25)
			fmt.Println("= " + ArgParam.Default)
		}
	}
	fmt.Println("Arguments:")
	for _, ArgParam := range con.Input.Argument {
		h.EchoSpace("  "+ArgParam.Name, 25)
		fmt.Println(ArgParam.Description)
	}
	fmt.Println("Option:")
	for _, ArgParam := range con.Input.Option {
		h.EchoSpace("  -"+ArgParam.Name, 25)
		fmt.Println(ArgParam.Description)
	}
	for _, ArgParam := range h.console.baseOption {
		h.EchoSpace("  -"+ArgParam.Name, 25)
		fmt.Println(ArgParam.Description)
	}
	fmt.Println("Has:")
	for _, ArgParam := range con.Input.Has {
		h.EchoSpace("  "+ArgParam.Name, 25)
		fmt.Println(ArgParam.Description)
	}
	fmt.Println("Description:")
	fmt.Println("  ", con.Description)
}
