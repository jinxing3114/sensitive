package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

/**
	基础返回值格式
 */
type baseResult struct {
	Code int8   `json:"code"`
	Msg  string `json:"msg"`
	Data interface{} `json:"data"`
}

/**
	http服务返回json结构
*/
type matchSearchJson struct {
	Word string	`json:"word"`
	Level int	`json:"level"`
}

/**
	查找匹配
*/
func search(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	content := r.Form.Get("content")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	result := baseResult{Code : 1, Msg : "ok"}
	if len(content) > 0 {
		contentRune := []rune(content)
		data := XT.Search(content)
		if len(data) > 0 {
			var resultData = make([]matchSearchJson, 0, len(data))
			for i:=0;i<len(data);i++{
				resultData = append(resultData, matchSearchJson{string(contentRune[data[i][0]:data[i][1] + 1]), data[i][2]})
			}
			result.Data = resultData
		} else {
			result.Code = 1
			result.Msg  = "result is empty"
		}
	} else {
		result.Code = 0
		result.Msg  = "search content is empty"
	}
	byteData, _ := json.Marshal(result)
	_, _ = w.Write(byteData)
}

/**
	前缀匹配请求处理
 */
func prefix(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	key := r.Form.Get("key")
	limit, _ := strconv.Atoi(r.Form.Get("limit"))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	result := baseResult{Code : 1, Msg : "ok"}
	if len(key) > 0 {
		data, err := XT.Prefix(key, limit)
		if err != nil {
			var resultData = make([]matchSearchJson, 0, len(data))
			for _,da := range data {
				resultData = append(resultData, matchSearchJson{Word: string(da)})
			}
			for i:=0;i<len(data);i++{
			}
			result.Data = resultData
			result.Code = 0
			result.Msg  = err.Error()
		}

	} else {
		result.Code = 0
		result.Msg  = "match key is empty"
	}
	byteData, _ := json.Marshal(result)
	_, _ = w.Write(byteData)

}

/**
	后缀匹配请求处理
 */
func suffix(w http.ResponseWriter, r *http.Request) {

}

/**
	模糊匹配请求处理
 */
func fuzzy(w http.ResponseWriter, r *http.Request) {

}

/**
	完全匹配
 */
func perfect(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	key := r.Form.Get("key")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	result := baseResult{Code : 1, Msg : "ok"}
	if len(key) > 0 {
		_, _, err := XT.Match(key, false)
		if err != nil {
			result.Code = 0
			result.Msg  = err.Error()
		}
	} else {
		result.Code = 0
		result.Msg  = "match key is empty"
	}
	byteData, _ := json.Marshal(result)
	_, _ = w.Write(byteData)
}

/**
	移除词
 */
func remove(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	key := r.Form.Get("key")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	result := baseResult{Code : 1, Msg : "ok"}
	if len(key) <= 0 {
		result.Code = 0
		result.Msg  = "key is empty"
	} else {
		err := XT.Remove(key)
		if err != nil {
			result.Code = 0
			result.Msg  = err.Error()
		}
	}

	byteData, _ := json.Marshal(result)
	_, _ = w.Write(byteData)
}

/**
	添加词
 */
func add(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	key := r.Form.Get("key")
	level, _ := strconv.Atoi(r.Form.Get("level"))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	result := baseResult{Code : 1, Msg : "ok"}
	if len(key) <= 0 {
		result.Code = 0
		result.Msg  = "key is empty"
	} else if level > 9 || level < 1 {
		result.Code = 0
		result.Msg  = "level not 1~9"
	} else {
		err := XT.Insert(key, level)
		if err != nil {
			result.Code = 0
			result.Msg  = err.Error()
		}
	}
	byteData, _ := json.Marshal(result)
	_, _ = w.Write(byteData)
}

/**
	启动http服务,绑定请求方法等

	@param addr string 监听地址和端口号。例如127.0.0.1:8888
	@return nil
 */
func startServe(addr string){
	//删除词
	http.HandleFunc("/delete", remove)

	//添加词
	http.HandleFunc("/add", add)

	//查找匹配-不会词库中的词匹配
	http.HandleFunc("/match/search", search)

	//前缀匹配
	http.HandleFunc("/match/prefix", prefix)

	//后缀匹配
	http.HandleFunc("/match/suffix", suffix)

	//模糊匹配-会拆分词库中的词进行匹配
	http.HandleFunc("/match/fuzzy", fuzzy)

	//完全匹配-
	http.HandleFunc("/match/perfect", perfect)

	//启动监听http地址和端口
	_ = http.ListenAndServe(addr, nil)
}

