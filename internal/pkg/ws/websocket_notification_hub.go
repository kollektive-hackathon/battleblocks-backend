package ws

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
)

var singletonMutex sync.Mutex

type WebSocketNotificationHub struct {
	registrationMutex sync.Mutex
	listeners         map[string][]*websocket.Conn
}

func (hub *WebSocketNotificationHub) getGameTopic(gameId string) string {
	return fmt.Sprintf("game/%s", gameId)
}

func (hub *WebSocketNotificationHub) RegisterListener(gameId string, conn *websocket.Conn) {
	hub.registrationMutex.Lock()
	defer hub.registrationMutex.Unlock()

	topic := hub.getGameTopic(gameId)
	if hub.listeners[topic] == nil {
		hub.listeners[topic] = []*websocket.Conn{conn}
		return
	}
	hub.listeners[topic] = append(hub.listeners[topic], conn)
}

func (hub *WebSocketNotificationHub) UnregisterListener(gameId string, conn *websocket.Conn) {
	hub.registrationMutex.Lock()
	defer hub.registrationMutex.Unlock()

	connAddrToClose := conn.RemoteAddr()

	topic := hub.getGameTopic(gameId)
	if len(hub.listeners[topic]) == 1 {
		delete(hub.listeners, topic)
		return
	}

	var indexToDelete int
	for i, listener := range hub.listeners[topic] {
		connAddr := listener.RemoteAddr()
		if connAddr == connAddrToClose {
			indexToDelete = i
			break
		}
	}

	hub.listeners[topic] = append(hub.listeners[topic][:indexToDelete], hub.listeners[topic][indexToDelete+1:]...)
}

func (hub *WebSocketNotificationHub) Publish(gameId string, event any) {
	targetTopic := hub.getGameTopic(gameId)
	for topic := range hub.listeners {
		if topic == targetTopic {
			for _, listener := range hub.listeners[topic] {
				_ = listener.WriteJSON(event)
			}
			break
		}
	}
}

var notificationHubSingleton *WebSocketNotificationHub

func NewNotificationHub() *WebSocketNotificationHub {
	singletonMutex.Lock()
	defer singletonMutex.Unlock()

	if notificationHubSingleton == nil {
		notificationHubSingleton = &WebSocketNotificationHub{
			listeners: make(map[string][]*websocket.Conn),
		}
	}

	return notificationHubSingleton
}
