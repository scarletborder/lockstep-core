package room

import (
	"lockstep-core/src/pkg/lockstep/client"
)

type RoomContextImpl struct {
	room *Room
}

func NewRoomContextImpl(room *Room) *RoomContextImpl {
	return &RoomContextImpl{
		room: room,
	}
}

// Ensure RoomContextImpl implements the world.IRoomContext interface
// by providing the required methods.
func (r *RoomContextImpl) Broadcast(data []byte) {
	if r == nil || r.room == nil {
		return
	}
	// 使用 ClientsContainer 的 BroadcastMessage 方法需要 protobuf 消息，
	// 但是 world.IRoomContext 接口期望传入 []byte，因此直接遍历 Clients 并写入。
	r.room.ClientsContainer.Clients.Range(func(key uint32, client *client.Client) bool {
		if client == nil || client.Session == nil || !client.Session.IsConnected() {
			return true
		}
		client.Write(data)
		return true
	})
}

func (r *RoomContextImpl) SendTo(uid uint32, data []byte) {
	if r == nil || r.room == nil {
		return
	}
	r.room.ClientsContainer.SendMessageToUser(data, uid)
}

func (r *RoomContextImpl) SendToMultiple(uids []uint32, data []byte) {
	if r == nil || r.room == nil {
		return
	}
	for _, uid := range uids {
		r.room.ClientsContainer.SendMessageToUser(data, uid)
	}
}

func (r *RoomContextImpl) GetRoomID() uint32 {
	if r == nil || r.room == nil {
		return 0
	}
	return r.room.ID
}

func (r *RoomContextImpl) GetAllPlayers() []uint32 {
	if r == nil || r.room == nil {
		return nil
	}
	players := make([]uint32, 0, r.room.ClientsContainer.GetPlayerCount())
	r.room.ClientsContainer.Clients.Range(func(key uint32, client *client.Client) bool {
		players = append(players, key)
		return true
	})
	return players
}

func (r *RoomContextImpl) GetNextFrame() uint32 {
	if r == nil || r.room == nil || r.room.SyncData == nil || r.room.SyncData.NextFrameID == nil {
		return 0
	}
	return r.room.SyncData.NextFrameID.Load()
}

func (r *RoomContextImpl) KickPlayer(uid uint32, reason string) {
	if r == nil || r.room == nil {
		return
	}
	// 从 ClientsContainer 中删除并关闭连接
	if client, ok := r.room.ClientsContainer.Clients.Load(uid); ok {
		if client != nil {
			// 关闭会话
			if client.Session != nil && client.Session.IsConnected() {
				client.Session.CloseWithError(0, "you have been kicked: "+reason)
			}
		}
	}
	r.room.ClientsContainer.DelUser(uid)
}

func (r *RoomContextImpl) DestroyRoom() {
	if r == nil || r.room == nil {
		return
	}
	r.room.Destroy()
}
