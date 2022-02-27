package net

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/AdamHyl/chatserver/common/log"
)

type TcpServer struct {
	addr            string
	maxConnNum      int32
	ln              net.Listener
	conns           map[*TcpConnection]interface{}
	pendingWriteNum int

	wgLn   sync.WaitGroup
	wgConn sync.WaitGroup

	mutex sync.Mutex

	connCb    ConnectionCallback
	messageCb MessageCallback
	closeCb   CloseCallback
}

func NewTcpServer(addr string, maxConnNum int32, pendingWriteNum int) *TcpServer {
	return &TcpServer{
		addr:            addr,
		pendingWriteNum: pendingWriteNum,
		conns:           map[*TcpConnection]interface{}{},
		maxConnNum:      maxConnNum,
	}
}

func (tcpserver *TcpServer) listen() error {
	var ln net.Listener
	var err error
	ln, err = net.Listen("tcp4", tcpserver.addr)
	if err != nil {
		return err
	}

	tcpserver.mutex.Lock()
	tcpserver.ln = ln
	tcpserver.mutex.Unlock()
	log.Release("listen tcp[%v]", tcpserver.addr)

	return nil
}

func (tcpserver *TcpServer) init() error {
	if atomic.LoadInt32(&tcpserver.maxConnNum) <= 0 {
		atomic.StoreInt32(&tcpserver.maxConnNum, MaxConnNum)
	}

	if tcpserver.pendingWriteNum <= 0 {
		tcpserver.pendingWriteNum = DefaultPendingWriteNum
	}

	return tcpserver.listen()
}

func (tcpserver *TcpServer) addConnection(conn net.Conn) {
	maxConnNum := atomic.LoadInt32(&tcpserver.maxConnNum)

	tcpserver.mutex.Lock()
	defer tcpserver.mutex.Unlock()

	if len(tcpserver.conns) >= int(maxConnNum) {
		_ = conn.Close()
		log.Error("too many connections %v", maxConnNum)
		return
	}

	tcpConn := newTcpConnection(conn, tcpserver.pendingWriteNum)
	tcpConn.setConnectionCb(tcpserver.connCb)
	tcpConn.setMessageCb(tcpserver.messageCb)
	tcpConn.setCloseCb(func(connection *TcpConnection) {
		if tcpserver.closeCb != nil {
			tcpserver.closeCb(connection)
		}

		tcpserver.mutex.Lock()
		defer tcpserver.mutex.Unlock()
		delete(tcpserver.conns, connection)
	})
	tcpserver.conns[tcpConn] = nil
	go tcpConn.run(&tcpserver.wgConn)
}

func (tcpserver *TcpServer) doAccept() {
	for {
		conn, e := tcpserver.ln.Accept()
		if e != nil {
			log.Error("Accept error: %v; retrying", e.Error())
			continue
		}

		tcpserver.addConnection(conn)
	}
}

func (tcpserver *TcpServer) accept() {
	tcpserver.wgLn.Add(1)
	defer tcpserver.wgLn.Done()

	tcpserver.doAccept()
}

func (tcpserver *TcpServer) Start() {
	err := tcpserver.init()
	if err != nil {
		log.Fatal("tcpserver init fail:%v", err)
		return
	}

	go tcpserver.accept()
}

func (tcpserver *TcpServer) closeConnections() {
	tcpserver.mutex.Lock()
	defer tcpserver.mutex.Unlock()

	for conn := range tcpserver.conns {
		conn.Close()
	}
	tcpserver.conns = map[*TcpConnection]interface{}{}
}

func (tcpserver *TcpServer) stopAccept() {
	tcpserver.mutex.Lock()
	defer tcpserver.mutex.Unlock()

	if tcpserver.ln != nil {
		_ = tcpserver.ln.Close()
		tcpserver.ln = nil
	}
}

func (tcpserver *TcpServer) Stop() {
	tcpserver.stopAccept()
	tcpserver.wgLn.Wait()

	tcpserver.closeConnections()
	tcpserver.wgConn.Wait()
}

func (tcpserver *TcpServer) SetConnectionCb(cb ConnectionCallback) {
	tcpserver.connCb = cb
}

func (tcpserver *TcpServer) SetMessageCb(cb MessageCallback) {
	tcpserver.messageCb = cb
}

func (tcpserver *TcpServer) SetCloseCb(cb CloseCallback) {
	tcpserver.closeCb = cb
}
