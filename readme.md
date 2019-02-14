## About Laravel

命令行应用，基本调用封装、参数获取

~~~~gotemplate
go get github.com/ctfang/command
~~~~

~~~~go
package main

import (
	"fmt"
	"github.com/ctfang/command"
)

func main() {
	app := command.New()
	app.AddCommand(Demo{})
	app.Run()
}

// 定义一个测试命令
type Demo struct {
}

func (d Demo) Configure() command.CommandConfig {
	return command.CommandConfig{
		Name:        "test",
		Description: "测试命令",
		Input:       command.InputConfig{},
	}
}

func (d Demo) Execute(input command.Input) {
	fmt.Println("执行输出")
}
~~~~

运行
~~~~
go run main.go
go run main.go test
~~~~