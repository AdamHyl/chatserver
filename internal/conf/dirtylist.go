package conf

import (
	"bufio"
	"io"
	"os"
)

var DirtyList []string

func load() {
	f, err := os.Open("conf/list.txt")
	if err != nil {
		return
	}
	defer f.Close()

	br := bufio.NewReader(f)
	for {
		s, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		DirtyList = append(DirtyList, string(s))
	}
	// log.Release("list:%v", DirtyList)
}

func init() {
	load()
}
