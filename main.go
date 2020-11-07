package main

import (
	"fmt"
)

var dat = new(DoubleArray)

/**
例子
*/
func example(){
	level, err, _, _ := dat.match("番", false)
	fmt.Println(level, err)
	level, err, _, _ = dat.match("中", false)
	fmt.Println(level, err)
	sk := "我要测试一下"
	contentRune := []rune(sk)
	result := dat.search(sk)
	fmt.Println(result)
	for i:=0;i<len(result);i++{
		fmt.Println("str:", string(contentRune[result[i][0]:result[i][1] + 1]), "level", result[i][2])
	}
}

var storeFile,dictFile = "", ""

/**
入口函数
*/
func main(){
	//storeFile := "data/dat.data"
	//dictFile  := "data/darts.txt"
	if len(storeFile) == 0 {
		storeFile = "data/dat.data"
	}
	if len(storeFile) == 0 {
		dictFile = "data/darts.txt"
	}
	dat.initHandle(storeFile, dictFile)
	//开始监听请求
	startServe(":8888")
}