package gdy

import (
	"mjserver/game/gdy/gdefine"
	"mjserver/game/gdy/gdylogic"
	"mjserver/message"
	"sort"

	"github.com/gogo/protobuf/proto"
)

// SortInt32 []int32
type SortInt32 []int32

func (a SortInt32) Len() int      { return len(a) }
func (a SortInt32) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SortInt32) Less(i, j int) bool {

	b := a.LogicValue(a[i])
	c := a.LogicValue(a[j])
	return b > c

}

// LogicValue  ...
func (a SortInt32) LogicValue(value int32) int32 {
	cardValue := value & 0x0f

	if cardValue >= 0xe {
		return cardValue + 2
	}

	if cardValue <= 2 {
		return cardValue + 13
	}
	return cardValue
}

// TableLogic 私有玩法
type TableLogic struct {
	table *gdylogic.Table
}

// SetLogicCfg 设置牌和手牌个数
func (t *TableLogic) SetLogicCfg() {

	t.table.SetPokerCount(54)
	t.table.SetHandPokerCount(5)

}

// SetPokerData 设置牌数据
func (t *TableLogic) SetPokerData() {
	pokerData := t.table.GetPokerData()
	pokerCount := t.table.GetPokerCount()
	if pokerCount == 54 {
		copy((*pokerData)[0:], gdylogic.RandomShuffle(gdylogic.Card54))
	}
}

// EnsureFirstBanker 确定首局庄家
func (t *TableLogic) EnsureFirstBanker() {
	if t.table.GetGameRule()&gdy.PyRandBanker != 0 {
		var tmp []int32
		for i := int32(0); i < t.table.GetPlayerCount(); i++ {

			tmp = append(tmp, i)

		}
		t.table.SetBankerUser(gdylogic.RandomShuffle(tmp)[0]) //第一局随机庄家
	} else {
		Instead := t.table.GetTableFrame().IsInsteadRoom() // 判断代开房间
		if !Instead {
			t.table.SetBankerUser(t.table.GetTableFrame().GetCreateUserChairId())
		} else {
			var tmp []int32
			for i := int32(0); i < t.table.GetPlayerCount(); i++ {

				tmp = append(tmp, i)

			}
			t.table.SetBankerUser(gdylogic.RandomShuffle(tmp)[0]) //第一局随机庄家
		}
	}
}

// EnsureGameOverBanker 确定游戏结束庄家
func (t *TableLogic) EnsureGameOverBanker() {
	if t.table.GetGameRule()&gdy.PyRandBanker != 0 {
		var tmp []int32
		for i := int32(0); i < t.table.GetPlayerCount(); i++ {

			tmp = append(tmp, i)

		}
		t.table.SetBankerUser(gdylogic.RandomShuffle(tmp)[0]) //第一局随机庄家
	} else {
		t.table.SetBankerUser(t.table.GetWinner())
	}
}

// CheckFirstRound 检查首轮出牌
func (t *TableLogic) CheckFirstRound(poker []int32, chairID int) gdy.UserOutPokerCode {

	return gdy.UserOutPokerSuccess
}

// CheckSpecialPoker 检查特殊牌
func (t *TableLogic) CheckSpecialPoker(poker []int32, pokerType int32, chairID int) gdy.UserOutPokerCode {
	return t.table.CheckSpecialPoker(poker, pokerType, chairID)
}

// CheckPassCard 检查过牌
func (t *TableLogic) CheckPassCard(chairID int, outPokerResult gdylogic.OutPokerResult) gdy.UserOutPokerCode {
	return t.table.CheckPassCard(chairID, outPokerResult)
}

// SearchOutCard 检查出牌
func (t *TableLogic) SearchOutCard(chairID int, outPokerResult *gdylogic.OutPokerResult) bool {
	if t.table.GetTurnPokerCount() == 0 {
		return true
	}
	if t.table.GetGameRule()&gdy.PyMust != 0 {
		return true
	}
	return false
}

// NotifyDeal 发牌
func (t *TableLogic) NotifyDeal() { // 暂时只针对一个人发牌
	leftCardCount := t.table.GetLeftPokerCount()
	pokerData := t.table.GetPokerData()
	handPoker := t.table.GetHandPoker()
	handCount := t.table.GetHandCount()
	if t.table.GetGameRule()&gdy.PyWinReplenishPoke != 0 {
		currentUser := t.table.GetCurrentUser()
		t.table.GetLog().Infof("[%d] 赢家发牌： %d", t.table.GetTableFrame().GetRoomId(), currentUser)
		(*leftCardCount)--
		card := (*pokerData)[*leftCardCount]
		var cards []int32
		(*handPoker)[currentUser] = append((*handPoker)[currentUser], card)
		sort.Sort(SortInt32((*handPoker)[currentUser])) // 发牌后排序
		(*handCount)[currentUser]++
		cards = append(cards, card)
		t.table.GetLog().Infof("[%d] %d玩家发牌：%x ", t.table.GetTableFrame().GetRoomId(), currentUser, card)
		t.table.GetLog().Infof("[%d] %d玩家当前手牌:%xv, 手牌数量：%d", t.table.GetTableFrame().GetRoomId(), currentUser, (*handPoker)[currentUser], (*handCount)[currentUser])
		notify := &message.GdyBroadcastDealPokers{
			Poker:      cards,
			ChairId:    proto.Int32(currentUser),
			PokerCount: proto.Int32((*handCount)[currentUser]),
		}
		t.table.GetTableFrame().SendChairPbMessage(int(currentUser), gdy.MAIN_GAME_ID, gdy.SUB_S_BroadcastDeal, notify)
		broad := &message.GdyBroadcastDealPokers{
			ChairId:    proto.Int32(currentUser),
			PokerCount: proto.Int32((*handCount)[currentUser]),
		}
		t.table.GetTableFrame().SendTableOtherPbMessage(int(currentUser), gdy.MAIN_GAME_ID, gdy.SUB_S_BroadcastDeal, broad)

	} else {
		t.table.GetLog().Infof("[%d] 所有玩家发牌", t.table.GetTableFrame().GetRoomId())
		dealUser := t.table.GetCurrentUser()
		for i := int32(0); i != t.table.GetPlayerCount(); i++ {
			if (*leftCardCount) == 0 { // 牌库发完
				return // == break
			}
			(*leftCardCount)--
			card := (*pokerData)[*leftCardCount]
			var cards []int32 // 声明重置

			(*handPoker)[dealUser] = append((*handPoker)[dealUser], card)
			sort.Sort(SortInt32((*handPoker)[dealUser])) // 发牌后排序
			(*handCount)[dealUser]++
			cards = append(cards, card)
			t.table.GetLog().Infof("[%d] %d玩家发牌：%x ", t.table.GetTableFrame().GetRoomId(), dealUser, card)
			t.table.GetLog().Infof("[%d] %d玩家当前手牌:%xv, 手牌数量：%d", t.table.GetTableFrame().GetRoomId(), dealUser, (*handPoker)[dealUser], (*handCount)[dealUser])
			notify := &message.GdyBroadcastDealPokers{
				Poker:      cards,
				ChairId:    proto.Int32(dealUser),
				PokerCount: proto.Int32((*handCount)[dealUser]),
			}
			t.table.GetTableFrame().SendChairPbMessage(int(dealUser), gdy.MAIN_GAME_ID, gdy.SUB_S_BroadcastDeal, notify)
			broad := &message.GdyBroadcastDealPokers{
				ChairId:    proto.Int32(dealUser),
				PokerCount: proto.Int32((*handCount)[dealUser]),
			}
			t.table.GetTableFrame().SendTableOtherPbMessage(int(dealUser), gdy.MAIN_GAME_ID, gdy.SUB_S_BroadcastDeal, broad)
			dealUser = (dealUser + 1) % t.table.GetPlayerCount()
		}
	}
}

// CalcGameScore 计算游戏分数
func (t *TableLogic) CalcGameScore() {

	userHandCount := t.table.GetHandCount()
	userGameScore := t.table.GetGameScore()
	userHandPoker := t.table.GetHandPoker()
	bombs := t.table.GetBombCount()

	var allBombScore int32
	for i := int32(0); i < t.table.GetPlayerCount(); i++ {
		allBombScore += (*bombs)[i] * 2
	}
	winuser := int32(0)
	winuserScore := int64(0)
	DefaultRatio := int64(t.table.GetTableFrame().GetGameConfig().BaseBet)
	for i := int32(0); i < t.table.GetPlayerCount(); i++ {
		times := 1 // 循环重置

		if (*userHandCount)[i] != 0 { // 输家
			if t.table.GetGameRule()&gdy.Py2Double != 0 {
				for j := 0; j != len((*userHandPoker)[i]); j++ {
					if t.table.GetGameLogic().GetPokerLogicValue((*userHandPoker)[i][j]) == 15 {
						times++
						//break
					}
				}
			}
			if t.table.GetGameRule()&gdy.PyKingDouble != 0 {
				for j := 0; j != len((*userHandPoker)[i]); j++ {
					if t.table.GetGameLogic().GetPokerLogicValue((*userHandPoker)[i][j]) == 16 || t.table.GetGameLogic().GetPokerLogicValue((*userHandPoker)[i][j]) == 17 {
						times++
						//break
					}
				}
			}
			if allBombScore == 0 {
				allBombScore = 1
			}
			(*userGameScore)[i] -= int64((*userHandCount)[i]*allBombScore*int32(times)) * t.table.GetHaveOutCard(i) * DefaultRatio
			item := t.table.GetTableFrame().GetTableUserItem(int(i))
			if item.GetUserScore() < -(*userGameScore)[i] { // 金币不足 跑得快分数结算跟庄家没特殊关联
				(*userGameScore)[i] = -item.GetUserScore()
			}
			winuserScore -= (*userGameScore)[i] // --为正
			t.table.SetPassivityNoCard(i)       // 记录输家被关次数
		} else {
			winuser = i
		}
	}
	(*userGameScore)[winuser] = winuserScore
	t.table.SetSpring(winuser) // 记录赢家春天

}

// CheckOutCard 检查出牌
func (t *TableLogic) CheckOutCard(chairID int, outPoker []int32, outType int32) gdy.UserOutPokerCode {
	return gdy.UserOutPokerSuccess
}

// CheckSpecial 特殊牌型
func (t *TableLogic) CheckSpecial() bool {
	return false
}

// NewTableLogic ...
func NewTableLogic(table *gdylogic.Table) *TableLogic {
	return &TableLogic{
		table: table,
	}
}
