package net

import (
	"time"
)

const (
	MaxConnNum               = 100 // tcp连接数
	DefaultPendingWriteNum   = 100 // 写队列长度
	ReleaseConnSleepDuration = 3 * time.Second
)

type ConnectionCallback func(*TcpConnection)
type MessageCallback func(*TcpConnection, []byte)
type CloseCallback func(*TcpConnection)
