package protocol

type MsgID int32

const (
	Start    MsgID = 0
	Login    MsgID = 1 // 登录 发送格式：1 account
	SetName  MsgID = 2 // 设置名字 发送格式：2 Name
	ShowRoom MsgID = 3 // 显示房间列表 格式：3
	JoinRoom MsgID = 4 // 加入房间 格式：4 ID
	Gm       MsgID = 5 //
	End      MsgID = 6
)
