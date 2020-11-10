package main

import (
	"fmt"
	"github.com/jinxing3114/xtrie"
)

//创建XTrie
var XT = new(xtrie.XTrie)

//基本配置信息
var storeFile,dictFile = "data/dat.data", "data/darts.txt"

/**
入口函数
*/
func main(){
	XT.InitHandle(storeFile, dictFile)
	//example()
	//开始监听请求
	startServe(":8888")
}


/**
例子
*/
func example(){
	index, level, err := XT.Match("番", false)
	fmt.Println(index, level, err)
	//level, err, _, _ = dat.match("中", false)
	//fmt.Println(level, err)
	//sk := "我要测试一下"
	//contentRune := []rune(sk)
	//result := dat.search(sk)
	//fmt.Println(result)
	//for i:=0;i<len(result);i++{
	//	fmt.Println("str:", string(contentRune[result[i][0]:result[i][1] + 1]), "level", result[i][2])
	//}
}
