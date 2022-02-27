package game

import (
	"fmt"
	"strconv"
	"strings"
)

// Command ...
type Command interface {
	Usage() string
	GetPrefix() string
	Execute(para []string, p *Player) bool
}

// GmMgr ...
type GmMgr struct {
	CmdList map[string]Command
}

// Parse ...
func (mgr *GmMgr) Parse(p *Player, s string) bool {
	s = strings.Trim(s, string(' '))

	paras := strings.Split(s, string(' '))
	if len(paras) > 0 {
		if cmd, ok := mgr.CmdList[paras[0]]; ok {
			return cmd.Execute(paras, p)
		}
	}

	return false
}

// NewGmMgr ...
func NewGmMgr() *GmMgr {
	mgr := &GmMgr{
		CmdList: map[string]Command{},
	}

	var cmd Command

	cmd = &StatsCmd{}
	mgr.CmdList[cmd.GetPrefix()] = cmd

	cmd = &PopularCmd{}
	mgr.CmdList[cmd.GetPrefix()] = cmd

	return mgr
}

type StatsCmd struct {
}

// Usage ...
func (cmd *StatsCmd) Usage() string {
	return "查询玩家登陆时间、在线时长、房间号：/stats 玩家名"
}

// GetPrefix ...
func (cmd *StatsCmd) GetPrefix() string {
	return "/stats"
}

// Execute ...
func (cmd *StatsCmd) Execute(para []string, p *Player) bool {
	if len(para) < 2 {
		return false
	}

	name := para[1]
	targetPlayer, ok := nameOnlineMap[name]
	if !ok || targetPlayer == nil {
		p.sendData(fmt.Sprintf("%v not online", name))
		return true
	}

	p.sendData(targetPlayer.onlineStats())

	return true
}

type PopularCmd struct {
}

// Usage ...
func (cmd *PopularCmd) Usage() string {
	return "查询10分钟内发送频率最高的词：/popular 房间ID"
}

// GetPrefix ...
func (cmd *PopularCmd) GetPrefix() string {
	return "/popular"
}

// Execute ...
func (cmd *PopularCmd) Execute(para []string, p *Player) bool {
	if len(para) < 2 {
		return false
	}

	ID, err := strconv.Atoi(para[1])
	if err != nil {
		p.sendData(fmt.Sprintf("room ID:%v err", para[1]))
		return true
	}

	room, ok := roomMap[ID]
	if !ok || room == nil {
		p.sendData(fmt.Sprintf("no room, ID:%v", ID))
		return true
	}

	p.sendData(fmt.Sprintf("most popular world in 10 minutes is :%v", room.mostPopular()))

	return true
}
