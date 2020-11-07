/**
	扩展double array 方法
	包括多种查找匹配方法
	增加和删除等
 */

package main

import (
	"errors"
)

/**
查找搜索的词是否在词库
参数 key string 查找的词
返回 最后一个字符的索引，偏移量，等级，error
*/
func (d *DoubleArray) match (key string, forceBack bool) (int,int,int,error) {
	keys, index, begin, level := []rune(key), 1, d.Base[1], 0

	for k,kv := range keys {
		if kv == 0 {
			return index, begin, level, errors.New("code error")
		}
		ind := begin + int(kv)
		abs := d.Check[ind]
		if abs < 0 {
			abs = -abs
		}
		if abs != index { // 说明上一个字符和当前字符不是上下级关系
			return index, begin, level, errors.New("not found key")
		}
		if k == len(keys) - 1 {//如果是最后一个字符
			if forceBack { //强制返回模式，一定返回查找到的结果，除非没有结果
				return index, begin, level, nil
			}
			if d.Check[ind] > 0 { //查找到最后一个字符，但是还没到单个词的结尾
				return index, begin, level, errors.New("not found1")
			}
		} else {                 //如果不是最后一个字符
			if d.Base[ind] < 0 { //说明没有后续可查的值了，返回查询失败
				return index, begin, level, errors.New("not found2")
			}
			index = ind
			if d.Base[ind] > 0 && d.Check[ind] < 0 {
				begin = d.Base[ind]/10
				level = d.Base[ind]%10
			} else {
				begin = d.Base[ind]
				level = -d.Base[ind]
			}
		}
	}
	return index, begin, level, nil
}

/**
内容匹配模式查找
可传入一段文本，逐字查找是否在词库中存在
*/
func (d *DoubleArray) search (key string) [][3]int {
	Keys := []rune(key)
	var start,index,begin,level int
	var result [][3]int
	for k := range Keys {
		start = -1
		index = 1
		begin = d.Base[index]
		level = 0
		for i:=k;i<len(Keys);i++ {
			//词库没有该字符重置状态继续查找
			//if int(Keys[i])>len(d.Char) || d.Char[Keys[i]] == 0 {
			//	break
			//}
			ind := begin + int(Keys[i])
			if ind > len(d.Base) { //越界base数组，结束查找
				break
			}
			abs := d.Check[ind]
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
			if d.Check[ind] < 0 { //说明该词是结尾标记
				if d.Base[ind] > 0 && d.Check[ind] < 0 {
					level = d.Base[ind] % 10
				} else {
					level = -d.Base[ind]
				}
				result = append(result, [3]int{start, i, level})
			}
			if d.Base[ind] < 0 { //如果是结尾状态，没有后续词可查找
				break
			}
			index = ind
			if d.Base[ind] > 0 && d.Check[ind] < 0 {
				begin = d.Base[ind]/10
			} else {
				begin = d.Base[ind]
			}
		}
	}
	return result
}

/**
	前缀相同字符，添加剩余不同字符
 */
func (d *DoubleArray) _add(index int, keys []rune, level int) error {
	offset := 0
	preLevel := -d.Base[index]
	for k,v := range keys {
		pos := int(v)
		for {
			pos++
			if pos > d.Size { //每次循环计算最大字符code位置是否超出范围
				d.resize(int(float64(pos) * 1.25))
			}
			if d.Base[pos] != 0 || d.Check[pos] != 0{
				continue
			}
			offset = pos - int(v)
			if k == 0 {
				d.Base[index]  = offset * 10 + preLevel
			} else if k == len(keys) - 1 {//如果是最后一个字符
				d.Base[index] = offset
				d.Base[pos]   = -level
				d.Check[pos]  = -index
			} else {
				d.Base[index] = offset
				d.Base[pos]   = 0
				d.Check[pos]  = index
			}
		}
	}
	return nil
}

/**
	重置树结构，保持最优结构状态
	移动树结构并添加词
	todo 待优化
 */
func (d *DoubleArray) _addMove(keys string, level int) error {

	d.KeyMap[keys] = level

	err := d.build()
	if err != nil {
		return errors.New("add key error" + err.Error())
	}

	err = dat.Store(storeFile)
	if err != nil {
		return errors.New("add key success, but store DAT is error" + err.Error())
	}

	return nil
}

/**
	动态添加数据
	复杂度：可能是O(1)也可能是O(root)
 */
func (d *DoubleArray) add(key string, level int) error {
	//keys := []rune(key)
	//先查找相同前缀的节点
	//获取相同前缀最后的base status，开始添加数据
	//读取已经入库相同前缀的词
	keys, index, begin := []rune(key), 1, d.Base[1]
	//先查找最长的相同前缀index
	for k,kv := range keys {
		if kv == 0 {
			return errors.New("code error")
		}
		ind := begin + int(kv)
		abs := d.Check[ind]
		if abs < 0 {
			abs = -abs
		}
		if abs != index { // 说明上一个字符和当前字符不是上下级关系
			return d._addMove(key, level)
		}
		if k == len(keys) - 1 {//如果是最后一个字符
			if d.Check[ind] > 0 { //查找到最后一个字符，但是还没到单个词的结尾
				d.Check[ind] = -d.Check[ind]
			}
			return nil
		} else {
			if d.Base[ind] < 0 { //说明没有后续可查的值了，
				//todo list 查找一个合适的位置放置词
				return d._add(index, keys[k:], level)
			}
			index = ind
			if d.Base[ind] > 0 && d.Check[ind] < 0 {
				begin = d.Base[ind]/10
				level = d.Base[ind]%10
			} else {
				begin = d.Base[ind]
				level = -d.Base[ind]
			}
		}
	}
	return nil
}

/**
前缀查找，递归方法
*/
func (d *DoubleArray) _prefix (rCode []rune, index int, offset int) [][]rune {
	result := make([][]rune, 0, 1)
	negative := 0
	zRune := rune(0)
	negative = -index
	for i:=2; i<len(d.Base); i++{
		zRune = rune(i-offset)
		if d.Check[i] == negative || d.Check[i] == index{
			result = append(result, append(rCode, zRune))
		}
		if d.Base[i] > 0 {
			nextRune := d._prefix(append(rCode, zRune), i, d.Base[i])
			if len(nextRune) > 0 {
				for _,v := range nextRune {
					result = append(result, v)
				}
			}
		}
	}
	return result
}
/**
前缀查找
*/
func (d *DoubleArray) prefix(pre string, limit int) ([][]rune, error) {
	index, begin, _, err := d.match(pre, true)
	result := make([][]rune, 0, limit)
	if err != nil {
		return result, err
	}
	pres := []rune(pre)
	if d.Check[index] < 0 { //说明搜索词是结束词
		result = append(result, pres)
	}
	result = append(result, d._prefix(make([]rune,0,1), index, begin)...)
	return result[0:limit], nil
}

/**
删除词
参数 key string 需要删除的词
*/
func (d *DoubleArray) remove(key string) error {

	index, begin, _, err := d.match(key, false)
	if err != nil {
		return err
	}
	if begin >= 0 { //该词还有子节点
		if d.Check[index] < 0 { //说明是可结束状态
			d.Check[index] = -d.Check[index]
		}
	} else {//没有子节点，直接清空数据
		d.Base[index]  = 0
		d.Check[index] = 0
	}
	delete(d.KeyMap, key)
	_ = DictRemove(key)

	return d.format()
}