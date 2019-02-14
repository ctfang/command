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
		h.EchoSpace(cmdName, macLen)
		kv := h.console.MapCommand[cmdName]
		fmt.Println(kv.CommandConfig.Description)
	}
}

// 字符串不足，空格补充输出
func (Help) EchoSpace(str string, mac int) {
	strCon := strconv.Itoa(mac)
	fmt.Printf("%-"+strCon+"s", str)
}
