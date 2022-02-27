package game

import (
	"time"

	"github.com/AdamHyl/chatserver/common/acmachine"

	"github.com/AdamHyl/chatserver/common/log"
	"github.com/AdamHyl/chatserver/common/net"
	"github.com/AdamHyl/chatserver/common/worker"
	"github.com/AdamHyl/chatserver/internal/conf"
)

var (
	tcpServer        *net.TcpServer
	mainWorker       *worker.Worker
	accountOnlineMap = make(map[string]*Player) // 在线玩家 玩家账号作为key
	nameOnlineMap    = make(map[string]*Player) // 在线玩家 玩家名作为key
	roomMap          = make(map[int]*Room)
	gmMgr            *GmMgr
	acMachine        *acmachine.Machine // ac自动机
)

func OnInit() {
	if mainWorker != nil || tcpServer != nil {
		log.Fatal("mainWorker or tcpServer is not nil")
		return
	}

	mainWorker = worker.New(0, 0)
	if mainWorker == nil {
		log.Fatal("mainWorker init err")
		return
	}

	for i := 1; i <= conf.Server.RoomNum; i++ {
		roomMap[i] = NewRoom(i)
	}

	gmMgr = NewGmMgr()
	if gmMgr == nil {
		log.Fatal("gmMgr is nil")
		return
	}

	acMachine = acmachine.New(conf.DirtyList)
	if acMachine == nil {
		log.Fatal("acMachine init err")
		return
	}

	// 注册消息处理
	registerHandler()

	tcpServer = net.NewTcpServer(conf.Server.ClientTCPAddr, conf.Server.MaxConnNum, conf.PendingWriteNum)
	if tcpServer == nil {
		log.Fatal("cant new tcpServer")
		return
	}

	tcpServer.SetConnectionCb(func(conn *net.TcpConnection) {
		if conn == nil {
			return
		}
		p := NewPlayer(conn)
		if p == nil {
			return
		}

		conn.SetUserData(p)

		AddAgentPlayer(p)
	})

	tcpServer.SetCloseCb(func(conn *net.TcpConnection) {
		if conn == nil {
			return
		}

		p, _ := conn.GetUserData().(*Player)
		if p == nil {
			log.Error("GetUserData convert to *Player failed")
			return
		}

		mainWorker.Post("OnClose", p.Close)
	})

	tcpServer.SetMessageCb(func(conn *net.TcpConnection, msg []byte) {
		p, _ := conn.GetUserData().(*Player)
		if p == nil {
			log.Error("conn GetUserData convert to *player failed")
			return
		}

		// log.Debug("recv msg %v", buffer)

		mainWorker.Post("msg handler",
			func() {
				handlerMsg(p, msg)
			})
	})

	mainWorker.NewTicker("update", time.Second, update)

	tcpServer.Start()

	mainWorker.Run()
}

func update() {
	for _, room := range roomMap {
		room.update()
	}

	ForeachAgentPlayer(func(p *Player) {
		p.update()
	})

}
