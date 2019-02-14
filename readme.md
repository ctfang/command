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
    // 设置命令
    app.AddCommand(Demo{})
    app.Run()
}

// 定义一个测试命令
type Demo struct {
}

func (d Demo) Configure() CommandConfig {
    return CommandConfig{
        Name:        "test",
        Description: "测试命令",
        Input:       InputConfig{},
    }
}

func (d Demo) Execute(input Input) {
    fmt.Println("执行输出")
}
~~~~