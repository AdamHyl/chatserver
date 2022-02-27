package game

import (
	"fmt"
	"strconv"

	"github.com/AdamHyl/chatserver/common/log"
	"github.com/AdamHyl/chatserver/internal/conf"
)

func checkLoginAndName(p *Player, req, data string) bool {
	if p == nil {
		return false
	}

	if !p.login {
		log.Debug("player:%v has not login, cant req:%v %v", req, data)
		return false
	}

	if p.name == "" {
		log.Debug("player:%v has not name, cant req:%v %v", req, data)
		return false
	}

	return true
}

func reqLogin(p *Player, data string) {
	if p.login {
		return
	}

	oldPlayer, ok := accountOnlineMap[data]
	if ok {
		// 将之前的登录的踢下线
		oldPlayer.Close()
	}

	p.Login(data)
}

func reqSetName(p *Player, data string) {
	if !p.login {
		return
	}
	if p.name != "" {
		return
	}

	if _, ok := nameMap[data]; ok {
		//p.sendData(fmt.Sprintf("already has name %v", data))
		p.sendData("please enter a different name:")
		return
	}

	nameOnlineMap[data] = p
	p.name = data
	p.sendData("set name ok")
	nameMap[data] = true
	p.randomEnterRoom()
	playerDataMap[p.account] = data
}

func reqShowRoom(p *Player, data string) {
	if !checkLoginAndName(p, "reqShowRoom", data) {
		return
	}

	p.sendData(fmt.Sprintf("room list:%v - %v", 1, conf.Server.RoomNum))
}

func reqJoinRoom(p *Player, data string) {
	if !checkLoginAndName(p, "reqJoinRoom", data) {
		return
	}

	roomID, err := strconv.Atoi(data)
	if err != nil {
		p.sendData(fmt.Sprintf("wrong room ID,please choose:%v- %v", 1, conf.Server.RoomNum))
		return
	}

	room, ok := roomMap[roomID]
	if !ok || room == nil {
		p.sendData(fmt.Sprintf("wrong room ID,please choose:%v- %v", 1, conf.Server.RoomNum))
		return
	}

	room.playerIn(p)

}

func reqGm(p *Player, data string) {
	if !checkLoginAndName(p, "reqGm", data) {
		return
	}

	if len(data) == 0 {
		return
	}

	if gmMgr != nil {
		success := gmMgr.Parse(p, data)
		if !success {
			reqChat(p, data)
		}
	}
}

func reqChat(p *Player, data string) {
	if !checkLoginAndName(p, "reqChat", data) {
		p.sendData("have not login or set name")
		return
	}

	if p.room == nil {
		return
	}

	p.room.playerTalk(p, data)
}
