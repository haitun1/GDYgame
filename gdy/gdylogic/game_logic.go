package gdylogic

import (
	"math/rand"
	"sort"
	"time"

	"mjserver/game/gdy/gdefine"
	"mjserver/utils"
)

var (

	// Card52 去掉两个王
	Card52 = []int32{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, //方块 A - K
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, //梅花 A - K
		0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D, //红桃 A - K
		0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3A, 0x3B, 0x3C, 0x3D, //黑桃 A - K
	}

	// Card108 两副牌
	Card108 = []int32{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, //方块 A - K
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, //梅花 A - K
		0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D, //红桃 A - K
		0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3A, 0x3B, 0x3C, 0x3D, //黑桃 A - K
		0x4E, 0x4F, //大小王
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, //方块 A - K
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, //梅花 A - K
		0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D, //红桃 A - K
		0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3A, 0x3B, 0x3C, 0x3D, //黑桃 A - K
		0x4E, 0x4F, //大小王
	}
	// Card54 一副牌
	Card54 = []int32{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, //方块 A - K
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, //梅花 A - K
		0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D, //红桃 A - K
		0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3A, 0x3B, 0x3C, 0x3D, //黑桃 A - K
		0x4E, 0x4F, //大小王
	}
)

// SortInt32 []int32
type SortInt32 []int32

func (a SortInt32) Len() int      { return len(a) }
func (a SortInt32) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SortInt32) Less(i, j int) bool {

	b := a.LogicValue(a[i])
	c := a.LogicValue(a[j])

	if b == c {
		return b > c
	}
	return b > c
}

// LogicValue real-> virtual
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

// AnalyseResult .
type AnalyseResult struct {
	FiveCount        int32        //五张数目
	FourCount        int32        //四张数目
	ThreeCount       int32        //三张数目
	DoubleCount      int32        //两张数目
	SignedCount      int32        //单张数目
	FiveLogicVolue   [5]int32     //五张列表
	FourLogicVolue   [7]int32     //四张列表
	ThreeLogicVolue  [9]int32     //三张列表
	DoubleLogicVolue [14]int32    //两张列表
	SignedLogicVolue [27]int32    //单张列表
	FiveCardData     [27]int32    //五张列表
	FourCardData     [27]int32    //四张列表
	ThreeCardData    [27]int32    //三张列表
	DoubleCardData   [27]int32    //两张列表
	SignedCardData   [27]int32    //单张数目
	PokerData        [8][27]int32 //扑克数据
	BlockCount       [8]int32
}

// OutPokerResult 出牌结果
type OutPokerResult struct {
	Count         int32
	CanOutPoker   [][]int32
	CanntOutPoker []int32
}

// GameLogic ..
type GameLogic struct {
	log        *utils.Logger //获取全局唯一的日志句柄
	magicPoker int32         //赖子
	cardCount  int32         //牌个数
}

// RandomShuffle 洗牌
func RandomShuffle(src []int32) []int32 {
	dest := make([]int32, len(src))
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	perm := r.Perm(len(src))
	for i, v := range perm {
		dest[v] = src[i]
	}

	return dest
}

// Init 初始化游戏规则
func (g *GameLogic) Init() {
	g.magicPoker = -1

}

// SetMagicPoker 设置赖子
func (g *GameLogic) SetMagicPoker(magic int32) {
	g.magicPoker = magic
}

// GetMagicCount 赖子数量
func (g *GameLogic) GetMagicCount(handCard []int32, handPokerCount int32) int32 {
	magicCount := int32(0)
	if g.magicPoker != -1 {
		for i := int32(0); i < handPokerCount; i++ {
			if g.GetPokerLogicValue(handCard[i]) == g.GetPokerLogicValue(g.magicPoker) {
				magicCount++
			}
		}
	}

	return magicCount
}

// GetPokerValue 返回牌值
func (g *GameLogic) GetPokerValue(poker int32) int32 {
	return poker & 0x0f
}

// GetPokerColor 返回花色
func (g *GameLogic) GetPokerColor(poker int32) int32 {
	return poker & 0xf0
}

// GetPokerLogicValue real -> virtual
func (g *GameLogic) GetPokerLogicValue(cardData int32) int32 {
	if cardData == 0 {
		return 0
	}

	cardValue := cardData & 0x0f

	if cardValue >= 0xe {
		return cardValue + 2
	}

	if cardValue <= 2 {
		return cardValue + 13
	}
	return cardValue
}

// VerifyShamCard 带赖子牌型校验
func (g *GameLogic) VerifyShamCard(poker []int32, realPoker []int32, currentUserMagicCard *int32) bool {
	if g.magicPoker == -1 { // king做赖子
		sort.Sort(SortInt32(poker)) // 逆序排列
		for i := 0; i != len(realPoker); i++ {
			if realPoker[i] > 13 && realPoker[i] < 1 {
				g.log.Errorf("赖子牌型错误：realRoker: %x\n", realPoker[i])
				return false
			}
		}
		kingCount := 0
		for i := 0; i < len(poker); i++ {
			if poker[i] == 0x4f {
				kingCount++
			} else if poker[i] == 0x4e {
				kingCount++
			}
		}
		if kingCount != len(realPoker) {
			g.log.Infof("校验失败kingCount：%d != len(realPoker): %d", kingCount, len(realPoker))
			return false
		}
		*currentUserMagicCard = int32(kingCount) // 获取当前出牌赖子数量
		sort.Sort(SortInt32(realPoker))          // 逆序排列
		j := 0
		for i := 0; i < len(poker) && j != len(realPoker); i++ {
			if poker[i] == 0x4f {
				realPoker[j] &= 0x0f
				g.log.Infof("大王顶替poker(0x4f): %x\n", realPoker[j])
				poker[i] = realPoker[j]
				j++
			} else if poker[i] == 0x4e {
				realPoker[j] &= 0x0f
				g.log.Infof("小王顶替poker(0x4e): %x\n", realPoker[j])
				poker[i] = realPoker[j]
				j++
			}
		}
		sort.Sort(SortInt32(poker)) // 逆序排列
		return true

	}
	return false
}

// CompareCard 对比扑克
func (g *GameLogic) CompareCard(firstList []int32, nextList []int32, firstCount int32, nextCount int32, firstMagicCardNum int32, nextMagicCardNum int32) bool {

	//获取类型
	nextType, _ := g.GetPokerType(nextList, nextCount, nextMagicCardNum)
	firstType, _ := g.GetPokerType(firstList, firstCount, firstMagicCardNum)

	//类型判断
	if firstType == gdy.CtError {
		return false
	}

	//王炸
	if nextType == gdy.CtKingBomb {
		return false
	} else if firstType == gdy.CtKingBomb {
		return true
	}
	//g.log.Infof("firstList: %xv, nextList: %xv", firstList, nextList)
	//g.log.Infof("firstCount: %d, nextCount: %d", firstCount, nextCount)
	//开始对比
	switch firstType {
	case gdy.CtSingle, gdy.CtDouble: // 可能需要改动
		g.log.Infoln("hegdy.CtSingle, gdy.CtDoublere")
		if firstCount != nextCount {
			return false
		}
		if firstType != nextType {
			return false
		}
		nextLogicValue := g.GetPokerLogicValue(nextList[nextCount-1])
		firstLogicValue := g.GetPokerLogicValue(firstList[firstCount-1])
		//	fmt.Printf("firstLogicValue: %d, nextLogicValue: %d", firstLogicValue, nextLogicValue)
		return firstLogicValue-1 == nextLogicValue || (firstLogicValue == 15 && nextLogicValue != 15) // 2(15)可以管一切
	case gdy.CtSingleLine:
		g.log.Infoln("gdy.CtSingleLine")
		if firstCount != nextCount {
			return false
		}
		if firstType != nextType {
			return false
		}
		nextLogicValue := g.GetPokerLogicValue(nextList[nextCount-1])
		firstLogicValue := g.GetPokerLogicValue(firstList[firstCount-1])
		return firstLogicValue-1 == nextLogicValue
	case gdy.CtDoubleLine: // 可能需要改动
		g.log.Infoln("gdy.CtDoubleLine")
		if firstCount != nextCount {
			return false
		}
		if firstType != nextType {
			return false
		}
		nextLogicValue := g.GetPokerLogicValue(nextList[nextCount-1])
		firstLogicValue := g.GetPokerLogicValue(firstList[firstCount-1])
		return firstLogicValue-1 == nextLogicValue
	case gdy.CtSoftThreeBomb, gdy.CtRealThreeBomb, gdy.CtSoftFourBomb, gdy.CtRealFourBomb, gdy.CtSoftFiveBomb:
		g.log.Infoln("gdy.CtSoftThreeBomb, gdy.CtRealThreeBomb, gdy.CtSoftFourBomb, gdy.CtRealFourBomb, gdy.CtSoftFiveBomb")
		if firstType != nextType {
			return firstType > nextType
		}
		nextLogicValue := g.GetPokerLogicValue(nextList[nextCount-1])
		firstLogicValue := g.GetPokerLogicValue(firstList[firstCount-1])
		return firstLogicValue > nextLogicValue

	}
	return false
}

// GetPokerType 这里不考虑双王做赖子的情况
func (g *GameLogic) GetPokerType(pokerData []int32, pokerCount int32, magicCardNum int32) (pokerType int32, shamKingCard int32) {
	sort.Sort(SortInt32(pokerData))
	switch pokerCount {
	case 1: //单牌
		if g.magicPoker == -1 { // king为赖子
			if (pokerData[0] != 0x4e) && (pokerData[0] != 0x4f) {
				return gdy.CtSingle, 0
			}
			return gdy.CtError, 0
		}
		return gdy.CtSingle, 0
	case 2: //对牌
		if g.GetPokerLogicValue(pokerData[0]) == g.GetPokerLogicValue(pokerData[1]) {
			return gdy.CtDouble, 0
		}
		kingCount := 0
		for i := int32(0); i < pokerCount; i++ {
			if pokerData[i] == 0x4e {
				kingCount++
			} else if pokerData[i] == 0x4f {
				kingCount++
			}
		}
		if kingCount == 2 {
			return gdy.CtKingBomb, 0
		}

		return gdy.CtError, 0

	}

	tempPoker := make([]int32, pokerCount)
	copy(tempPoker, pokerData)

	return g.AnalysebPokerType(tempPoker, pokerCount, magicCardNum)

}

// GetPokerTypeEx ?
func (g *GameLogic) GetPokerTypeEx(pokerData []int32, pokerCount int32, magicCardNum int32) (pokerType int32, shamKingCard int32) {
	switch pokerCount {
	case 1: //单牌
		return gdy.CtSingle, 0
	case 2: //对牌这里需要修改赖子判断
		kingCount := 0

		for i := int32(0); i < pokerCount; i++ {
			if pokerData[i] == 0x4e {
				kingCount++

			} else if pokerData[i] == 0x4f {
				kingCount++

			}

		}
		if g.GetPokerLogicValue(pokerData[0]) == g.GetPokerLogicValue(pokerData[1]) {
			return gdy.CtDouble, 0
		} else if kingCount == 2 {
			return gdy.CtKingBomb, 0
		} else if g.GetPokerLogicValue(pokerData[0]) != g.GetPokerLogicValue(pokerData[1]) && kingCount == 1 {
			if g.GetPokerLogicValue(pokerData[0]) < g.GetPokerLogicValue(pokerData[1]) {
				return gdy.CtDouble, g.GetPokerLogicValue(pokerData[0])
			}
		} else {
			return gdy.CtError, 0
		}

	}

	tempPoker := make([]int32, pokerCount)
	copy(tempPoker, pokerData)

	return g.AnalysebPokerType(tempPoker, pokerCount, magicCardNum)

}

// AnalysebPokerType 3牌以上判断，牌型判断都是已经将赖子顶替普通牌型成功的数据
func (g *GameLogic) AnalysebPokerType(pokerData []int32, pokerCount int32, magicCardNum int32) (pokerType int32, shamKingCard int32) {

	var analyseResult AnalyseResult
	g.AnalysebCardData(pokerData, pokerCount, &analyseResult)

	//炸弹 : 该判断中赖子都被替换成对应牌型， 可以用magicCardNum来区分真假炸弹
	if analyseResult.FiveCount == 1 && pokerCount == 5 && magicCardNum != 0 { // 软五炸
		return gdy.CtSoftFiveBomb, 0
	}

	if analyseResult.FourCount == 1 && pokerCount == 4 { // 4炸
		if magicCardNum == 0 {
			return gdy.CtRealFourBomb, 0
		}
		return gdy.CtSoftFourBomb, 0 // 软4炸弹
	}

	if analyseResult.ThreeCount == 1 && pokerCount == 3 {
		if magicCardNum != 0 {
			g.log.Infoln("软三炸")
			return gdy.CtSoftThreeBomb, 0
		}
		g.log.Infoln("真三炸")
		return gdy.CtRealThreeBomb, 0 // 3炸
	}

	// 单连和双连先不考虑双王搭配情况
	//两连判断
	if analyseResult.DoubleCount > 1 {
		//连牌判断
		seriesCard := false
		g.log.Infof("连牌判断analyseResult.DoubleCount:%d", analyseResult.DoubleCount)
		g.log.Infof("analyseResult.DoubleLogicVolue:%v", analyseResult.DoubleLogicVolue)
		i := int32(1)
		logicValue := analyseResult.DoubleLogicVolue[0]
		if logicValue < 15 {
			for ; i < analyseResult.DoubleCount; i++ {

				if analyseResult.DoubleLogicVolue[i] != (logicValue-1) && logicValue != 15 {
					g.log.Infof("连牌判断logicValue:%d,analyseResult.DoubleLogicVolue[i]:%d", logicValue, analyseResult.DoubleLogicVolue[i])
					break
				}
				logicValue = analyseResult.DoubleLogicVolue[i]
			}
		}
		if i == analyseResult.DoubleCount {
			seriesCard = true
		}

		//连对判断
		if seriesCard == true && analyseResult.DoubleCount*2 == pokerCount {
			g.log.Infoln("连牌判断成功")
			return gdy.CtDoubleLine, 0
		}
	}

	//单连判断
	//	fmt.Printf("单连判断\n")
	if analyseResult.SignedCount > 2 {
		g.log.Infoln("开始单连判断")
		//变量定义
		seriesCard := false
		logicValue := g.GetPokerLogicValue(pokerData[0])
		//连牌判断
		if logicValue < 15 {
			i := int32(1)
			for ; i < analyseResult.SignedCount; i++ {
				if g.GetPokerLogicValue(pokerData[i]) != logicValue-1 {
					break
				}
				logicValue = g.GetPokerLogicValue(pokerData[i])
			}

			if i == analyseResult.SignedCount && analyseResult.SignedCount == pokerCount {
				seriesCard = true
			}
		}

		//单连判断
		if seriesCard == true {
			g.log.Infoln("单连判断（无赖子）成功")
			return gdy.CtSingleLine, 0
		}

	}
	return gdy.CtError, 0
}

// AnalysebCardData 组合数据
func (g *GameLogic) AnalysebCardData(cardData []int32, cardCount int32, analyseResult *AnalyseResult) {

	for i := int32(0); i < cardCount; i++ {
		//变量定义
		sameCount := int32(1)
		sameCardData := []int32{cardData[i], 0, 0, 0, 0, 0, 0, 0}
		logicValue := g.GetPokerLogicValue(cardData[i])

		//获取同牌
		for j := i + 1; j < cardCount; j++ {
			//逻辑对比
			if g.GetPokerLogicValue(cardData[j]) != logicValue {
				break
			}

			//设置扑克
			sameCardData[sameCount] = cardData[j]
			sameCount++
		}

		//保存结果
		switch sameCount {
		case 1: //单张
			analyseResult.SignedLogicVolue[analyseResult.SignedCount] = logicValue
			copy(analyseResult.SignedCardData[(analyseResult.SignedCount)*sameCount:], sameCardData)
			analyseResult.SignedCount++
			break

		case 2: //两张
			analyseResult.DoubleLogicVolue[analyseResult.DoubleCount] = logicValue
			copy(analyseResult.DoubleCardData[(analyseResult.DoubleCount)*sameCount:], sameCardData)
			analyseResult.DoubleCount++
			break

		case 3: //三张
			analyseResult.ThreeLogicVolue[analyseResult.ThreeCount] = logicValue
			copy(analyseResult.ThreeCardData[(analyseResult.ThreeCount)*sameCount:], sameCardData)
			analyseResult.ThreeCount++
			break

		case 4: //四张
			analyseResult.ThreeLogicVolue[analyseResult.ThreeCount] = logicValue
			copy(analyseResult.ThreeCardData[(analyseResult.ThreeCount)*sameCount:], sameCardData)
			analyseResult.ThreeCount++

			analyseResult.FourLogicVolue[analyseResult.FourCount] = logicValue
			copy(analyseResult.FourCardData[(analyseResult.FourCount)*sameCount:], sameCardData)
			analyseResult.FourCount++
			break
		case 5: //五张
			analyseResult.ThreeLogicVolue[analyseResult.ThreeCount] = logicValue
			copy(analyseResult.ThreeCardData[(analyseResult.ThreeCount)*sameCount:], sameCardData)
			analyseResult.ThreeCount++

			analyseResult.FourLogicVolue[analyseResult.FourCount] = logicValue
			copy(analyseResult.FourCardData[(analyseResult.FourCount)*sameCount:], sameCardData)
			analyseResult.FourCount++

			analyseResult.FiveLogicVolue[analyseResult.FourCount] = logicValue
			copy(analyseResult.FiveCardData[(analyseResult.FiveCount)*sameCount:], sameCardData)
			analyseResult.FiveCount++
			break

		}

		//设置递增
		i += (sameCount - 1)
	}
}

// SortCardList 排列扑克
func (g *GameLogic) SortCardList(cardData []int32, cardCount int32) {
	//转换数值
	var logicVolue [16]int32
	for i := int32(0); i < cardCount; i++ {
		logicVolue[i] = g.GetPokerLogicValue(cardData[i])
	}

	//排序操作
	sorted := true
	tempData := int32(0)
	last := cardCount - 1
	for {
		sorted = true
		for i := int32(0); i < last; i++ {
			if (logicVolue[i] < logicVolue[i+1]) ||
				((logicVolue[i] == logicVolue[i+1]) && (cardData[i] < cardData[i+1])) {
				//交换位置
				tempData = cardData[i]
				cardData[i] = cardData[i+1]
				cardData[i+1] = tempData
				tempData = logicVolue[i]
				logicVolue[i] = logicVolue[i+1]
				logicVolue[i+1] = tempData
				sorted = false
			}
		}
		last--
		if sorted == true {
			break
		}
	}
}

// RemovePoker 删除扑克
func (g *GameLogic) RemovePoker(removePoker []int32, removeCount int32, pokerData *[]int32, pokerCount int32) bool {

	deleteCount := int32(0)
	tempPokerData := make([]int32, pokerCount)
	if pokerCount > int32(len(tempPokerData)) {
		return false
	}

	copy(tempPokerData[0:], (*pokerData)[0:pokerCount])

	for i := int32(0); i < removeCount; i++ {
		for j := int32(0); j < pokerCount; j++ {
			if removePoker[i] == tempPokerData[j] {
				deleteCount++
				g.log.Infof("找到待删除扑克： %x\n", removePoker[i])
				tempPokerData[j] = 0
				break
			}
		}
	}
	if deleteCount != removeCount {
		return false
	}

	//清理扑克
	pos := int32(0)
	for i := int32(0); i < pokerCount; i++ {
		if tempPokerData[i] != 0 {
			(*pokerData)[pos] = tempPokerData[i]
			pos++
		}
	}
	(*pokerData) = (*pokerData)[:pos] // 删除扑克 1244 delete 2 --> 144
	return true
}

// NewGameLogic new
func NewGameLogic() *GameLogic {
	gameLogic := GameLogic{
		log: utils.NewLogger(),
	}

	gameLogic.Init()

	return &gameLogic
}
