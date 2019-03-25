## golang cli 应用封装

命令行应用，基本调用封装、参数获取

~~~~gotemplate
go get github.com/ctfang/command
~~~~

### 基础使用

代码在 go get github.com/ctfang/command/examples/main.go

~~~~go
package main

import (
	"github.com/ctfang/command"
	"log"
)

func main() {
	app := command.New()
	app.AddCommand(Echo{})
	app.AddCommand(Hello{})
	app.Run()
}
// Hello 需要实现接口 CommandInterface
type Echo struct {
}

func (Echo) Configure() command.Configure {
	return command.Configure{
		Name:"hello",
		Description:"示例命令 hello",
	}
}

func (Echo) Execute(input command.Input) {
	log.Println("hello command")
}
~~~~

运行
~~~~
go run main.go
-------------------------------------
Usage:
  command [options] [arguments] [has]
Base Has Param:
  -d  守护进程启动
  -h  显示帮助信息参数
Available commands:
  hello  示例命令 hello
  help   帮助命令
  
go run main.go echo
-------------------------------------
2019/03/25 17:01:46 hello command
~~~~

### 设置参数
参数分为三种类型，必须参数，可选参数，匹配参数

~~~~go
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
			Option:   []command.ArgParam{
				{Name: "age", Description: "年龄选项参数"},
			},
		},
	}
}

func (Hello) Execute(input command.Input) {
	fmt.Println("hello")
	fmt.Println("名称：",input.GetArgument("name"))
	fmt.Println("性别：",input.GetArgument("sex"))
	fmt.Println("年龄 ：",input.GetOption("age"))
	fmt.Println("是否输入了 one ：",input.GetHas("one"))
	fmt.Println("是否输入了 -t ：",input.GetHas("-t"))
}
~~~~
#### 查看帮助命令 
~~~~
go run main.go hello -h
-------------------------------------
Usage:
  hello <name> <sex>
Arguments:
  name                   命令后面第一个参数
  sex                    命令后面第二个参数
Option:
  -age                   年龄选项参数
Has:
  one                    是否拥有one字符串
  -t                     是否拥有 -t 字符串
  -d                     守护进程启动
  -h                     显示帮助信息
Description:
   示例命令 hello
   
-------------------------------------
go run main.go hello 李四 男 -age=18 -t one
-------------------------------------
hello
名称： 李四
性别： 男
年龄 ： 18
是否输入了 one ： true
是否输入了 -t ： true

~~~~
