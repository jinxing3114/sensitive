// double array 基础和基本方法
// 2020年11月开发。
// 最优的使用场景通过自行维护词典文件来更新词库，该库的优点查询速度快。
// 待实现的 单个词插入(效率问题，可能会计算调整到整个结构所有的值)
// 删除

package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
)

const InitSize = 65536 * 32

//double array
type DoubleArray struct {
	Fmd5  string  // 词典文件md5
	Base  []int   // 基础切片，存储字符offset，正值和负值分别代表不同的状态
	Check []int   // 检查字符状态数组，防止查找冲突以及确认多种状态
	Keys  [][]rune// 所有词典转成rune切片
	Size  int     // 切片长度
	KeyMap map[string]int //所有词对应等级
}

/**
	重置扩容base和check切片
	参数 newSize int 新的切片大小
*/
func (d *DoubleArray) resize(newSize int) int {
	base2  := make([]int, newSize, newSize)
	check2 := make([]int, newSize, newSize)
	if len(d.Base) > 0 {
		copy(base2, d.Base)
		copy(check2, d.Check)
	}
	d.Base  = base2
	d.Check = check2
	d.Size  = newSize
	return newSize
}

/**
	插入词，递归函数，直到找不到下一层深度的词
	参数 keyPre rune切片 前缀字符切片，查询词等级而设计的
	参数 children *Node切片 所有子节点
	参数 index int 上层index值
 */
func (d *DoubleArray) insert(keyPre []rune, children []*Node, index int) {
	pos := children[0].Code //每次都以字符code起开始查找查找
	//初始查找位置，根据父层索引加子节点最小code算出初始位置
	childLen := len(children)
	offset := 0
outer:
	for {
		pos++
		if d.Base[pos] != 0 {
			continue
		}
		offset = pos - children[0].Code
		if s := offset + children[childLen - 1].Code; s > d.Size { //每次循环计算最大字符code位置是否超出范围
			d.resize(int(float64(s) * 1.25))
		}

		for i := 0; i < childLen; i++ {
			//确保每一个子节点都能落到base和check中
			ind := offset + children[i].Code
			if d.Check[ind] != 0 || d.Base[ind] != 0 {
				continue outer
			}
		}
		break
	}
	if d.Base[index] < 0 { //
		d.Base[index] = (pos - 1) * 10 + (-d.Base[index])
	} else {
		d.Base[index] = offset
	}

	keyPre = append(keyPre, 1)
	//写入所有的子节点到base中
	//写入所有的子节点到check中
	//必须先把所有节点写入完之后，再去查找添加下一层节点
	for i := 0; i < childLen; i++ {
		ind := offset + children[i].Code
		if children[i].End {
			keyPre[len(keyPre)-1] = rune(children[i].Code)
			d.Base[ind] = -d.KeyMap[string(keyPre)]
			d.Check[ind] = -index
		} else {
			d.Base[ind] = 0
			d.Check[ind] = index
		}
	}
	//循环查找下一层节点并且插入dat结构中
	for i := 0; i < childLen; i++ {
		nodes := children[i].fetch(d)
		if len(nodes) > 0 {
			ind := offset + children[i].Code
			keyPre[len(keyPre)-1] = rune(children[i].Code)
			d.insert(keyPre, nodes, ind)
		}
	}
	return
}

/**
	格式化词库
	将待格式的词集合，排序之后转为rune字符切片。utf8格式
*/
func (d *DoubleArray) format() error {
	if len(d.KeyMap) == 0 {
		return errors.New("dict is empty")
	}
	allKey := make([]string, 0, len(d.KeyMap))
	for k := range d.KeyMap {
		allKey = append(allKey, k)
	}
	sort.Strings(allKey)
	//
	d.Keys = make([][]rune, 0, len(allKey))
	for _,key := range allKey {
		zs := []rune(key)
		d.Keys = append(d.Keys, zs)
	}
	return nil
}

/**
编译词库
*/
func (d *DoubleArray) build () error {
	if len(d.KeyMap) == 0 {
		return errors.New("empty Keys")
	}
	err := d.format()
	if err != nil {
		return err
	}
	d.resize(InitSize)
	root := new(Node)
	root.Left = 0
	root.Right = len(d.Keys)
	root.Depth = 0
	children := root.fetch(d)
	rootIndex := 1
	d.insert([]rune{}, children, rootIndex)
	fmt.Println(d.Base[0:50], len(d.Base), cap(d.Base))
	fmt.Println(d.Check[0:50])
	return nil
}

// 使用gob协议
func (d *DoubleArray) Store(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(d)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// 从指定路径加载DAT
func (d *DoubleArray) Load(path string) error {
	file, err := os.Open(path)
	if err != nil {
		log.Println("dat build file open error", err)
		return err
	}

	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(d)
	if err != nil {
		log.Println("dat build file load error:", err)
		return err
	}

	return nil
}

/**
	初始化dat
    加载store文件，读取词典，编译dat，保存store等
*/
func (d *DoubleArray) initHandle(storeFile string, dictFile string) {

	err := dat.Load(storeFile)
	if err != nil { //加载失败
		log.Println("load store", storeFile, "error:", err)
	} else {
		log.Println("load store", storeFile, "success")
	}

	status, err := DictRead(d)

	if status == false {
		if err != nil {
			log.Fatalln("dict file read error:", err)
		} else {
			log.Println("dict file read success")
		}

		err = dat.build()
		if err != nil {
			log.Fatalln("build error:", err)
		} else {
			log.Println("build success")
		}

		err = dat.Store(storeFile)
		if err != nil {
			log.Fatalln("store error:", err)
		} else {
			log.Println("store save success ")
		}
	} else {
		log.Println("store and dict no difference, do not recompile")
	}
}
