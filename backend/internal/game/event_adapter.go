// backend/internal/game/event_adapter.go

package game

import "sync"

type EventChannelAdapter struct {
	gameEventChan    <-chan GameEvent
	adaptedEventChan chan GameEvent
	stopChan         chan struct{}
	stopOnce         sync.Once
}

func NewEventChannelAdapter(gameEventChan <-chan GameEvent) *EventChannelAdapter {
	adapter := &EventChannelAdapter{
		gameEventChan:    gameEventChan,
		adaptedEventChan: make(chan GameEvent, 100),
		stopChan:         make(chan struct{}),
	}
	go adapter.run()
	return adapter
}

func (a *EventChannelAdapter) GetEventChannel() <-chan GameEvent {
	return a.adaptedEventChan
}

func (a *EventChannelAdapter) Stop() {
	a.stopOnce.Do(func() { close(a.stopChan) })
}

func (a *EventChannelAdapter) run() {
	for {
		select {
		case event, ok := <-a.gameEventChan:
			if !ok {
				close(a.adaptedEventChan)
				return
			}
			select {
			case a.adaptedEventChan <- event:
			case <-a.stopChan:
				return
			}
		case <-a.stopChan:
			close(a.adaptedEventChan)
			return
		}
	}
}
