package game

import (
	"fmt"
	"strings"
	"time"

	"github.com/AdamHyl/chatserver/common/log"
	"github.com/AdamHyl/chatserver/internal/conf"
)

type Room struct {
	ID              int
	playerList      map[string]*Player
	history         []string               // 历史聊天数组
	historyMsg      string                 // 每次批量发送给玩家的历史信息
	worldStat       map[int]map[string]int // 每分钟单词统计,如果每秒会比较费 // todo 统计的数量相加没有判断最大值
	statMinute      int                    // 上次进行删除的分钟数
	beforeWorldStat map[string]int         // 当前分钟以前的单词统计数 每分钟刷新
}

func NewRoom(ID int) *Room {
	return &Room{
		ID:              ID,
		playerList:      make(map[string]*Player),
		history:         make([]string, 0),
		worldStat:       make(map[int]map[string]int),
		statMinute:      time.Now().Minute(),
		beforeWorldStat: make(map[string]int),
	}
}

func (r *Room) playerIn(p *Player) {
	if !p.login {
		return
	}

	if p.room != nil {
		p.room.playerOut(p)
	}
	p.room = r
	r.playerList[p.account] = p
	msg := fmt.Sprintf("player %v enter room %v", p.name, r.ID)
	log.Release(msg)

	for account, player := range r.playerList {
		if account == p.account {
			continue
		}

		player.sendData(msg)
	}
	p.sendData(fmt.Sprintf("enter room %v", r.ID))
	p.sendData(r.historyMsg)
}

func (r *Room) playerOut(p *Player) {
	if !p.login {
		return
	}
	p.room = nil
	delete(r.playerList, p.account)
	msg := fmt.Sprintf("player %v out room %v", p.name, r.ID)
	log.Release(msg)
	for account, player := range r.playerList {
		if p.account == account {
			continue
		}
		player.sendData(msg)
	}
}

func (r *Room) playerTalk(p *Player, content string) {
	if !p.login {
		return
	}
	if _, ok := r.playerList[p.account]; !ok {
		log.Error("player:%v not in room:%v", p.name, r.ID)
		return
	}

	r.worldStatistic(content)

	content = acMachine.MatchAndReplace(content, conf.ReplaceRune)

	log.Release("player:%v name:%v said:%v", p.account, p.name, content)
	sNow := fmt.Sprintf("%v ", time.Now().Format("2006-01-02 15:04:05"))
	msg := fmt.Sprintf("%v %v: %v \n", sNow, p.name, content)
	for _, player := range r.playerList {
		player.sendData(msg)
	}

	r.history = append(r.history, msg)
	if len(r.history) > conf.Server.HistoryNum {
		// 超过最大历史数据 清楚第一条
		r.history = r.history[1:]
	}
	r.updateHistoryMsg()
}

func (r *Room) updateHistoryMsg() {
	for _, msg := range r.history {
		r.historyMsg += msg
	}
}

func (r *Room) worldStatistic(content string) {
	minute := time.Now().Minute()

	stat := r.worldStat[minute]
	if stat == nil {
		stat = make(map[string]int)
		r.worldStat[minute] = stat
	}

	sList := strings.Split(content, " ")
	for _, s := range sList {
		n, ok := stat[s]
		if !ok {
			stat[s] = 1
		} else {
			stat[s] = n + 1
		}
	}
}

// 获取有效时间  比如当前是6分钟，检查10分钟 则返回
// map[0:true 1:true 2:true 3:true 4:true 5:true 6:true 57:true 58:true 59:true]
func getValidMinute(minute int, checkMinute int) map[int]bool {
	if checkMinute <= 0 || checkMinute >= 60 {
		return map[int]bool{minute: true}
	}
	valid := make(map[int]bool)
	if minute+1 >= checkMinute {
		start := minute + 1 - checkMinute
		for i := start; i <= minute; i++ {
			valid[i] = true
		}
	} else {

		start := minute + 1 + 60 - checkMinute
		end := 59
		for i := start; i <= end; i++ {
			valid[i] = true
		}

		start = 0
		end = minute
		for i := start; i <= end; i++ {
			valid[i] = true
		}
	}
	return valid
}

func (r *Room) update() {
	minute := time.Now().Minute()
	if minute == r.statMinute {
		return
	}

	r.statMinute = minute
	r.beforeWorldStat = make(map[string]int)
	valid := getValidMinute(minute, conf.Server.PopularCheckMinute)
	if len(valid) == 0 {
		return
	}

	delMinute := make([]int, 0)
	for m, stat := range r.worldStat {
		if _, ok := valid[m]; !ok {
			delMinute = append(delMinute, m)
			continue
		}

		for s, num := range stat {
			n, ok := r.beforeWorldStat[s]
			if !ok {
				r.beforeWorldStat[s] = 1
				continue
			}
			r.beforeWorldStat[s] = n + num
		}

	}

	for _, m := range delMinute {
		delete(r.worldStat, m)
	}
}

func (r *Room) mostPopular() string {
	nowMinute := time.Now().Minute()
	nowMinuteStat := r.worldStat[nowMinute]
	if nowMinuteStat == nil {
		nowMinuteStat = make(map[string]int)
	}

	mostNum := 0
	mostString := ""
	for s, num := range r.beforeWorldStat {
		nowNum := 0
		if n, ok := nowMinuteStat[s]; ok {
			nowNum = n
		}
		total := num + nowNum
		if mostNum < total {
			mostNum = total
			mostString = s
		}
	}
	return mostString
}
