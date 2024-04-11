package system

type EventMessage struct {
	ChannelID string
	Text      string
}

type EventManager struct {
	channels map[string]chan EventMessage
}

func (em *EventManager) EmitToChannel(channelID string, event EventMessage) {
	em.channels[channelID] <- event
}

func (em *EventManager) Emit(event EventMessage) {
	for _, channel := range em.channels {
		channel <- event
	}
}

func (em *EventManager) DeleteChannel(channelID string) {
	delete(em.channels, channelID)
}

func (em *EventManager) HasChannel(channelID string) bool {
	return em.channels[channelID] != nil
}

func (em *EventManager) GetChannel(channelID string) chan EventMessage {
	return em.channels[channelID]
}

func (em *EventManager) RegisterChannel(channelID string) chan EventMessage {
	em.channels[channelID] = make(chan EventMessage)
	return em.channels[channelID]
}

func NewEventManager() *EventManager {
	return &EventManager{
		channels: make(map[string]chan EventMessage),
	}
}

var EventManagerInstance = NewEventManager()
