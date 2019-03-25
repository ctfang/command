package main

import (
	"fmt"
	"github.com/ctfang/command"
	"log"
)

func main() {
	app := command.New()
	app.AddCommand(Echo{})
	app.AddCommand(Hello{})
	app.Run()
}

type Echo struct {
}

func (Echo) Configure() command.Configure {
	return command.Configure{
		Name:        "echo",
		Description: "示例命令 echo",
	}
}

func (Echo) Execute(input command.Input) {
	log.Println("hello command")
}

type Hello struct {
}

func (Hello) Configure() command.Configure {
	return command.Configure{
		Name:        "hello",
		Description: "示例命令 hello",
		Input: command.Argument{
			// Argument参数为必须的输入的，不输入不执行
			Argument: []command.ArgParam{
				{Name: "name", Description: "命令后面第一个参数"},
				{Name: "sex", Description: "命令后面第二个参数"},
			},
			// 匹配字符参数，匹配不到就是 value = false
			Has: []command.ArgParam{
				{Name: "one", Description: "是否拥有one字符串"},
				{Name: "-t", Description: "是否拥有 -t 字符串"},
			},
			// 可选的参数，不输入也能执行
			Option: []command.ArgParam{
				{Name: "age", Description: "年龄选项参数"},
			},
		},
	}
}

func (Hello) Execute(input command.Input) {
	fmt.Println("hello")
	fmt.Println("名称：", input.GetArgument("name"))
	fmt.Println("性别：", input.GetArgument("sex"))
	fmt.Println("年龄 ：", input.GetOption("age"))
	fmt.Println("是否输入了 one ：", input.GetHas("one"))
	fmt.Println("是否输入了 -t ：", input.GetHas("-t"))
}
