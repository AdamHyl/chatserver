package conf

import (
	"encoding/json"
	"io/ioutil"

	"github.com/AdamHyl/chatserver/common/log"
)

type ServerConfig struct {
	LogLevel              string // 日志等级
	LogPath               string // 日志路径
	ClientTCPAddr         string // 监听地址
	MaxConnNum            int32  // 最大允许连接数
	PlayerInteractiveTime int32  // 最大不发送协议时间
	HistoryNum            int    // 历史聊天数量
	RoomNum               int    // 启动聊天房间数量
	MsgMaxLen             int    // 协议最大长度
	PopularCheckMinute    int    // 最经常使用词的统计时间 10分钟
}

var Server ServerConfig

func LoadServerConfig() {
	data, err := ioutil.ReadFile("conf/server.json")
	if err != nil {
		log.Fatal("%v", err)
	}

	var serverCfg ServerConfig
	err = json.Unmarshal(data, &serverCfg)
	if err != nil {
		log.Fatal("%v", err)
	} else {
		Server = serverCfg
	}
}

func init() {
	LoadServerConfig()
}
