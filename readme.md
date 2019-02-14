## cli

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

type Demo struct {

}

func (Demo) Configure() command.Configure {
	return command.Configure{
		Name:"test",
		Description:"测试命令",
		Input:command.Argument{
			Argument: map[int]command.KeyValue{
				0:{Key:"one"},// 参数在 0 位置获取
			},
			Has: []string{"one"},
		},
	}
}

func (Demo) Execute(input command.Input) {
	fmt.Println("必须输入的参数 one =",input.GetArgument("one"))
	fmt.Println("是否有输入字符串 one ",input.GetHas("one"))
}
~~~~

运行
~~~~
go run main.go
go run main.go test
~~~~