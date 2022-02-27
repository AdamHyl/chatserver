package game

// todo 应该将数据存储在db中的 这里缓存内存中
var nameMap = map[string]bool{}         // 已经拥有的名字
var playerDataMap = map[string]string{} // 玩家账号与名字映射
