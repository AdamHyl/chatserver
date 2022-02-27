package game

import (
	"fmt"
	"github.com/AdamHyl/chatserver/common/log"
	"github.com/AdamHyl/chatserver/internal/conf"
	"github.com/AdamHyl/chatserver/internal/game/protocol"
	"strconv"
	"sync/atomic"
	"time"
)

var handler = map[protocol.MsgID]func(*Player, string){}

func registerHandler() {
	handler[protocol.Login] = reqLogin
	handler[protocol.SetName] = reqSetName
	handler[protocol.ShowRoom] = reqShowRoom
	handler[protocol.JoinRoom] = reqJoinRoom
	handler[protocol.Gm] = reqGm
}

// 能保证在主worker执行
func handlerMsg(p *Player, data []byte) {
	if p == nil || len(data) <= 0 {
		return
	}

	atomic.StoreInt64(&p.activeTime, time.Now().Unix())

	//log.Release(buffer.String())

	s := string(data)
	log.Release("player:%v %v msg:%v", p.account, p.name, s)
	if len([]rune(s)) > conf.Server.MsgMaxLen {
		p.sendData(fmt.Sprintf("msg max len:%v", conf.Server.MsgMaxLen))
		return
	}

	isTalk := false // 非解析的协议 全部认定为聊天
	defer func() {
		if isTalk {
			reqChat(p, s)
		}
	}()

	// todo 暂时只一位ID 之后用空格分隔
	ID, err := strconv.Atoi(string(s[0]))
	if err != nil {
		isTalk = true
		return
	}
	msgID := protocol.MsgID(ID)
	msgData := ""
	if len(s) > 2 {
		// 第一位是空格
		msgData = s[2:]
	}

	if hand, ok := handler[msgID]; ok {
		hand(p, msgData)
	} else {
		isTalk = true
	}
}
