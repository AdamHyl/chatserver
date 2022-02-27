package net

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/AdamHyl/chatserver/common/log"
)

const SetLingerSecs = 5

type TcpConnection struct {
	conn      net.Conn
	writeChan chan []byte
	mutex     sync.Mutex

	connCb    ConnectionCallback
	messageCb MessageCallback
	closeCb   CloseCallback

	destroyed bool

	userData interface{}
}

func newTcpConnection(conn net.Conn, pendingWriteNum int) *TcpConnection {
	return &TcpConnection{
		conn:      conn,
		writeChan: make(chan []byte, pendingWriteNum),
	}
}

func (tcpConn *TcpConnection) Conn() net.Conn {
	return tcpConn.conn
}

func (tcpConn *TcpConnection) releaseConnection() {
	_ = tcpConn.conn.(*net.TCPConn).SetLinger(SetLingerSecs)
	time.Sleep(ReleaseConnSleepDuration)
	_ = tcpConn.conn.Close()

	log.Release("releaseConnection")
}

func (tcpConn *TcpConnection) writeTask(wg *sync.WaitGroup) {
	if wg != nil {
		wg.Add(1)
		defer wg.Done()
	}

	for b := range tcpConn.writeChan {
		_, err := tcpConn.conn.Write(b)
		if err != nil {
			log.Error("write error %v", err.Error())
			continue
		}
	}

	tcpConn.releaseConnection()
}

func (tcpConn *TcpConnection) setConnectionCb(cb ConnectionCallback) {
	tcpConn.connCb = cb
}

func (tcpConn *TcpConnection) setMessageCb(cb MessageCallback) {
	tcpConn.messageCb = cb
}

func (tcpConn *TcpConnection) setCloseCb(cb CloseCallback) {
	tcpConn.closeCb = cb
}

func (tcpConn *TcpConnection) SendData(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	tcpConn.mutex.Lock()
	defer tcpConn.mutex.Unlock()

	if tcpConn.destroyed {
		return errors.New("tcpConn is destroyed")
	}

	select {
	case tcpConn.writeChan <- data:

	default:
		return errors.New(fmt.Sprintf("writeChan is full %v", cap(tcpConn.writeChan)))
	}

	return nil
}

func (tcpConn *TcpConnection) Close() {
	tcpConn.mutex.Lock()
	defer tcpConn.mutex.Unlock()

	if tcpConn.destroyed {
		return
	}

	close(tcpConn.writeChan)
	tcpConn.destroyed = true
}

func (tcpConn *TcpConnection) run(wg *sync.WaitGroup) {
	if wg != nil {
		wg.Add(1)
		defer wg.Done()
	}

	go tcpConn.writeTask(wg)

	if tcpConn.connCb != nil {
		tcpConn.connCb(tcpConn)
	}

	onceBuffer := make([]byte, 4096)

	for {
		n, err := tcpConn.conn.Read(onceBuffer)
		if err != nil {
			log.Error("read error %v", err.Error())
			break
		}

		if tcpConn.messageCb != nil {
			// 去除一个末尾换行符
			tcpConn.messageCb(tcpConn, onceBuffer[:n-1])
		}
	}

	tcpConn.Close()
	if tcpConn.closeCb != nil {
		tcpConn.closeCb(tcpConn)
	}
}

func (tcpConn *TcpConnection) RemoteAddr() string {
	return tcpConn.conn.RemoteAddr().String()
}

func (tcpConn *TcpConnection) GetUserData() interface{} {
	return tcpConn.userData
}

func (tcpConn *TcpConnection) SetUserData(userData interface{}) {
	tcpConn.userData = userData
}
