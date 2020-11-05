package main

import (
	"bufio"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
)

const InitSize = 65536 * 32

//将字符转换成内部code，减小base和check切片大小
//切片充当map使用，比map高效，维护难度更小，可动态管理
type MapDoubleArrayTrie struct {
	Fmd5  string // 词典文件md5
	Base  []int  // 基础切片，存储字符offset
	Check []int  // 检查字符状态数组，防止查找冲突
	Keys  [][]rune//所有词典转成rune切片
	Size  int    // 切片长度
	Char  []int  // 字符对应code切片
	KeyMap map[string]int //所有词对应等级
}

// 构建trie树使用的节点
type node struct {
	original rune //原始编码
	code  int // 字符对应code值
	depth int // 所处树的层级，正好对应子节点在key中索引
	left  int // 当前字符在key list中搜索的左边界索引 （包括）
	right int // 当前字符在key list中搜索的右边界索引（不包括）
	end   bool// 是否结束
}

type ResJson struct {
	Word string	`json:"word"`
	Level int	`json:"level"`
}

/**
	重置扩容base和check切片
 */
func (m *MapDoubleArrayTrie) resize(newSize int) int {
	base2  := make([]int, newSize, newSize)
	check2 := make([]int, newSize, newSize)
	if len(m.Base) > 0 {
		copy(base2, m.Base)
		copy(check2, m.Check)
	}
	m.Base  = base2
	m.Check = check2
	m.Size  = newSize
	return newSize
}

/**
	重置扩容char字符切片大小
 */
func (m *MapDoubleArrayTrie) resizeChar(Size int) {
	char2 := make([]int, Size, Size)
	if len(m.Char) > 0 {
		copy(char2, m.Char)
	}
	m.Char = char2
}

//查找某一节点下的子节点
func (m *MapDoubleArrayTrie) fetch (parent *node) []*node {
	//按照parent节点范围查找
	var pre rune
	children := make([]*node, 0)
	endStart := false
	for i:=parent.left;i<parent.right;i++ {
		if len(m.Keys[i]) <= parent.depth {
			continue
		}
		//如果字符前缀相同跳过
		if endStart {
			if len(m.Keys[i]) > (parent.depth + 1) {
				endStart = false
				//设置上一个字符节点right范围
				children[len(children)-1].left = i
				continue
			}
		} else if pre == m.Keys[i][parent.depth] {
			continue
		}
		pre = m.Keys[i][parent.depth]
		newNode := new(node)
		newNode.original = m.Keys[i][parent.depth]
		newNode.code     = m.Char[newNode.original]
		newNode.depth    = parent.depth + 1
		newNode.left     = i
		newNode.end      = len(m.Keys[i]) <= (parent.depth + 1)
		if newNode.end {
			endStart = true
		}
		if len(children) > 0 { //设置上一个字符节点right范围
			children[len(children)-1].right = i
		}
		children = append(children, newNode)
	}
	if len(children) > 0 { //如果有节点的情况下，设置最后一个节点的right
		children[len(children)-1].right = parent.right
	}
	return children
}

func (m *MapDoubleArrayTrie) insert(keyPre []rune, children []*node, index int) {
	pos := index + children[0].code - 1 //初始查找位置，根据父层索引加子节点最小code算出初始位置
	begin := 0
	childLen := len(children)
outer:
	for {
		pos++
		begin = pos - index - children[0].code
		if m.Base[pos] != 0 {
			continue
		}
		if s := begin + children[childLen - 1].code; s > m.Size {
			m.resize(int(float64(s) * 1.25))
		}

		for i := 0; i < childLen; i++ {
			//确保每一个子节点都能落到base和check中
			indC := index + begin + children[i].code
			if m.Check[indC] != 0 || m.Base[indC] != 0 {
				continue outer
			}
		}
		break
	}
	if m.Base[index] < 0 { //
		m.Base[index] = begin * 10 + (-m.Base[index])
	} else {
		m.Base[index] = begin
	}

	keyPre = append(keyPre, 1)
	//写入所有的子节点到base中
	//写入所有的子节点到check中
	//必须先把所有节点写入完之后，再去查找添加下一层节点
	for i := 0; i < childLen; i++ {
		ind := index + begin + children[i].code
		if children[i].end {
			keyPre[len(keyPre)-1] = children[i].original
			m.Base[ind] = -m.KeyMap[string(keyPre)]
			m.Check[ind] = -index
		} else {
			m.Base[ind] = 0
			m.Check[ind] = index
		}
	}
	//循环查找下一层节点并且插入dat结构中
	for i := 0; i < childLen; i++ {
		nodes := m.fetch(children[i])
		if len(nodes) > 0 {
			ind := index + begin + children[i].code
			keyPre[len(keyPre)-1] = children[i].original
			m.insert(keyPre, nodes, ind)
		}
	}
	return
}

/**
	格式化词库
 */
func (m *MapDoubleArrayTrie) format() error {
	if len(m.KeyMap) == 0 {
		return errors.New("dict is empty")
	}
	allKey := make([]string, 0, len(m.KeyMap))
	for k := range m.KeyMap {
		allKey = append(allKey, k)
	}
	sort.Strings(allKey)
	//字符串转化为字符code切片
	m.Char = make([]int, 65535, 65535)
	m.Keys = make([][]rune, 0, len(allKey))
	for _,key := range allKey {
		zs := []rune(key)
		m.Keys = append(m.Keys, zs)
		for _, z := range zs {
			m.Char[z] = -1
		}
	}
	c := 1
	for k,v := range m.Char{
		if v == -1 {
			m.Char[k] = c
			c++
		}
	}
	m.Char[0] = c
	return nil
}

/**
	查找搜索的词是否在词库
 */
func (m *MapDoubleArrayTrie) match (key string) (int,error) {
	keys, index, begin, level, charLen := []rune(key), 1, m.Base[1], 0, len(m.Char)

	for k,kv := range keys {
		if kv == 0 || m.Char[kv] == 0 || int(kv)>charLen {
			return level,errors.New("not found")
		}
		ind := index + begin + m.Char[kv]
		abs := m.Check[ind]
		if abs < 0 {
			abs = -abs
		}
		if abs != index { // 说明上一个字符和当前字符不是上下级关系
			return level,errors.New("not found key")
		}
		if k == len(keys) - 1 {//如果是最后一个字符
			if m.Check[ind] > 0 { //查找到最后一个字符，但是还没到单个词的结尾
				return level,errors.New("not found1")
			}
		} else { //如果不是最后一个字符
			if m.Base[ind] < 0 { //说明没有后续可查的值了，返回查询失败
				return level,errors.New("not found2")
			}
		}
		index = ind
		if m.Base[ind] > 0 && m.Check[ind] < 0 {
			begin = m.Base[ind]/10
			level = m.Base[ind]%10
		} else {
			begin = m.Base[ind]
			level = -m.Base[ind]
		}
	}
	return level,nil
}

/**
	内容匹配模式查找
	可传入一段文本，逐字查找是否在词库中存在
 */
func (m *MapDoubleArrayTrie) search (key string) [][3]int {
	Keys := []rune(key)
	var start,index,begin,level int
	var result [][3]int
	for k := range Keys {
		start = -1
		index = 1
		begin = m.Base[index]
		level = 0
		for i:=k;i<len(Keys);i++ {
			if int(Keys[i])>len(m.Char) || m.Char[Keys[i]] == 0 {//词库没有该字符重置状态继续查找
				break
			}
			ind := index + begin + m.Char[Keys[i]]
			abs := m.Check[ind]
			if abs < 0 {
				abs = -abs
			}
			if abs != index { // 说明上一个字符和当前字符不是上下级关系
				start = -1
				break
			}
			if start == -1 {
				start = i
			}
			if m.Check[ind] < 0 { //说明该词是结尾标记
				if m.Base[ind] > 0 && m.Check[ind] < 0 {
					level = m.Base[ind] % 10
				} else {
					level = -m.Base[ind]
				}
				result = append(result, [3]int{start, i, level})
			}
			if m.Base[ind] < 0 { //如果是结尾状态，没有后续词可查找
				break
			}
			index = ind
			if m.Base[ind] > 0 && m.Check[ind] < 0 {
				begin = m.Base[ind]/10
			} else {
				begin = m.Base[ind]
			}
		}
	}
	return result
}

// 以后扩展
//func (m *MapDoubleArrayTrie) add(key string) error {
//	err := m.match(key)
//	if err != nil {
//		return err
//	}
//	Keys := []rune(key)
//	for k,v := range Keys {
//		if int(v) > m.Char[0] {
//
//		}
//	}
//
//	return nil
//}

//func (m *MapDoubleArrayTrie) delete(key string) error {
//
//	return nil
//}

/**
	编译词库
 */
func (m *MapDoubleArrayTrie) build () error {
	if len(m.KeyMap) == 0 {
		return errors.New("empty Keys")
	}
	err := m.format()
	if err != nil {
		return err
	}
	m.resize(InitSize)
	root := new(node)
	root.left = 0
	root.right = len(m.Keys)
	root.depth = 0
	children := m.fetch(root)
	rootIndex := 1
	m.insert([]rune{}, children, rootIndex)
	fmt.Println(m.Base[0:50], len(m.Base), cap(m.Base))
	fmt.Println(m.Check[0:50])
	return nil
}

// 使用gob协议
func (m *MapDoubleArrayTrie) Store(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(m)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// 从指定路径加载DAT
func (m *MapDoubleArrayTrie) Load(path string) error {
	file, err := os.Open(path)
	if err != nil {
		//log.Fatalln(err)
		return err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(m)
	if err != nil {
		//log.Fatalln(err)
		return err
	}

	return nil
}

/**
	读取文件加载词库
 */
func (m *MapDoubleArrayTrie) readDict (filePath string) (bool,error) {
	f, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	md5h := md5.New()
	_, _ = io.Copy(md5h, f)
	md5Str := hex.EncodeToString(md5h.Sum(nil))
	if m.Fmd5 == md5Str {
		return true, nil
	} else {
		m.Fmd5 = md5Str
		m.Base = make([]int, 1)
		m.Check = make([]int, 1)
		m.Size = 0
	}

	m.KeyMap = make(map[string]int)
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
		in,_ := strconv.Atoi(string(line[0]))
		m.KeyMap[string(line[2:len(line)-2])] = in
	}
}

/**
	初始化dat
    加载store文件，读取词典，编译dat，保存store等
 */
func (m *MapDoubleArrayTrie) initHandle() {

	storeFile := "data/dat.data"
	dictFile  := "data/darts.txt"
	//storeFile := "dat.data"
	//dictFile  := "darts.txt"

	err := dat.Load(storeFile)
	if err != nil { //加载失败
		log.Println("load store", storeFile, "error:", err)
	} else {
		log.Println("load store", storeFile, "success")
	}

	status, err := dat.readDict(dictFile)

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

var dat = new(MapDoubleArrayTrie)

/**
	例子
 */
func example(){
	level, err := dat.match("番")
	fmt.Println(level, err)
	level, err = dat.match("中")
	fmt.Println(level, err)
	sk := "我要测试一下"
	contentRune := []rune(sk)
	result := dat.search(sk)
	fmt.Println(result)
	for i:=0;i<len(result);i++{
		fmt.Println("str:", string(contentRune[result[i][0]:result[i][1] + 1]), "level", result[i][2])
	}
}

/**
	入口函数
 */
func main(){
	dat.initHandle()
	//开始监听请求
	http.HandleFunc("/search", handle)
	_ = http.ListenAndServe(":8888", nil)
}

/**
	请求处理函数
 */
func handle(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	content := r.Form.Get("content")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if len(content) <= 0 {
		_, _ = fmt.Fprintf(w, "[\"err\":\"search content empty\"]")
		return
	}

	contentRune := []rune(content)
	data := dat.search(content)
	if len(data) > 0 {
		var result = make([]ResJson, 0, len(data))
		for i:=0;i<len(data);i++{
			result = append(result, ResJson{string(contentRune[data[i][0]:data[i][1] + 1]), data[i][2]})
		}
		byteData, _ := json.Marshal(result)
		_, _ = w.Write(byteData)
	} else {
		_, _ = fmt.Fprintf(w, "[\"err\":\"client\"]")
	}
}