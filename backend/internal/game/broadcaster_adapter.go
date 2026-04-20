// backend/internal/game/broadcaster_adapter.go

package game

type BroadcasterAdapter struct {
	manager GameManagerInterface
	adapter *EventChannelAdapter
}

func NewBroadcasterAdapter(
	manager GameManagerInterface,
	) *BroadcasterAdapter {
	adapter := NewEventChannelAdapter(
		manager.GetEventChannel(),
	)
	return &BroadcasterAdapter{
		manager: manager,
		adapter: adapter,
	}
}

func (ba *BroadcasterAdapter) GetEventChannel() <-chan GameEvent {
	return ba.adapter.GetEventChannel()
}

func (ba *BroadcasterAdapter) Stop() {
	ba.adapter.Stop()
}
