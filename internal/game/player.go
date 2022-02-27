package game

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/AdamHyl/chatserver/common/log"
	"github.com/AdamHyl/chatserver/common/net"
	"github.com/AdamHyl/chatserver/common/utility"
	"github.com/AdamHyl/chatserver/internal/conf"
)

type Player struct {
	conn       *net.TcpConnection //玩家连接
	fd         string
	activeTime int64 //活跃时间
	destroy    bool  //已销毁标志
	//encryptKey aes.Key
	login bool

	account   string
	name      string
	loginTime time.Time

	room *Room
}

func NewPlayer(conn *net.TcpConnection) *Player {
	if conn == nil {
		log.Error("conn is nil")
		return nil
	}

	p := &Player{
		conn: conn,
	}

	atomic.StoreInt64(&p.activeTime, time.Now().Unix())

	p.fd = utility.GetUID()

	log.Release("player[%v][%v] connect.", p.Addr(), p.fd)
	return p
}

func (p *Player) Login(account string) {
	p.account = account
	p.login = true
	p.loginTime = time.Now()

	name, ok := playerDataMap[p.account]
	if ok {
		p.name = name
		p.sendData(fmt.Sprintf("login success"))
		p.randomEnterRoom()
		nameOnlineMap[name] = p
	} else {
		p.sendData("please set your name:")
	}
	log.Release("%v login success", p)

	accountOnlineMap[account] = p
}

func (p *Player) Close() {
	if p.destroy {
		return
	}

	//player唯一销毁处
	p.Destroy()
}

func (p *Player) Addr() string {
	return p.conn.RemoteAddr()
}

// Destroy ...
func (p *Player) Destroy() {
	if p.room != nil {
		p.room.playerOut(p)
	}
	p.sendData("conn close")

	p.conn.Close()

	//清理用户集
	p.destroy = true

	DelAgentPlayer(p.fd)

	p.login = false

	if p.account != "" {
		delete(accountOnlineMap, p.account)
	}
	if p.name != "" {
		delete(nameOnlineMap, p.name)
	}

	log.Release("player[%v][%v] destroy", p.conn.RemoteAddr(), p.fd)
}

func (p *Player) update() {
	//不活跃踢线
	nowUnix := time.Now().Unix()

	if nowUnix-atomic.LoadInt64(&p.activeTime) > int64(conf.Server.PlayerInteractiveTime) {
		p.sendData("too long not send msg")
		p.Close()
		log.Release("player[%v][%v] not active, force kick", p.Addr(), p.fd)
		return
	}
}

func (p *Player) sendData(msg string) {
	if p.conn == nil {
		log.Error("conn is nil")
		return
	}
	l := len(msg)
	if l == 0 {
		return
	}

	// 让客户端输出换行
	if msg[l-1:] != "\n" {
		msg = msg + "\n"
	}
	e := p.conn.SendData([]byte(msg))
	if e != nil {
		log.Error("sendData err  %v", e)
	}
}

func (p *Player) randomEnterRoom() {
	if p.room != nil {
		return
	}

	for _, room := range roomMap {
		room.playerIn(p)
		break
	}
}

func (p *Player) onlineStats() string {
	roomID := 0
	if p.room != nil {
		roomID = p.room.ID
	}

	sLoginTime := fmt.Sprintf("%v ", p.loginTime.Format("2006-01-02 15:04:05"))
	return fmt.Sprintf("player:%v login time:%v, online time:%v second. room num:%v",
		p.name, sLoginTime, time.Now().Sub(p.loginTime).Seconds(), roomID)
}
