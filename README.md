# sensitive
敏感词服务，仅支持utf8字符集

使用go语言开发，未使用第三方库或框架。

基于double array trie结构，扩展文本敏感词检索功能。

压缩base和check数组结构，减少内存使用量

增加敏感词等级，可设置敏感词等级为1-9

词典文件结构
等级 词
例如：
1 a
6 c

快速开始：

  docker pull cccjinxing/sensitive
  
  mkdir data
  
  echo "1 a" > darts.txt
  
  docker run -d --name sensitive -p 8888:8888 --mount "type=bind,src=data,dst=/sensitive/data" cccjinxing/sensitive
  
  curl http://localhost:8888/search?content=search%20content
  [
    {
        "word": "a",
        "level": 1
    }
  ]

