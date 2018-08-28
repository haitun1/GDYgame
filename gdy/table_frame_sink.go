package gdy

import (
	"mjserver/common"
	"mjserver/game/gdy/gdefine"
	"mjserver/game/gdy/gdylogic"
	"mjserver/message"
	"mjserver/utils"
	"sync"

	"github.com/gogo/protobuf/proto"
)

// TableFrameSink 桌子sink
type TableFrameSink struct {
	tableFrame common.ITableFrame
	log        *utils.Logger //日志句柄
	lock       *sync.RWMutex

	table *gdylogic.Table // 游戏桌子
	//	schedule *utils.Schedule //定時器
}

// Initializtion 创建桌子实例
func (t *TableFrameSink) Initializtion() {
	t.table = gdylogic.NewTable()
	table := NewTableLogic(t.table)
	t.table.SetStrategy(table)
	t.table.Init(t.tableFrame)
	//	t.schedule = utils.NewSchedule(t)
	//	go t.schedule.Start()
}

/*
// OnTimer 超时处理
func (t *TableFrameSink) OnTimer(id int, parameter interface{}) {
	t.log.Infoln("超时处理 OnTimer")
	switch t.table.GetGameStatus() {
	case gdy.StatusPlay: // 出牌超时
	}

}

// SurplusTime 剩余时间
func (t *TableFrameSink) SurplusTime(id int) int32 {
	if n := int32(int(t.schedule.Surplus(id).Seconds()) - gdy.TimeOffset); n > 0 {
		return n
	}

	return 0
}
*/
// RepositionSink ...
func (t *TableFrameSink) RepositionSink() {
}

// OnSetPrivateRoom 加载私有房配置
func (t *TableFrameSink) OnSetPrivateRoom() {
	// t.table.Init(t.tableFrame)
}

// Init .
func (t *TableFrameSink) Init() {

}

// Release 重置桌子
func (t *TableFrameSink) Release() {
	t.table.Release()
}

// OnEventGameStart 游戏开始
func (t *TableFrameSink) OnEventGameStart() bool {
	t.lock.Lock()
	defer t.lock.Unlock()

	//游戏开始
	return t.table.GameStart()
}

// OnEventGameConclude 游戏结算
func (t *TableFrameSink) OnEventGameConclude() bool {
	t.table.GameConclude(false)
	return true
}

// OnEventGameConcludeByTable 解散房间
func (t *TableFrameSink) OnEventGameConcludeByTable(calc bool) bool {
	t.table.GameConclude(calc)
	return true
}

// OnEventGameEnd 游戏总结束
func (t *TableFrameSink) OnEventGameEnd() bool {
	t.table.NotifyGameEnd()
	return true
}

// OnActionUserStandUp 玩家站起
func (t *TableFrameSink) OnActionUserStandUp(chairID int, userItem common.IServerUserItem) bool {
	return true
}

// OnActionUserSitDown 玩家坐下
func (t *TableFrameSink) OnActionUserSitDown(chairID int, userItem common.IServerUserItem, siteType common.SitdownType) bool {
	t.table.OnActionUserSitDown(chairID, userItem, siteType)
	return true
}

// OnGameMessage 游戏消息
func (t *TableFrameSink) OnGameMessage(subID uint16, data []byte, item common.IServerUserItem) bool {
	t.lock.Lock()
	defer t.lock.Unlock()
	switch subID {
	case gdy.SUB_C_OutPoker:
		return t.table.OnUserOutPoker(data, item)
	}
	return true
}

// OnFrameMessage 框架消息
func (t *TableFrameSink) OnFrameMessage(subID uint16, message []byte, userItem common.IServerUserItem) bool {
	return true
}

// OnOnlineEventChange 在线场景消息（断线重连？？？）
func (t *TableFrameSink) OnOnlineEventChange(item common.IServerUserItem, changeType common.OnlineStatusChangeType) {
	return
}

// OnNewTimer 定时器
func (t *TableFrameSink) OnNewTimer(id int, parameter interface{}) error {
	t.log.Infof("超时处理 OnTimer t.turnPokerCount:%d", t.table.GetTurnPokerCount())
	switch t.table.GetGameStatus() {
	case gdy.StatusPlay: // 出牌超时
		if t.table.GetTurnPokerCount() != 0 { // 管牌选择过
			data, _ := proto.Marshal(&message.GdyOutPoker{
				Type: proto.Int32(2),
			})
			item := t.tableFrame.GetTableUserItem(int(t.table.GetCurrentUser()))
			t.table.OnUserOutPoker(data, item)
			return nil
		}
		if len((*t.table.GetHandPoker())[t.table.GetCurrentUser()]) == 2 && (*t.table.GetHandPoker())[t.table.GetCurrentUser()][1] == 0x4e { // 剩下双王
			data, _ := proto.Marshal(&message.GdyOutPoker{
				Type:  proto.Int32(1),
				Poker: (*t.table.GetHandPoker())[t.table.GetCurrentUser()],
			})
			item := t.tableFrame.GetTableUserItem(int(t.table.GetCurrentUser()))
			t.table.OnUserOutPoker(data, item)
			return nil
		}

		data, _ := proto.Marshal(&message.GdyOutPoker{
			Type:  proto.Int32(1),
			Poker: (*t.table.GetHandPoker())[t.table.GetCurrentUser()][len((*t.table.GetHandPoker())[t.table.GetCurrentUser()])-1:],
		})

		item := t.tableFrame.GetTableUserItem(int(t.table.GetCurrentUser()))
		t.table.OnUserOutPoker(data, item)

	}
	return nil
}

//NewTableFrameSink 创建桌子框架
func NewTableFrameSink(tableFrame common.ITableFrame) common.ITableFrameSink {
	self := &TableFrameSink{
		tableFrame: tableFrame,
		log:        utils.NewLogger(),
		lock:       new(sync.RWMutex),
	}

	//桌子初始化
	self.Initializtion()

	return self
}
