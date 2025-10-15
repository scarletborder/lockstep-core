package game

import (
	"lockstep-core/src/logic/clients"
	messages "lockstep-core/src/messages/pb"
)

// PlantCard 处理种植卡片请求
func (g *GameLogic) PlantCard(player *clients.Player, request *messages.RequestCardPlant) {
	gridOp := &messages.ResponseGridOperation{
		Uid: uint32(player.GetID()),
		Col: request.Base.Col,
		Row: request.Base.Row,
	}
	plantOp := &messages.ResponseCardPlant{
		Pid:   request.Pid,
		Level: request.Level,
		Cost:  request.Cost,
		Base:  gridOp,
	}

	if request.Base.Col < 0 || request.Base.Col >= MAXCOLS ||
		request.Base.Row < 0 || request.Base.Row >= MAXROWS {
		// 无效的 grid
		return
	}

	// 判断这个 grid 有没有做过操作
	if g.cards[request.Base.Row][request.Base.Col] {
		// true => 有操作
		return
	}

	// 成功
	g.cards[request.Base.Row][request.Base.Col] = true

	g.operationChan <- &messages.InGameOperation{
		Payload: &messages.InGameOperation_CardPlant{
			CardPlant: plantOp,
		},
		ProcessFrameId: request.Base.ProcessFrameId,
	}
}

// RemoveCard 处理移除卡片请求
func (g *GameLogic) RemoveCard(player *clients.Player, request *messages.RequestRemovePlant) {
	gridOp := &messages.ResponseGridOperation{
		Uid: uint32(player.GetID()),
		Col: request.Base.Col,
		Row: request.Base.Row,
	}
	removeOp := &messages.ResponseRemovePlant{
		Pid:  request.Pid,
		Base: gridOp,
	}

	if request.Base.Col < 0 || request.Base.Col >= MAXCOLS ||
		request.Base.Row < 0 || request.Base.Row >= MAXROWS {
		// 铲除失败不会有回复
		return
	}

	// 判断这个 grid 有没有做过操作
	if g.cards[request.Base.Row][request.Base.Col] {
		// true => 有操作
		// 铲除失败不会有回复
		return
	}

	g.cards[request.Base.Row][request.Base.Col] = true
	g.operationChan <- &messages.InGameOperation{
		Payload: &messages.InGameOperation_RemovePlant{
			RemovePlant: removeOp,
		},
		ProcessFrameId: request.Base.ProcessFrameId,
	}
}

// UseStarShards 处理使用星之碎片请求
func (g *GameLogic) UseStarShards(player *clients.Player, request *messages.RequestStarShards) {
	gridOp := &messages.ResponseGridOperation{
		Uid: uint32(player.GetID()),
		Col: request.Base.Col,
		Row: request.Base.Row,
	}
	starshardsOp := &messages.ResponseUseStarShards{
		Pid:  request.Pid,
		Cost: request.Cost,
		Base: gridOp,
	}

	if request.Base.Col < 0 || request.Base.Col >= MAXCOLS ||
		request.Base.Row < 0 || request.Base.Row >= MAXROWS {
		return
	}

	// 判断这个 grid 有没有做过操作
	if g.cards[request.Base.Row][request.Base.Col] {
		// true => 有操作
		return
	}

	// 成功
	g.cards[request.Base.Row][request.Base.Col] = true

	g.operationChan <- &messages.InGameOperation{
		Payload: &messages.InGameOperation_UseStarShards{
			UseStarShards: starshardsOp,
		},
		ProcessFrameId: request.Base.ProcessFrameId,
	}
}
