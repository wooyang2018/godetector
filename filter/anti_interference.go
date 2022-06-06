package filter

func (tree *Trie) AntiFindAll(text string) []string {
	var (
		cur   = tree.Root
		next  *Node
		runes = []rune(text)
	)

	var ac = new(ac)
	for position := 0; position < len(runes); position++ {
		next = ac.next(cur, runes[position]) //遍历每一个Children以找到能够匹配runes[position]的Children
		//对上一行进行如下扩写
		/*
			while node.Children:
				if 匹配：
				next = node.Chidldren

		*/
		if next == nil {
			next = ac.fail(cur, runes[position])
		}

		cur = next
		ac.output(cur, runes, position)
	}

	return ac.results
}
