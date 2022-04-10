package common

import (
	"flag"
	"testing"
)

// 定义命令行参数对应的变量，这三个变量都是指针类型
var cliName = flag.String("name", "nick", "Input Your Name")
var cliAge = flag.Int("age", 28, "Input Your Age")
var cliGender = flag.String("gender", "male", "Input Your Gender")

// 命令行参数对应变量的定义和初始化是可以分开的，比如下面例子
var cliFlag int

func TestInitFlag(t *testing.T) {
	flag.IntVar(&cliFlag, "flagname", 1234, "Just for demo")
	// 把用户传递的命令行参数解析为对应变量的值
	flag.Parse()

	// flag.Args() 函数返回没有被解析的命令行参数
	// func NArg() 函数返回没有被解析的命令行参数的个数
	t.Logf("args=%s, num=%d\n", flag.Args(), flag.NArg())
	for i := 0; i != flag.NArg(); i++ {
		t.Logf("arg[%d]=%s\n", i, flag.Arg(i))
	}

	// 输出命令行参数
	t.Log("name=", *cliName)
	t.Log("age=", *cliAge)
	t.Log("gender=", *cliGender)
	t.Log("flagname=", cliFlag)
}
