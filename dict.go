package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"strconv"
)

/**
	读取文件加载词库
*/
func DictRead (d *DoubleArray) (bool,error) {
	f, err := os.Open(dictFile)
	if err != nil {
		return false, err
	}
	defer f.Close()

	md5h := md5.New()
	_, _ = io.Copy(md5h, f)
	md5Str := hex.EncodeToString(md5h.Sum(nil))
	if d.Fmd5 == md5Str {
		return true, nil
	} else {
		d.Fmd5 = md5Str
		d.Base = make([]int, 1)
		d.Check = make([]int, 1)
		d.Size = 0
	}

	d.KeyMap = make(map[string]int)
	_, _ = f.Seek(0, 0)
	bfRd := bufio.NewReader(f)
	for {
		line, err := bfRd.ReadBytes('\n')
		if err != nil { //遇到任何错误立即返回，并忽略 EOF 错误信息
			if err == io.EOF {
				return false, nil
			}
			return false, err
		}
		if len(line) < 4 {//字符太少跳过处理
			continue
		}
		in,_ := strconv.Atoi(string(line[0]))
		d.KeyMap[string(line[2:len(line)-2])] = in
	}
}

/**
	添加词
 */
func DictAdd(key string) {
	fd,_:=os.OpenFile(dictFile, os.O_RDWR|os.O_CREATE|os.O_APPEND,os.ModePerm)
	_, _ = fd.Write([]byte(key))
	defer fd.Close()
}

/**
	移除词
 */
func DictRemove(key string) error {
	f, err := os.Open(dictFile)
	if err != nil {
		return err
	}
	defer f.Close()
	fn, err :=os.OpenFile(dictFile+"_tmp", os.O_RDWR | os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer fn.Close()
	bfRd := bufio.NewReader(f)
	for {
		line, err := bfRd.ReadBytes('\n')
		if err != nil { //遇到任何错误立即返回，并忽略 EOF 错误信息
			if err == io.EOF {
				break
			}
			return err
		}
		if string(line[2:len(line)-2]) == key {
			continue
		}
		fn.Write(line)
	}

	return nil
}

