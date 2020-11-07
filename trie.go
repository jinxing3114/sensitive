package main

// 构建trie树使用的节点
type Node struct {
	Code  int // 字符对应code值
	Depth int  // 所处树的层级，正好对应子节点在key中索引
	Left  int  // 当前字符在key list中搜索的左边界索引 （包括）
	Right int  // 当前字符在key list中搜索的右边界索引（不包括）
	End   bool // 是否结束
}

//根据父节点查找 double array 下的子节点
func (n *Node) fetch (dat *DoubleArray) []*Node {
	//按照parent节点范围查找
	var pre rune
	children := make([]*Node, 0)
	endStart := false
	for i:=n.Left;i<n.Right;i++ {
		if len(dat.Keys[i]) <= n.Depth {
			continue
		}
		//如果字符前缀相同跳过
		if endStart {
			if len(dat.Keys[i]) > (n.Depth + 1) {
				endStart = false
				//设置上一个字符节点right范围
				children[len(children)-1].Left = i
				continue
			}
		} else if pre == dat.Keys[i][n.Depth] {
			continue
		}
		pre = dat.Keys[i][n.Depth]
		newNode := new(Node)
		newNode.Code     = int(dat.Keys[i][n.Depth])
		newNode.Depth    = n.Depth + 1
		newNode.Left     = i
		newNode.End      = len(dat.Keys[i]) <= (n.Depth + 1)
		if newNode.End {
			endStart = true
		}
		if len(children) > 0 { //设置上一个字符节点right范围
			children[len(children)-1].Right = i
		}
		children = append(children, newNode)
	}
	if len(children) > 0 { //如果有节点的情况下，设置最后一个节点的right
		children[len(children)-1].Right = n.Right
	}
	return children
}