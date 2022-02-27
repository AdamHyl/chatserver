package game

import "sync"

var (
	playerSet     = make(map[string]*Player) //所有已连接玩家
	playerSetLock sync.Mutex
)

func AddAgentPlayer(p *Player) {
	if p == nil {
		return
	}

	playerSetLock.Lock()
	defer playerSetLock.Unlock()

	playerSet[p.fd] = p
}

func GetAgentPlayer(fd string) *Player {
	playerSetLock.Lock()
	defer playerSetLock.Unlock()

	return playerSet[fd]
}

func DelAgentPlayer(fd string) {
	playerSetLock.Lock()
	defer playerSetLock.Unlock()

	delete(playerSet, fd)
}

func GetAgentPlayerCount() int {
	playerSetLock.Lock()
	defer playerSetLock.Unlock()

	return len(playerSet)
}

func ForeachAgentPlayer(f func(*Player)) {
	if f == nil {
		return
	}

	ls := make([]*Player, 0, GetAgentPlayerCount())

	playerSetLock.Lock()
	for _, p := range playerSet {
		ls = append(ls, p)
	}
	playerSetLock.Unlock()

	for _, p := range ls {
		f(p)
	}
}
