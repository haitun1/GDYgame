package gdy

import (
	"fmt"
	"mjserver/gameprivate/gdy/gdylogic"
	"sort"
	"testing"
)

func Test_CompareCard(t *testing.T) {
	gameLogic := gdylogic.NewGameLogic()
	var a, b, d []int32
	a = append(a, 0x24, 0x04, 0x05, 0x15)
	b = append(b, 0x05, 0x25, 0x16, 0x26)
	//	gameLogic.CompareCard(b, a, 4, 4)
	c := gameLogic.CompareCard(b, a, 4, 4)
	fmt.Println(c)
	d = append(d, 0x1b, 0xc, 0xc, 0xb, 0x1a, 0xa, 0x29, 0x9, 0x8, 0x38)
	sort.Sort(SortInt32(d))
	//	gameLogic.GetPokerType(d, 10)
}
