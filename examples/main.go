package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ctfang/command"
)

// resolveConfigExample 在「仓库根」或「examples 目录」下执行 go run 时都能找到 ini。
func resolveConfigExample() string {
	for _, p := range []string{"config.example.ini", filepath.Join("examples", "config.example.ini")} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "config.example.ini"
}

func main() {
	app := command.New()

	app.AddBaseOption(command.ArgParam{
		Name:        "verbose",
		Description: "示例：全局可选开关",
		Default:     "false",
		Call:        nil,
	})
	app.SetConfig(resolveConfigExample())
	if err := app.IniConfig(); err != nil {
		log.Fatal(err)
	}

	app.AddCommand(Echo{})
	app.AddCommand(Hello{})
	app.AddCommand(ConfigDemo{})
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
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
	log.Println("echo command")
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
				{Name: "age", Description: "年龄选项参数", Default: "18"},
				{Name: "age", Description: "年龄选项参数", Default: "24"},
			},
		},
	}
}

func (Hello) Execute(input command.Input) {
	fmt.Println("hello")
	fmt.Println("名称：", input.GetArgument("name"))
	fmt.Println("性别：", input.GetArgument("sex"))
	fmt.Println("年龄 ：", input.GetOption("age"))
	fmt.Println("verbose(全局)：", input.GetOption("verbose"))
	fmt.Println("是否输入了 one ：", input.GetHas("one"))
	fmt.Println("是否输入了 -t ：", input.GetHas("-t"))
	fmt.Println("守护进程：", input.IsDaemon())
}

// ConfigDemo 演示 INI 默认值与命令行覆盖（需存在 config.example.ini）
type ConfigDemo struct{}

func (ConfigDemo) Configure() command.Configure {
	return command.Configure{
		Name:        "configdemo",
		Description: "读取 INI 中 url/port；命令行 -url= / -port= 覆盖",
		Input: command.Argument{
			Option: []command.ArgParam{
				{Name: "url", Description: "服务 URL", Default: "http://127.0.0.1:8080"},
				{Name: "port", Description: "端口", Default: "9000"},
			},
		},
	}
}

func (ConfigDemo) Execute(input command.Input) {
	fmt.Println("url ：", input.GetOption("url"))
	fmt.Println("port：", input.GetOption("port"))
	fmt.Println("verbose(全局)：", input.GetOption("verbose"))
	fmt.Println("程序路径：", input.GetFilePath())
}
