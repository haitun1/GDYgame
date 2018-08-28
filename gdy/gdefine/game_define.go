package gdy

const (
	// PyOperate10 操作10秒
	PyOperate10 int64 = 0x00000001
	// PyOperate15 操作15秒
	PyOperate15 int64 = 0x00000002
	// PyOperateInfinite 操作无限制
	PyOperateInfinite int64 = 0x00000004
	// PyWinReplenishPoke 赢家补牌
	PyWinReplenishPoke int64 = 0x00000008
	// PyAllReplenishPoke 全部补牌
	PyAllReplenishPoke int64 = 0x00000010
	// PyMust 必须管
	PyMust int64 = 0x00000020
	// PyMaybe 可不管
	PyMaybe int64 = 0x00000040
	// PyKingDouble 余牌有王翻倍
	PyKingDouble int64 = 0x00000080
	// Py2Double 余牌有2翻倍
	Py2Double int64 = 0x00000100
	// PyOwerBanker 赢家上庄
	PyOwerBanker int64 = 0x00000200
	// PyRandBanker 随机庄家
	PyRandBanker int64 = 0x00000400
)

const (
	// CtError 错误
	CtError int32 = 0
	// CtSingle 单牌
	CtSingle int32 = 1
	// CtDouble 单对
	CtDouble int32 = 2
	// CtSingleLine 顺子（3张以上）
	CtSingleLine int32 = 3
	// CtDoubleLine 连对 （两对以上）
	CtDoubleLine int32 = 4
	// CtSoftThreeBomb 软三张炸弹
	CtSoftThreeBomb int32 = 5
	// CtRealThreeBomb 三张炸弹
	CtRealThreeBomb int32 = 6
	// CtSoftFourBomb 软四张炸弹
	CtSoftFourBomb int32 = 7
	// CtRealFourBomb 四张炸弹
	CtRealFourBomb int32 = 8
	// CtSoftFiveBomb 软五张炸弹
	CtSoftFiveBomb int32 = 9
	// CtKingBomb 王炸
	CtKingBomb int32 = 10
)

const (
	// DefaultTimer 默认定时器
	DefaultTimer = iota
)

// TimeOffset 时间偏移
const TimeOffset = 5

//内部维护游戏状态
type GameStatusType int

const (
	// StatusFree 空闲场景
	StatusFree GameStatusType = 1
	// StatusPlay 游戏场景
	StatusPlay GameStatusType = 2
)

// UserOutPokerCode 状态码
type UserOutPokerCode int32

const (
	// UserOutPokerSuccess 成功
	UserOutPokerSuccess UserOutPokerCode = 0
	// UserOutPokerError 出牌错误
	UserOutPokerError UserOutPokerCode = 1
	// UserOutPokerMustMode 该模式下能管必出
	UserOutPokerMustMode UserOutPokerCode = 2
	// UserOutPokerStatus 不是出牌状态
	UserOutPokerStatus UserOutPokerCode = 3
	// UserOutPokerCurrent 当前玩家不是自己
	UserOutPokerCurrent UserOutPokerCode = 4
	// UserOutPokerNotExist  当前牌不存在
	UserOutPokerNotExist UserOutPokerCode = 5
)

// GameTime 游戏定时器
type GameTime int

const (
	// GAMETIMERPLAY10 定时10S
	GAMETIMERPLAY10 GameTime = 10
	// GAMETIMERPLAY15 定时15S
	GAMETIMERPLAY15 GameTime = 15
)

const (
	MAIN_GAME_ID = 309 //游戏内部的主命令id
)

//S->C
const (
	SUB_S_NotifyGameStart        = 100 // 游戏开始
	SUB_S_NotifyOutCard          = 101
	SUB_S_BroadcastOutCard       = 102
	SUB_S_BroadcastGameEnd       = 103
	SUB_S_BroadcastGameOver      = 104
	SUB_S_BroadcastSceneGameFree = 105
	SUB_S_BroadcastSceneGamePlay = 106
	SUB_S_BroadcastDeal          = 107
	SUB_S_NotifyChooseOutCard    = 108
)

//C->S
const (
	SUB_C_OutPoker = 1
)
