package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

func GetRootPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		dir, _ = os.Getwd()
		return dir
	}
	return dir
}

/*
保存pid
命令key 唯一标识
*/
func SavePidToFile(key string) {
	pid := os.Getpid()
	dir := GetRootPath()
	path := dir + key + ".lock"
	_ = ioutil.WriteFile(path, []byte(fmt.Sprintf("%d", pid)), 0666)
}

/*
获取pid
命令key 唯一标识
*/
func GetPidForFile(key string) int {
	dir := GetRootPath()
	path := dir + key + ".lock"
	str, err := ioutil.ReadFile(path)
	if err != nil {
		return 0
	}
	pid, err := strconv.Atoi(string(str))
	if err != nil {
		return 0
	}
	return pid
}
