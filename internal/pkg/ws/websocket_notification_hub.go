package ws

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

var singletonMutex sync.Mutex

type WebSocketNotificationHub struct {
	registrationMutex sync.Mutex
	listeners         map[string][]*websocket.Conn
}

func (hub *WebSocketNotificationHub) RegisterListener(topic string, conn *websocket.Conn) {
	hub.registrationMutex.Lock()
	defer hub.registrationMutex.Unlock()

	if hub.listeners[topic] == nil {
		hub.listeners[topic] = []*websocket.Conn{conn}
		return
	}
	hub.listeners[topic] = append(hub.listeners[topic], conn)
}

func (hub *WebSocketNotificationHub) UnregisterListener(topic string, conn *websocket.Conn) {
	hub.registrationMutex.Lock()
	defer hub.registrationMutex.Unlock()

	if conn == nil {
		return
	}
	connAddrToClose := conn.RemoteAddr()

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

func (hub *WebSocketNotificationHub) Publish(targetTopic string, event any) {
	log.Info().Interface("targetTopic", targetTopic).Msg("[WEBSOCKET] Publishing to websocet topic")
	for topic := range hub.listeners {
		if topic == targetTopic {
			for _, listener := range hub.listeners[topic] {
				err := listener.WriteJSON(event)
				if err != nil {
					log.Warn().Msg("[WEBSOCKET] Error writing json to connection")
				}
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
