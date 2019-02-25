package command

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type ini struct {
	config map[string]interface{}
}

func (i *ini) GetFileName(path string)string {
	return filepath.Join(path)
}

func (i *ini) Load(path string) {
	i.config = make(map[string]interface{})

	path = i.GetFileName(path)

	file, err := os.Open(path)
	if err != nil && os.IsNotExist(err) {
		return
	}
	defer file.Close()

	r := bufio.NewReader(file)
	prefix := ""

	for {
		b, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		//去除单行属性两端的空格
		s := strings.TrimSpace(string(b))
		if len(s) == 0 {
			continue
		}

		firstStr := s[0:1]
		// 注释
		if firstStr == ";" {
			continue
		} else if firstStr == "[" {
			// new map
			index := strings.Index(s, "]")
			if index < 0 {
				continue
			}
			prefix = string(b[1:index]) + "."
		} else {
			key, value := i.getKeyValue(s)
			if key != "" {
				i.config[prefix+key] = value
			}
		}

	}
}

func (i *ini) getKeyValue(s string) (string, interface{}) {
	//判断等号=在该行的位置
	index := strings.Index(s, "=")
	if index < 0 {
		return "", ""
	}
	//取得等号左边的key值，判断是否为空
	key := strings.TrimSpace(s[:index])
	if len(key) == 0 {
		return "", ""
	}

	//取得等号右边的value值，判断是否为空
	value := strings.TrimSpace(s[index+1:])
	if len(value) == 0 {
		return key, ""
	}

	// 值转换
	if value=="true" {
		return key,true
	}else if value=="false" {
		return key,false
	}else if value[0:1]=="\"" {
		return key,value[1:len(value)-1]
	}else if strings.Index(value, ".")>0 {
		float,_ := strconv.ParseFloat(value, 32)
		return key,float
	}else{
		num,_ := strconv.Atoi(value)
		return key,num
	}
}

// 获取值
func (i *ini) GetInt(key string, value int) int {
	va, ok := i.config[key]
	if !ok {
		return value
	}
	value, ok = va.(int)
	if ok {
		return value
	}
	return value
}

// 获取值
func (i *ini) GetString(key string, value string) string {
	va, ok := i.config[key]
	if !ok {
		return value
	}
	value, ok = va.(string)
	if ok {
		return value
	}
	return value
}


// 获取值
func (i *ini) GetBool(key string, value bool) bool {
	va, ok := i.config[key]
	if !ok {
		return value
	}
	value, ok = va.(bool)
	if ok {
		return value
	}
	return value
}

// 判断是否存在
func (i *ini) Has(key string) bool {
	_, ok := i.config[key]
	return ok
}
