package game

import (
	messages "lockstep-core/src/messages/pb"
)

const (
	MAXROWS  = 6
	MAXCOLS  = 9
	MAXCARDS = 8
)

// GameLogic 游戏的纯逻辑
type GameLogic struct {
	operationChan chan<- *messages.InGameOperation
	cards         [MAXROWS][MAXCOLS]bool
}

// NewGameLogic 创建新的游戏逻辑实例
func NewGameLogic(operationChan chan<- *messages.InGameOperation) *GameLogic {
	return &GameLogic{
		operationChan: operationChan,
	}
}

// Reset 重置游戏逻辑状态
func (g *GameLogic) Reset() {
	for i := 0; i < MAXROWS; i++ {
		for j := 0; j < MAXCOLS; j++ {
			g.cards[i][j] = false
		}
	}
}
