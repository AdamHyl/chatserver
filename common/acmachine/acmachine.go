package acmachine

// AC自动机

type Node struct {
	r        rune           // 当前节点字符
	endChar  bool           // 是否是匹配串的最后一个字符
	charLen  int            // 匹配串的长度
	children map[rune]*Node // 后续子节点
	fail     *Node          // 失败指针
}

type Machine struct {
	root *Node
}

func New(sList []string) *Machine {
	m := &Machine{root: &Node{
		children: map[rune]*Node{},
	}}

	m.insert(sList)
	m.setFailPointer()

	return m
}

func (m *Machine) insert(sList []string) {
	for _, s := range sList {
		runeS := []rune(s)
		p := m.root
		for _, r := range runeS {
			if p.children == nil {
				p.children = make(map[rune]*Node)
			}
			if p.children[r] == nil {
				p.children[r] = &Node{r: r}
			}
			p = p.children[r]
		}
		p.endChar = true
		p.charLen = len(runeS)
	}
}

func (m *Machine) setFailPointer() {
	q := make([]*Node, 0)
	q = append(q, m.root)
	for len(q) != 0 {
		p := q[0]
		q = q[1:]

		for r, child := range p.children {
			if p == m.root {
				child.fail = m.root
			} else {
				fail := p.fail
				for fail != nil {
					if _, ok := fail.children[r]; ok {
						child.fail = fail.children[r]
						break
					}
					fail = fail.fail
				}

				if fail == nil {
					child.fail = m.root
				}
			}
			q = append(q, child)
		}
	}
}

func (m *Machine) MatchAndReplace(content string, replace rune) string {
	runeC := []rune(content)
	p := m.root
	res := make([]string, 0)
	result := []rune(content)
	for i, r := range runeC {
		_, ok := p.children[r]
		for !ok && p != m.root {
			p = p.fail
		}

		if _, ok = p.children[r]; ok {
			p = p.children[r]
			if p.endChar {
				res = append(res, string(runeC[i-p.charLen+1:i]))
				for k := i - p.charLen + 1; k <= i; k++ {
					result[k] = replace
				}
			}
		}
	}

	return string(result)
}
