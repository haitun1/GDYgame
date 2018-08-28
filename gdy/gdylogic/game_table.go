package gdylogic

import (
	"reflect"
	"sort"
	"time"

	"github.com/gogo/protobuf/proto"

	"mjserver/common"
	"mjserver/define"
	"mjserver/game/gdy/gdefine"
	"mjserver/message"
	"mjserver/utils"
)

// Strategy 抽象私人接口
type Strategy interface {
	//设置牌和手牌个数
	SetLogicCfg()
	//设置牌数据
	SetPokerData()
	//确定首局庄家
	EnsureFirstBanker()
	//确定游戏结束庄家
	EnsureGameOverBanker()
	//检查首轮出牌
	CheckFirstRound(poker []int32, chairID int) gdy.UserOutPokerCode
	//检查特殊牌
	CheckSpecialPoker(poker []int32, pokerType int32, chairID int) gdy.UserOutPokerCode
	//检查过牌
	CheckPassCard(chairID int, outPokerResult OutPokerResult) gdy.UserOutPokerCode
	//检查出牌
	SearchOutCard(chairID int, outPokerResult *OutPokerResult) bool

	//计算游戏分数
	CalcGameScore()
	//检查出牌
	CheckOutCard(chairID int, outPoker []int32, outType int32) gdy.UserOutPokerCode
	//特殊牌型
	CheckSpecial() bool
	// 发牌
	NotifyDeal()
}

// Table .
type Table struct {
	tableFrame common.ITableFrame
	gameLogic  *GameLogic    //游戏逻辑
	log        *utils.Logger //日志句柄
	roomID     int32
	strategy   Strategy

	//配置
	gameRule       int64   // 游戏规则
	roomType       int32   // 房间类型
	playerCount    int32   // 玩家数
	pokerCount     int32   // 牌个数
	pokerData      []int32 // 牌库
	handPokerCount int32   // 初始手牌

	bankerUser              int32     // 庄家
	currentUser             int32     // 当前操作玩家
	currentUserMagicCardNum int32     // 当前操作玩家出牌赖子数量
	handPoker               [][]int32 // 玩家手牌
	handCount               []int32   // 玩家手牌数量
	outCard                 []bool    // 出过牌true
	spring                  []int32   // 春天
	passivityNoCard         []int32   // 被关
	SpringChairID           int32     // 春天玩家

	turnPokerType    int32   // 上个操作玩家出牌类型
	turnPokerCount   int32   // 上个操作玩家出牌数量
	turnPoker        []int32 // 上个操作玩家出牌牌数据
	turnRealPoker    []int32 // 上个操作玩家真实出牌牌数据
	turnMagicCardNum int32   //  上个操作玩家出牌赖子数量
	turnShamCard     []int32 //  上个操作玩家出牌赖子顶替的牌
	turnMinCard      int32   //  上个操作玩家出牌最小值
	turnWiner        int32
	turnChairID      int32 // 上个操作玩家ID
	laiziPoker       int32 // 赖子数据
	leftCardCount    int32 // 剩余牌数量
	bombNum          int32 // 炸弹数量

	currentStatus gdy.GameStatusType //游戏状态

	//结算
	gameScore []int64
	bombScore []int64
	bombCount []int32 //炸弹数

	//总结算
	maxWinScore    []int64 //单局最高得分
	bombTotalCount []int32 //炸弹数
	winCount       []int32 //赢次数
	lostCount      []int32 //数次数
	totalScore     []int64 //总分
}

// SetStrategy 私有接口
func (t *Table) SetStrategy(strategy Strategy) {
	t.strategy = strategy
}

// Init 创建房间初始化游戏数据
func (t *Table) Init(tableFrame common.ITableFrame) {
	t.tableFrame = tableFrame
	t.gameRule = t.tableFrame.GetGameConfig().GameRule
	t.roomType = int32(t.tableFrame.GetGameConfig().RoomType)
	t.playerCount = int32(t.tableFrame.GetChairCount())
	//t.log.Infof("[玩家人数]：%d", t.playerCount)
	t.roomID = t.tableFrame.GetRoomId()

	//设置游戏逻辑的一些配置信息
	t.strategy.SetLogicCfg()

	t.pokerData = make([]int32, t.pokerCount)

	t.currentStatus = gdy.StatusFree
	t.bankerUser = define.INVALID_CHAIR_ID
	t.currentUser = define.INVALID_CHAIR_ID
	t.turnWiner = define.INVALID_CHAIR_ID
	t.turnChairID = define.INVALID_CHAIR_ID
	t.SpringChairID = define.INVALID_CHAIR_ID

	t.handPoker = make([][]int32, t.playerCount)
	/*
		for i := int32(0); i < t.playerCount; i++ {
			t.handPoker[i] = make([]int32, t.handPokerCount)
		}
	*/
	t.handCount = make([]int32, t.playerCount)
	t.turnPoker = make([]int32, t.handPokerCount)
	t.gameScore = make([]int64, t.playerCount)
	t.bombScore = make([]int64, t.playerCount)
	t.bombCount = make([]int32, t.playerCount)

	t.maxWinScore = make([]int64, t.playerCount)
	t.bombTotalCount = make([]int32, t.playerCount)
	t.winCount = make([]int32, t.playerCount)
	t.lostCount = make([]int32, t.playerCount)
	t.totalScore = make([]int64, t.playerCount)
	t.outCard = make([]bool, t.playerCount)
	t.spring = make([]int32, t.playerCount)
	t.passivityNoCard = make([]int32, t.playerCount)
	t.bombNum = 0
	t.turnPokerType = 0
	t.turnMinCard = 0
	t.turnPokerCount = 0
	t.turnMagicCardNum = 0
	t.currentUserMagicCardNum = 0
	t.laiziPoker = 0

}

// OnTimer 超时处理
func (t *Table) OnTimer(id int, parameter interface{}) {
	t.log.Infof("超时处理 OnTimer t.turnPokerCount:%d", t.turnPokerCount)
	switch t.currentStatus {
	case gdy.StatusPlay: // 出牌超时
		if t.turnPokerCount != 0 { // 管牌选择过
			data, _ := proto.Marshal(&message.GdyOutPoker{
				Type: proto.Int32(2),
			})
			item := t.tableFrame.GetTableUserItem(int(t.currentUser))
			t.OnUserOutPoker(data, item)
			return
		}
		if len(t.handPoker[t.currentUser]) == 2 && t.handPoker[t.currentUser][1] == 0x4e { // 剩下双王
			data, _ := proto.Marshal(&message.GdyOutPoker{
				Type:  proto.Int32(1),
				Poker: t.handPoker[t.currentUser],
			})
			item := t.tableFrame.GetTableUserItem(int(t.currentUser))
			t.OnUserOutPoker(data, item)
			return
		}

		data, _ := proto.Marshal(&message.GdyOutPoker{
			Type:  proto.Int32(1),
			Poker: t.handPoker[t.currentUser][len(t.handPoker[t.currentUser])-1:],
		})

		item := t.tableFrame.GetTableUserItem(int(t.currentUser))
		t.OnUserOutPoker(data, item)

	}

}

// SurplusTime 剩余时间
func (t *Table) SurplusTime(id int) int32 {
	if int64(t.roomType)&gdy.PyOperateInfinite != 0 {
		return 15
	}

	if n := int32(int(t.tableFrame.SurplusDuration(id).Seconds()) - gdy.TimeOffset); n > 0 {
		return n
	}

	return 0
}

// Release 每局结束重置数据
func (t *Table) Release() {
	t.currentStatus = gdy.StatusFree
	t.currentUser = define.INVALID_CHAIR_ID
	t.turnWiner = define.INVALID_CHAIR_ID
	t.leftCardCount = 0
	t.turnPokerType = 0
	t.turnPokerCount = 0
	t.turnMagicCardNum = 0
	t.turnMinCard = 0
	t.currentUserMagicCardNum = 0
	t.turnChairID = define.INVALID_CHAIR_ID
	t.laiziPoker = 0
	t.bombNum = 0

	zeroSlice(t.handCount, reflect.Int32)
	zeroSlice(t.gameScore, reflect.Int64)
	zeroSlice(t.bombScore, reflect.Int64)
	zeroSlice(t.turnPoker, reflect.Int32)
	zeroSlice(t.turnRealPoker, reflect.Int32)
	zeroSlice(t.turnShamCard, reflect.Int32)
	zeroSlice(t.bombCount, reflect.Int32)
	zeroSlice(t.outCard, reflect.Bool)
	for i := int32(0); i < t.playerCount; i++ {
		zeroSlice(t.handPoker[i], reflect.Int32)
	}
}

// GetTableFrame 获取桌子框架接口
func (t *Table) GetTableFrame() common.ITableFrame {
	return t.tableFrame
}

// GetGameRule 获取房间配置
func (t *Table) GetGameRule() int64 {
	return t.gameRule
}

// GetLog 获取日志接口
func (t *Table) GetLog() *utils.Logger {
	return t.log
}

// GetRoomType 获取GetRoomType
func (t *Table) GetRoomType() int32 {
	return t.roomType
}

// GetPlayerCount 获取玩家数量
func (t *Table) GetPlayerCount() int32 {
	return t.playerCount
}

// GetPokerCount 获取牌库牌数
func (t *Table) GetPokerCount() int32 {
	return t.pokerCount
}

// GetHaveOutCard 是否被关
func (t *Table) GetHaveOutCard(chairID int32) int64 {
	if t.outCard[chairID] {
		return int64(1)
	}
	return int64(2)
}

// SetPassivityNoCard 被关次数
func (t *Table) SetPassivityNoCard(chairID int32) {
	if t.outCard[chairID] == false {
		t.passivityNoCard[chairID]++
	}
}

// SetSpring 春天
func (t *Table) SetSpring(chairID int32) {
	t.spring[chairID]++
	t.SpringChairID = chairID
	for i := int32(0); i != t.playerCount; i++ {
		if i == chairID {
			continue
		}
		if t.outCard[i] == true {
			t.spring[chairID]--
			t.SpringChairID = define.INVALID_CHAIR_ID
			break
		}
	}
}

// SetPokerCount 设置牌库牌数
func (t *Table) SetPokerCount(pokerCount int32) {
	t.pokerCount = pokerCount
}

// GetPokerData 获取牌库
func (t *Table) GetPokerData() *[]int32 {
	return &t.pokerData
}

// GetHandPokerCount 获取玩家初始手牌数量
func (t *Table) GetHandPokerCount() int32 {
	return t.handPokerCount
}

// SetHandPokerCount 设置玩家初始手牌数量
func (t *Table) SetHandPokerCount(handPokerCount int32) {
	t.handPokerCount = handPokerCount
}

// GetHandPoker 获取所有玩家手牌信息地址
func (t *Table) GetHandPoker() *[][]int32 {
	return &t.handPoker
}

// GetLeftPokerCount 返回剩余扑克数量
func (t *Table) GetLeftPokerCount() *int32 {
	return &t.leftCardCount
}

// GetHandCount 获取所有玩家手牌数量地址
func (t *Table) GetHandCount() *[]int32 {
	return &t.handCount
}

// GetBankerUser 获取庄家
func (t *Table) GetBankerUser() int32 {
	return t.bankerUser
}

// SetBankerUser 设置庄家
func (t *Table) SetBankerUser(banker int32) {
	t.bankerUser = banker
}

// GetWinner 获取上个管牌成功玩家ChairID
func (t *Table) GetWinner() int32 {
	return t.turnWiner
}

// GetGameLogic 获取游戏规则
func (t *Table) GetGameLogic() *GameLogic {
	return t.gameLogic
}

// GetTurnPoker 获取上个出牌玩家出牌数据
func (t *Table) GetTurnPoker() *[]int32 {
	return &t.turnPoker
}

// GetTurnChairID 获取上个出牌玩家ChairID
func (t *Table) GetTurnChairID() int32 {
	return t.turnChairID
}

// GetTurnPokerCount 获取上个出牌玩家出牌数量
func (t *Table) GetTurnPokerCount() int32 {
	return t.turnPokerCount
}

// GetCurrentUser 获取当前操作玩家
func (t *Table) GetCurrentUser() int32 {
	return t.currentUser
}

// SetCurrentUser 设置当前操作玩家
func (t *Table) SetCurrentUser(user int32) {
	t.currentUser = user
}

// GetGameScore 获取游戏分数地址方便算分
func (t *Table) GetGameScore() *[]int64 {
	return &t.gameScore
}

// SetTurnPokerCount 设置上个出牌玩家出牌数量
func (t *Table) SetTurnPokerCount(turnCount int32) {
	t.turnPokerCount = turnCount
}

// GetBombCount 获取炸弹列表
func (t *Table) GetBombCount() *[]int32 {
	return &t.bombCount
}

// SetGameStatus 设置游戏当前场景
func (t *Table) SetGameStatus(status gdy.GameStatusType) {
	t.currentStatus = status
}

// GetGameStatus 设置游戏当前场景
func (t *Table) GetGameStatus() gdy.GameStatusType {
	return t.currentStatus
}

// GameStart 游戏开始(发牌)
func (t *Table) GameStart() bool {
	t.log.Infof("[%d] [游戏开始] GameStart\n", t.roomID)
	if t.currentStatus == gdy.StatusPlay {
		return true
	}

	if t.DispatchPoker() == false {
		return false
	}

	t.NotifyGameStart()

	return true
}

// NotifyGameStart 游戏开始（发牌后）
func (t *Table) NotifyGameStart() bool {
	gamesCount, startGamesCount := t.tableFrame.GetGamesCount()
	//通知游戏开始
	for i := int32(0); i < t.playerCount; i++ {
		item := t.tableFrame.GetTableUserItem(int(i))
		if item == nil {
			continue
		}

		var handCount []int32
		t.GetPlayersHandCount(&handCount)

		notify := &message.GdyNotifyGameStart{
			BankerUser:      proto.Int32(t.bankerUser),
			CurrentUser:     proto.Int32(t.currentUser),
			PokerCount:      handCount,
			Poker:           t.handPoker[i],
			GamesCount:      proto.Int32(gamesCount),
			StartGamesCount: proto.Int32(startGamesCount),
			ChairId:         proto.Int32(int32(i)),
		}

		t.log.Infof("[%d] [通知玩家游戏开始pb] %v\n", t.tableFrame.GetRoomId(), notify)
		t.tableFrame.SendChairPbMessage(int(i), gdy.MAIN_GAME_ID, gdy.SUB_S_NotifyGameStart, notify)
	}

	if t.strategy.CheckSpecial() == true {
		return true
	}

	notify := &message.GdyNotifyOutPoker{
		ChairId: proto.Int32(t.bankerUser),
		OutType: proto.Int32(1), // 庄家必须出牌
	}

	t.log.Infof("[%d] [通知玩家出牌] %d %v\n", t.tableFrame.GetRoomId(), t.currentUser, notify)
	t.tableFrame.SendChairPbMessage(int(t.bankerUser), gdy.MAIN_GAME_ID, gdy.SUB_S_NotifyOutCard, notify)
	//  ???
	broadcast := &message.GdyNotifyOutPoker{
		ChairId: proto.Int32(t.bankerUser),
	}
	t.tableFrame.SendTableOtherPbMessage(int(t.bankerUser), gdy.MAIN_GAME_ID, gdy.SUB_S_NotifyOutCard, broadcast)

	if t.gameRule&gdy.PyOperate15 != 0 {
		t.log.Infof("[%d] [15s定时器]\n", t.tableFrame.GetRoomId())
		t.tableFrame.AddTimer(gdy.DefaultTimer, time.Duration(gdy.GAMETIMERPLAY15+gdy.TimeOffset)*time.Second, nil, false)
	} else if t.gameRule&gdy.PyOperate10 != 0 {
		t.log.Infof("[%d] [10s定时器]\n", t.tableFrame.GetRoomId())
		t.tableFrame.AddTimer(gdy.DefaultTimer, time.Duration(gdy.GAMETIMERPLAY10+gdy.TimeOffset)*time.Second, nil, false)
	}

	t.currentStatus = gdy.StatusPlay
	return true
}

// DispatchPoker 发牌
func (t *Table) DispatchPoker() bool {

	//设置牌数据
	t.strategy.SetPokerData()

	userCount := int32(0)

	//遍历用户 确保游戏可以开始
	for i := int32(0); i < t.playerCount; i++ {
		item := t.tableFrame.GetTableUserItem(int(i))
		if item == nil {
			continue
		}

		userCount++
	}

	//游戏不能开始
	if userCount < t.playerCount {
		return false
	}

	t.leftCardCount = t.pokerCount

	//确定赖子
	//	t.EnsureMagic()
	//确定首局庄家
	if t.bankerUser == define.INVALID_CHAIR_ID {
		t.strategy.EnsureFirstBanker()
	}
	/*
		card1 := []int32{
			0x01, 0x02, 0x03, 0x04, 0x05, 0x1B, 0x1C, 0x1D, 0x2B, 0x2C, 0x2D, 0x4e, 0x4f, //黑桃 A - K
		}
	*/
	/*
		card1 := []int32{
			0x1a, 0x2a, 0x3a, 0x0d, 0x2d, 0x3d, 0x4e, 0x4f, //黑桃 A - K
		}
	*/
	for i := int32(0); i < t.playerCount; i++ {
		/*
			if i == t.bankerUser {
				continue
			}
		*/
		t.handPoker[i] = t.handPoker[i][:0] // 游戏开始重置手牌

		t.handPoker[i] = make([]int32, t.handPokerCount)

		t.leftCardCount -= t.handPokerCount
		copy(t.handPoker[i][0:], t.pokerData[t.leftCardCount:t.leftCardCount+t.handPokerCount])
		t.handCount[i] = t.handPokerCount

		/*
			t.handPoker[i] = append(t.handPoker[i], card1...)
			t.leftCardCount -= int32(len(t.handPoker[i]))
			t.handCount[i] = int32(len(t.handPoker[i]))
		*/
		sort.Sort(SortInt32(t.handPoker[i]))
	}
	/*
		cards := []int32{
			0x01, 0x02, 0x03, 0x04, 0x05, 0x1B, 0x1C, 0x1D, 0x2B, 0x2C, 0x2D, 0x4e, 0x4f, //黑桃 A - K
		}
	*/
	/*
		cards := []int32{
			0x0c, 0x1c, 0x2c, 0x3c, 0x21,
		}
	*/

	t.leftCardCount--
	t.handPoker[t.bankerUser] = append(t.handPoker[t.bankerUser], t.pokerData[t.leftCardCount]) // 庄家初始手牌为6张
	t.handCount[t.bankerUser]++

	/*
		t.leftCardCount -= int32(len(cards))
		t.handPoker[t.bankerUser] = append(t.handPoker[t.bankerUser], cards...) // 特殊手牌
		t.handCount[t.bankerUser] += int32(len(cards))
	*/
	t.currentUser = t.bankerUser
	sort.Sort(SortInt32(t.handPoker[t.bankerUser]))
	return true
}

// EnsureBanker 确定庄家
func (t *Table) EnsureBanker() {
	if t.bankerUser == define.INVALID_CHAIR_ID {

		t.bankerUser = t.tableFrame.GetCreateUserChairId()

	}
}

// GetPlayersHandCount 获得所有玩家手牌个数
func (t *Table) GetPlayersHandCount(handCount *[]int32) {
	(*handCount) = append((*handCount), t.handCount...)
}

// GetPlayerHandCount 获得玩家手牌个数
func (t *Table) GetPlayerHandCount(chairID int) int32 {
	return t.handCount[chairID]

}

// OnUserOutPoker 用户出牌
func (t *Table) OnUserOutPoker(data []byte, item common.IServerUserItem) bool {
	chairID := item.GetChairID()
	if chairID == define.INVALID_CHAIR_ID {
		t.log.Errorln("[%d] OnUserOutPoker chairId null\n", t.tableFrame.GetRoomId())
		return true
	}

	if t.currentStatus != gdy.StatusPlay {
		t.log.Errorf("[%d]  OnUserOutPoker currentStatus is free\n", t.tableFrame.GetRoomId())
		broadcast := &message.GdyBroadcastOutPoker{
			ChairId:   proto.Int32(int32(chairID)),
			ErrorCode: message.GdyBroadcastOutPoker_POP_Status.Enum(),
		}

		t.log.Infof("[%d] [广播玩家打牌pb] %d %v\n", t.tableFrame.GetRoomId(), chairID, broadcast)
		t.tableFrame.SendChairPbMessage(chairID, gdy.MAIN_GAME_ID, gdy.SUB_S_BroadcastOutCard, broadcast)
		return true
	}

	if t.currentUser != int32(chairID) {
		t.log.Errorf("[%d] OnUserOutPoker current user(%d) not match, not %d", t.tableFrame.GetRoomId(), t.currentUser, chairID)

		broadcast := &message.GdyBroadcastOutPoker{
			ChairId:   proto.Int32(int32(chairID)),
			ErrorCode: message.GdyBroadcastOutPoker_POP_Current.Enum(),
		}

		t.log.Infof("[%d] [广播玩家打牌pb] %d %v\n", t.tableFrame.GetRoomId(), chairID, broadcast)
		t.tableFrame.SendChairPbMessage(chairID, gdy.MAIN_GAME_ID, gdy.SUB_S_BroadcastOutCard, broadcast)
		return true
	}

	outCard := &message.GdyOutPoker{}
	err := proto.Unmarshal(data, outCard)
	if err != nil {
		t.log.Errorln("Unmarshal GdyOutPoker", err)
		return true
	}

	//过牌
	if outCard.GetType() == 2 {
		return t.OnUserPassPoker(chairID)
	}

	if len(outCard.GetOutPoker()) != 0 { // 存在多个牌型选择时
		chooseOutCard := &message.GdyNotifyChooseOutPoker{}
		t.log.Infof("[%d] 多个牌型选择len(outCard.GetOutPoker()):%d", t.tableFrame.GetRoomId(), len(outCard.GetOutPoker()))
		for i := 0; i != len(outCard.GetOutPoker()); i++ {
			t.log.Infof("[%d]turnPokerCount: %d", t.turnPokerCount)
			if t.turnPokerCount != 0 { // 管牌
				outCardNum := int32(len(outCard.OutPoker[i].GetPoker()))
				t.log.Infof("[%d] 校验牌型:%xv, 上个操作玩家牌型：%xv", t.tableFrame.GetRoomId(), outCard.OutPoker[i].GetPoker(), t.turnPoker)
				if t.gameLogic.CompareCard(outCard.OutPoker[i].GetPoker(), t.turnPoker, outCardNum, t.turnPokerCount, t.currentUserMagicCardNum, t.turnMagicCardNum) == true {
					chooseOutCard.OutPokerIndex = append(chooseOutCard.OutPokerIndex, int32(i))
				}
			} else {
				t.log.Errorf("[%d] turnPokerCount is nul!", t.tableFrame.GetRoomId()) // 自由出牌不符合
				broadcast := &message.GdyBroadcastOutPoker{
					ChairId:   proto.Int32(int32(chairID)),
					ErrorCode: message.GdyBroadcastOutPoker_POP_Error.Enum(),
				}

				t.log.Infof("[%d] [广播玩家打牌pb] %d %v\n", t.tableFrame.GetRoomId(), chairID, broadcast)
				t.tableFrame.SendChairPbMessage(chairID, gdy.MAIN_GAME_ID, gdy.SUB_S_BroadcastOutCard, broadcast)
				return true
			}
		}
		if len(chooseOutCard.OutPokerIndex) == 0 {
			chooseOutCard.OutPokerIndex = append(chooseOutCard.OutPokerIndex, int32(-1))
		}
		t.log.Infof("[%d] [玩家选择出牌对应下标]：%v", t.tableFrame.GetRoomId(), chooseOutCard)
		t.tableFrame.SendChairPbMessage(chairID, gdy.MAIN_GAME_ID, gdy.SUB_S_NotifyChooseOutCard, chooseOutCard)
		return true
	}
	outCardData := make([]int32, len(outCard.GetPoker())) // 客户端想要发来牌型顺序不变返回
	copy(outCardData, outCard.GetPoker())
	outlaiCard := make([]int32, len(outCard.GetPoker()))
	copy(outlaiCard, outCard.GetPoker())
	if len(outCard.GetRealPoker()) != 0 { // 存在赖子
		if t.gameLogic.VerifyShamCard(outlaiCard, outCard.GetRealPoker(), &t.currentUserMagicCardNum) == false { //牌型校验失败
			t.log.Errorf("[%d] The card type check error", t.tableFrame.GetRoomId())

			broadcast := &message.GdyBroadcastOutPoker{
				ChairId:   proto.Int32(int32(chairID)),
				ErrorCode: message.GdyBroadcastOutPoker_POP_NotExist.Enum(),
			}

			t.log.Infof("[%d] [广播玩家打牌pb] %d %v\n", t.tableFrame.GetRoomId(), chairID, broadcast)
			t.tableFrame.SendChairPbMessage(chairID, gdy.MAIN_GAME_ID, gdy.SUB_S_BroadcastOutCard, broadcast)
			return true
		}
	}

	outType := int32(2)
	if t.turnPokerCount == 0 {
		outType = 1
	}
	var pokerType int32
	if len(outCard.GetRealPoker()) != 0 { // 存在赖子
		pokerType, _ = t.gameLogic.GetPokerType(outlaiCard, int32(len(outCard.GetPoker())), t.currentUserMagicCardNum)

	} else {
		pokerType, _ = t.gameLogic.GetPokerType(outCard.GetPoker(), int32(len(outCard.GetPoker())), t.currentUserMagicCardNum)

	}
	code := t.ProcessUserOutPoker(chairID, outCard.GetPoker(), pokerType, outlaiCard)
	if code == gdy.UserOutPokerSuccess {

		pokerCount := t.GetPlayerHandCount(chairID)
		//	sort.Sort(SortInt32(outCard.GetPoker()))
		minCard := (outlaiCard[len(outCard.GetPoker())-1]) & 0xf
		t.turnMinCard = minCard // 记录上个操作玩家最小出牌
		t.turnShamCard = t.turnShamCard[:0]
		t.turnShamCard = append(t.turnShamCard, outCard.GetRealPoker()...) // 记录上个操作玩家赖子顶替的牌
		broadcast := &message.GdyBroadcastOutPoker{
			ChairId:    proto.Int32(int32(chairID)),
			Poker:      outCardData,
			PokerType:  proto.Int32(int32(pokerType)),
			OperType:   proto.Int32(outCard.GetType()),
			ErrorCode:  message.GdyBroadcastOutPoker_ErrorCodeType(code).Enum(),
			PokerCount: proto.Int32(pokerCount),
			OutType:    proto.Int32(outType),
			BombCount:  proto.Int32(t.bombNum),
			MinCard:    proto.Int32(minCard),
			ShamCard:   outCard.GetRealPoker(),
		}

		t.log.Infof("[%d] [广播玩家打牌pb] %d %v\n", t.tableFrame.GetRoomId(), chairID, broadcast)
		t.tableFrame.SendTablePbMessage(gdy.MAIN_GAME_ID, gdy.SUB_S_BroadcastOutCard, broadcast)
		t.log.Infof("[%d] [移除定时器]\n", t.tableFrame.GetRoomId())
		t.tableFrame.RemoveTimer(gdy.DefaultTimer)
		if t.currentUser != define.INVALID_CHAIR_ID {
			//通知玩家出牌
			t.NotifyUserOutPoker()
		}

		//结束判断
		if t.currentUser == define.INVALID_CHAIR_ID {
			t.GameConclude(false)
		}

	} else {
		broadcast := &message.GdyBroadcastOutPoker{
			ChairId:   proto.Int32(int32(chairID)),
			Poker:     outCardData,
			PokerType: proto.Int32(int32(pokerType)),
			OperType:  proto.Int32(outCard.GetType()),
			ErrorCode: message.GdyBroadcastOutPoker_ErrorCodeType(code).Enum(),
		}

		t.log.Infof("[%d] [广播玩家打牌pb] %d %v\n", t.tableFrame.GetRoomId(), chairID, broadcast)
		t.tableFrame.SendChairPbMessage(chairID, gdy.MAIN_GAME_ID, gdy.SUB_S_BroadcastOutCard, broadcast)
	}

	return true
}

// OnUserPassPoker 用户过牌
func (t *Table) OnUserPassPoker(chairID int) bool {

	t.log.Infof("[%d][玩家过牌] %d\n", t.roomID, chairID)
	var outPokerResult OutPokerResult

	code := t.strategy.CheckPassCard(chairID, outPokerResult)
	outType := int32(2)
	if t.turnPokerCount == 0 {
		code = gdy.UserOutPokerError // 自由出牌不允许过牌
		outType = 1
	}
	if code == gdy.UserOutPokerSuccess {

		//是否显示玩家牌个数
		pokerCount := t.GetPlayerHandCount(chairID)

		broadcast := &message.GdyBroadcastOutPoker{
			ChairId:    proto.Int32(int32(chairID)),
			OperType:   proto.Int32(2),
			ErrorCode:  message.GdyBroadcastOutPoker_ErrorCodeType(code).Enum(),
			PokerCount: proto.Int32(pokerCount),
			OutType:    proto.Int32(outType),
			BombCount:  proto.Int32(t.bombNum),
		}
		t.log.Infof("[%d] [广播玩家过牌pb] %d %v\n", t.tableFrame.GetRoomId(), chairID, broadcast)
		t.tableFrame.SendTablePbMessage(gdy.MAIN_GAME_ID, gdy.SUB_S_BroadcastOutCard, broadcast)

		t.log.Infof("[%d] [移除定时器]\n", t.tableFrame.GetRoomId())
		t.tableFrame.RemoveTimer(gdy.DefaultTimer)

	} else {

		//是否显示玩家牌个数
		pokerCount := t.GetPlayerHandCount(chairID)
		broadcast := &message.GdyBroadcastOutPoker{
			ChairId:    proto.Int32(int32(chairID)),
			OperType:   proto.Int32(2),
			ErrorCode:  message.GdyBroadcastOutPoker_ErrorCodeType(code).Enum(),
			PokerCount: proto.Int32(pokerCount),
			OutType:    proto.Int32(outType),
			BombCount:  proto.Int32(t.bombNum),
		}
		t.log.Infof("[%d] [广播玩家过牌pb] %d %v\n", t.tableFrame.GetRoomId(), chairID, broadcast)
		t.tableFrame.SendChairPbMessage(chairID, gdy.MAIN_GAME_ID, gdy.SUB_S_BroadcastOutCard, broadcast)

		return true
	}

	//切换用户
	t.currentUser = (t.currentUser + 1) % t.playerCount

	//计算炸弹分数
	if t.currentUser == t.turnWiner { // 过一轮没人管牌
		t.turnPokerCount = 0
	}
	//通知玩家出牌
	t.NotifyUserOutPoker()
	return true
}

// ProcessUserOutPoker 校验玩家出牌
func (t *Table) ProcessUserOutPoker(chairID int, outPoker []int32, pokerType int32, realOutPoker []int32) gdy.UserOutPokerCode {
	t.log.Infof("[%d] [玩家打牌] %d %xv %d\n", t.tableFrame.GetRoomId(), chairID, outPoker, pokerType)
	if len(outPoker) == 0 {
		t.log.Errorf("[%d] [出牌错误]：nil", t.tableFrame.GetRoomId())
		return gdy.UserOutPokerError
	}
	if pokerType == gdy.CtError && (len(t.handPoker[chairID]) != 1 || t.gameLogic.GetPokerLogicValue(outPoker[0]) < 15) {
		return gdy.UserOutPokerError
	}

	sort.Sort(SortInt32(outPoker))
	sort.Sort(SortInt32(realOutPoker))

	//检查首轮出牌
	code := t.strategy.CheckFirstRound(realOutPoker, chairID)
	if code != gdy.UserOutPokerSuccess {
		return code
	}

	if t.turnPokerCount == 0 {
		//检查特殊牌
		code = t.strategy.CheckSpecialPoker(realOutPoker, pokerType, chairID)
		if code != gdy.UserOutPokerSuccess {
			return code
		}
	}

	code = t.strategy.CheckOutCard(chairID, realOutPoker, pokerType)
	if code != gdy.UserOutPokerSuccess {
		return code
	}

	cardCount := int32(len(realOutPoker))
	turnPokerCountBak := t.turnPokerCount

	//判断是否炸弹
	if pokerType == gdy.CtSoftThreeBomb || pokerType == gdy.CtRealThreeBomb || pokerType == gdy.CtSoftFourBomb ||
		pokerType == gdy.CtRealFourBomb || pokerType == gdy.CtSoftFiveBomb || pokerType == gdy.CtKingBomb { //
		t.bombCount[chairID]++
		t.bombNum++
		t.bombTotalCount[chairID]++

	}
	if !t.outCard[chairID] && t.turnChairID != define.INVALID_CHAIR_ID { //
		t.outCard[chairID] = true // 玩家出牌成功不会被关（春天）
	}

	if t.turnPokerCount == 0 {
		t.turnPokerCount = cardCount
	} else {
		//	t.log.Infof("t.turnPokerCount : %d, cardCount:%d\n", t.turnPokerCount, cardCount)
		//	t.log.Infof("outPoker:%x t.turnPoker:%x\n", outPoker, t.turnPoker)

		if t.gameLogic.CompareCard(realOutPoker, t.turnPoker, cardCount, t.turnPokerCount, t.currentUserMagicCardNum, t.turnMagicCardNum) == false {
			flag := true
			if t.turnPokerCount == 1 && len(t.handPoker[chairID]) == 1 { // 剩下一张牌
				if t.gameLogic.GetPokerLogicValue(t.turnPoker[0]) == 15 {
					if t.gameLogic.GetPokerLogicValue(t.handPoker[chairID][0]) > 15 {
						flag = false
					}
				} else if t.gameLogic.GetPokerLogicValue(t.turnPoker[0]) == 16 {
					if t.gameLogic.GetPokerLogicValue(t.handPoker[chairID][0]) > 15 {
						flag = false
					}
				}
			}
			if flag {
				t.log.Errorf("[%d] UserOutPoker CompareCard", t.tableFrame.GetTableID())
				return gdy.UserOutPokerError
			}
		}
	}
	//删除扑克
	t.log.Infof("t.handCount[chairId] : %xv, t.handCount[chairId]:%d\n", t.handPoker[chairID][0:], t.handCount[chairID])
	t.log.Infof("outPoker:%xv ,cardCount:%x\n", outPoker, cardCount)
	if t.gameLogic.RemovePoker(outPoker, cardCount, &t.handPoker[chairID], t.handCount[chairID]) == false { // 删除原始手牌 赖子做赖子本身
		t.log.Errorf("[%d] UserOutPoker RemovePoker", t.tableFrame.GetTableID())
		t.turnPokerCount = turnPokerCountBak
		return gdy.UserOutPokerNotExist
	}

	t.handCount[chairID] -= cardCount

	//出牌记录
	t.turnPokerType = pokerType
	t.turnPokerCount = cardCount
	t.turnChairID = int32(chairID)
	t.turnMagicCardNum = t.currentUserMagicCardNum // 记录上个玩家赖子数量
	//copy(t.turnPoker, outPoker)
	t.turnPoker = t.turnPoker[:0]                          // 出牌成功重置前者手牌
	t.turnPoker = append(t.turnPoker, realOutPoker...)     // 记录新的手牌记录手牌不包括赖子
	t.turnRealPoker = t.turnRealPoker[:0]                  // 出牌成功重置前者手牌
	t.turnRealPoker = append(t.turnRealPoker, outPoker...) // 记录新的手牌记录手牌不包括赖子
	//切换用户
	t.turnWiner = int32(chairID)
	t.currentUserMagicCardNum = 0 // 重置新的玩家出牌赖子数量
	if t.handCount[int32(chairID)] != 0 {
		t.currentUser = (t.currentUser + 1) % t.playerCount
	} else {
		t.currentUser = define.INVALID_CHAIR_ID
	}

	return gdy.UserOutPokerSuccess
}

// NotifyUserOutPoker 通知用户出牌
func (t *Table) NotifyUserOutPoker() {
	if t.leftCardCount != 0 && t.turnPokerCount == 0 { // 底牌不为空且为赢家（庄家先出牌不在此判断内）
		t.strategy.NotifyDeal()
	}
	var outPokerResult OutPokerResult

	outType := int32(0)

	if t.turnPokerCount == 0 {
		outType = 1
	} else {
		must := t.strategy.SearchOutCard(int(t.currentUser), &outPokerResult)
		if must {
			outType = 1
		}

	}

	notify := &message.GdyNotifyOutPoker{
		ChairId: proto.Int32(t.currentUser),
		OutType: proto.Int32(outType),
	}

	t.log.Infof("[%d] [通知玩家出牌] %d %v\n", t.tableFrame.GetRoomId(), t.currentUser, notify)
	t.tableFrame.SendChairPbMessage(int(t.currentUser), gdy.MAIN_GAME_ID, gdy.SUB_S_NotifyOutCard, notify)

	broadcast := &message.GdyNotifyOutPoker{
		ChairId: proto.Int32(t.currentUser),
	}
	t.tableFrame.SendTableOtherPbMessage(int(t.currentUser), gdy.MAIN_GAME_ID, gdy.SUB_S_NotifyOutCard, broadcast)

	if t.gameRule&gdy.PyOperate15 != 0 {
		t.log.Infof("[%d] [15s定时器]\n", t.tableFrame.GetRoomId())
		t.tableFrame.AddTimer(gdy.DefaultTimer, time.Duration(gdy.GAMETIMERPLAY15+gdy.TimeOffset)*time.Second, nil, false)
	} else if t.gameRule&gdy.PyOperate10 != 0 {
		t.log.Infof("[%d] [10s定时器]\n", t.tableFrame.GetRoomId())
		t.tableFrame.AddTimer(gdy.DefaultTimer, time.Duration(gdy.GAMETIMERPLAY10+gdy.TimeOffset)*time.Second, nil, false)
	}

}

// CalcBombScore 计算炸弹数量
func (t *Table) CalcBombScore() {
	//计算炸弹分数
	if t.currentUser == t.turnWiner { // 过一轮没人管牌
		//判断是否炸弹
		pokerType, _ := t.gameLogic.GetPokerType(t.turnPoker, t.turnPokerCount, t.currentUserMagicCardNum)
		if pokerType == gdy.CtSoftThreeBomb || pokerType == gdy.CtRealThreeBomb || pokerType == gdy.CtSoftFourBomb ||
			pokerType == gdy.CtRealFourBomb || pokerType == gdy.CtSoftFiveBomb || pokerType == gdy.CtKingBomb { //
			t.bombCount[t.turnWiner]++
			t.bombNum++
			t.bombTotalCount[t.turnWiner]++

		}
		t.turnPokerCount = 0
	}
}

// CalcResult 计算游戏结果
func (t *Table) CalcResult() {

	t.strategy.CalcGameScore() // 算分

	scoreInfo := make([]define.TagScoreInfo, t.playerCount)
	for i := int32(0); i < t.playerCount; i++ {
		item := t.tableFrame.GetTableUserItem(int(i))
		if item == nil {
			continue
		}

		//	t.gameScore[i] += t.bombScore[i]

		if t.gameScore[i] > t.maxWinScore[i] {
			t.maxWinScore[i] = t.gameScore[i]
		}

		if t.gameScore[i] > 0 {
			t.winCount[i]++
		} else if t.gameScore[i] <= 0 {
			t.lostCount[i]++
		}

		t.totalScore[i] += t.gameScore[i]
		scoreInfo[i].Score = t.gameScore[i]
		scoreInfo[i].Type = define.ScoreChangeType_Game
		scoreInfo[i].UserID = item.GetUserID()
	}
	t.tableFrame.WriteTableScore(scoreInfo)

	t.strategy.EnsureGameOverBanker()
	//t.bankerUser = t.turnWiner
}

// NotifyGameConclude 通知单局游戏结束
func (t *Table) NotifyGameConclude(dismiss bool) {
	userScore := make([]int64, t.playerCount)

	var leftPoker []*message.GdyPokers

	for i := int32(0); i < t.playerCount; i++ {
		item := t.tableFrame.GetTableUserItem(int(i))
		if item != nil {
			userScore[i] = item.GetUserScore()
		}

		leftPoker = append(leftPoker, &message.GdyPokers{
			Poker: t.handPoker[i][0:t.handCount[i]],
		})
	}

	gamesCount, startGamesCount := t.tableFrame.GetGamesCount()
	if !t.outCard[t.bankerUser] && (t.handCount[t.bankerUser] == 0) {
		t.outCard[t.bankerUser] = true
	}
	broadcast := &message.GdyBroadcastGameEnd{
		WinChairId:      proto.Int32(t.turnWiner),
		StartGamesCount: proto.Int32(startGamesCount),
		GamesCount:      proto.Int32(gamesCount),
		UserScore:       userScore,
		GameScore:       t.gameScore,
		BombCount:       t.bombCount,
		LeftCount:       t.handCount,
		Time:            proto.Int64(time.Now().UnixNano()),
		LeftPoker:       leftPoker,
		Dismiss:         proto.Bool(dismiss),
		SpringChairID:   proto.Int32(t.SpringChairID),
		NoOutCard:       t.outCard,
	}

	t.log.Infof("[%d] [广播游戏结束pb] %v\n", t.tableFrame.GetRoomId(), broadcast)
	t.tableFrame.SendTablePbMessage(gdy.MAIN_GAME_ID, gdy.SUB_S_BroadcastGameEnd, broadcast)

}

// GameConclude 单局游戏结束
func (t *Table) GameConclude(dismiss bool) {
	t.log.Infof("[%d] [游戏结束移除定时器]\n", t.tableFrame.GetRoomId())
	t.tableFrame.RemoveTimer(gdy.DefaultTimer)
	if dismiss == false {
		// 计算游戏结果
		t.CalcResult()
	}

	// 通知游戏结束
	t.NotifyGameConclude(dismiss)
	// 框架结算游戏分数
	go utils.SafeCall(func(args ...interface{}) error {
		t.tableFrame.ConcludeGame()
		return nil
	})

	t.Release()
}

// NotifyGameEnd 游戏总结束
func (t *Table) NotifyGameEnd() bool {
	broadcast := &message.GdyBroadcastGameOver{
		HouseOwner:      proto.Int32(t.tableFrame.GetCreateUserId()),
		LostCount:       t.lostCount[0:],
		WinCount:        t.winCount[0:],
		MaxWinScore:     t.maxWinScore[0:],
		BombCount:       t.bombTotalCount[0:],
		TotalScore:      t.totalScore[0:],
		Time:            proto.Int64(time.Now().UnixNano()),
		PassivityNoCard: t.passivityNoCard,
		Spring:          t.spring,
	}

	t.log.Infof("[%d] [广播牌局结束pb] %v\n", t.tableFrame.GetRoomId(), broadcast)
	t.tableFrame.SendTablePbMessage(gdy.MAIN_GAME_ID, gdy.SUB_S_BroadcastGameOver, broadcast)

	zeroSlice(t.lostCount[0:], reflect.Int32)
	zeroSlice(t.winCount[0:], reflect.Int32)
	zeroSlice(t.maxWinScore[0:], reflect.Int64)
	zeroSlice(t.bombCount[0:], reflect.Int32)
	zeroSlice(t.totalScore[0:], reflect.Int64)
	t.bankerUser = define.INVALID_CHAIR_ID
	return true
}

// OnActionUserSitDown 玩家进入并坐下
func (t *Table) OnActionUserSitDown(chairID int, userItem common.IServerUserItem, siteType common.SitdownType) bool {
	t.log.Infof("[%d] [用户坐下] [场景] %d %d %d\n", t.roomID, chairID, userItem.GetUserID(), t.currentStatus)
	switch t.currentStatus {
	case gdy.StatusFree:
		t.BroadcastSceneSceneGameFree(chairID, userItem, siteType)
	case gdy.StatusPlay:
		t.BroadcastSceneSceneGamePlay(chairID, userItem, siteType)
	}

	return true
}

// BroadcastSceneSceneGameFree 游戏空闲场景
func (t *Table) BroadcastSceneSceneGameFree(chairID int, item common.IServerUserItem, siteType common.SitdownType) {
	gamesCount, startGamesCount := t.tableFrame.GetGamesCount()
	costMode, _ := t.tableFrame.GetPrivateCostMode()

	//发送空闲场景
	free := &message.GdyBroadcastSceneGameFree{
		RoomType:        proto.Int32(t.roomType),
		GameRule:        proto.Int64(t.gameRule),
		OperateTime:     proto.Int32(15),
		GamesCount:      proto.Int32(gamesCount),
		StartGamesCount: proto.Int32(startGamesCount),
		HouseOwner:      proto.Int32(t.tableFrame.GetCreateUserId()),
		CostMode:        proto.Int32(costMode),
	}
	t.log.Infof("[%d] [用户] [空闲场景] %d %d %d, [场景消息] %v\n", t.roomID, chairID, item.GetUserID(), t.currentStatus, free)
	t.tableFrame.SendChairPbSceneMessage(chairID, gdy.MAIN_GAME_ID, gdy.SUB_S_BroadcastSceneGameFree, free)
}

// BroadcastSceneSceneGamePlay 游戏进行场景
func (t *Table) BroadcastSceneSceneGamePlay(chairID int, item common.IServerUserItem, siteType common.SitdownType) {
	currentUser := t.currentUser

	gamesCount, startGamesCount := t.tableFrame.GetGamesCount()
	costMode, _ := t.tableFrame.GetPrivateCostMode()

	var outPokerResult OutPokerResult
	outType := int32(0)
	if t.currentUser == int32(chairID) {

		if t.turnPokerCount == 0 {
			outType = 1
		} else {
			must := t.strategy.SearchOutCard(int(t.currentUser), &outPokerResult)
			if must {
				outType = 1
			}
		}
	}

	play := &message.GdyBroadcastSceneGamePlay{
		RoomType:        proto.Int32(t.roomType),
		GameRule:        proto.Int64(t.gameRule),
		GamesCount:      proto.Int32(gamesCount),
		StartGamesCount: proto.Int32(startGamesCount),
		HouseOwner:      proto.Int32(t.tableFrame.GetCreateUserId()),
		CostMode:        proto.Int32(costMode),
		LeftTime:        proto.Int32(t.SurplusTime(gdy.DefaultTimer)),
		BankerUser:      proto.Int32(t.bankerUser),
		CurrentUser:     proto.Int32(currentUser),
		Pokers:          t.handPoker[chairID][0:t.handCount[chairID]],
		PokerCount:      t.handCount,
		OutType:         proto.Int32(outType),
		TurnPoker:       t.turnRealPoker[0:t.turnPokerCount],
		TurnChairId:     proto.Int32(t.turnChairID),
		LeftPokerCount:  proto.Int32(t.leftCardCount),
		BombCount:       proto.Int32(t.bombNum),
		TurnPokerType:   proto.Int32(t.turnPokerType),
		TurnMinCard:     proto.Int32(t.turnMinCard),
		TurnShamCard:    t.turnShamCard,
	}

	t.tableFrame.SendChairPbSceneMessage(chairID, gdy.MAIN_GAME_ID, gdy.SUB_S_BroadcastSceneGamePlay, play)
	t.log.Infof("[%d] [游戏场景] %v\n", t.roomID, play)
}

// CheckSpecialPoker 检查特殊牌
func (t *Table) CheckSpecialPoker(poker []int32, pokerType int32, chairID int) gdy.UserOutPokerCode {

	return gdy.UserOutPokerSuccess
}

// CheckPassCard 检查出牌
func (t *Table) CheckPassCard(chairID int, outPokerResult OutPokerResult) gdy.UserOutPokerCode {
	var code gdy.UserOutPokerCode
	if t.gameRule&gdy.PyMust != 0 {
		code = gdy.UserOutPokerMustMode
	} else {
		code = gdy.UserOutPokerSuccess
	}

	return code
}

// CalcGameScore 算分
func (t *Table) CalcGameScore() {
	for i := int32(0); i < t.playerCount; i++ {
		if t.handCount[i] > 1 {
			times := int32(1)
			if t.handCount[i] == t.handPokerCount {
				times = 2
			} else {
				times = 1
			}

			t.gameScore[i] -= int64(t.handCount[i] * times)
			t.gameScore[t.turnWiner] += int64(t.handCount[i] * times)
		}
	}
}

// NewTable table实例化
func NewTable() *Table {
	return &Table{
		log:       utils.NewLogger(),
		gameLogic: NewGameLogic(),
	}
}

// zeroSlice 重置数据
func zeroSlice(v interface{}, refType reflect.Kind) {
	if reflect.TypeOf(v).Kind() != reflect.Slice {
		panic("zeroSlice wrong type")
	}
	slice := reflect.ValueOf(v)

	for i := 0; i < slice.Len(); i++ {
		value := slice.Index(i)
		if refType >= reflect.Uint && refType <= reflect.Uintptr {
			value.SetUint(0)
		} else if refType >= reflect.Int && refType <= reflect.Int64 {
			value.SetInt(0)
		} else if refType == reflect.Bool {
			value.SetBool(false)
		}
	}
}
