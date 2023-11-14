// Nebulant
// Copyright (C) 2020  Develatio Technologies S.L.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//
// The code of this file was bassed on WebSocket Chat example from
// gorilla websocket lib: https://github.com/gorilla/websocket/blob/master/examples/chat/client.go

package cast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ExecutionsRegistry struct {
	ByClientUUID map[string]map[string]bool
	mu           sync.Mutex
}

func (e *ExecutionsRegistry) GetByClient(clientUUID string) (map[string]bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if _, exists := e.ByClientUUID[clientUUID]; exists {
		return nil, fmt.Errorf("not found")
	}
	return e.ByClientUUID[clientUUID], nil
}

var ER = &ExecutionsRegistry{
	ByClientUUID: make(map[string]map[string]bool),
}

// WSocketLogger is a middleman between the websocket connection and the SBus.
type WSocketLogger struct {
	conn  *websocket.Conn
	fLink *BusConsumerLink
	mu    sync.Mutex
}

type clientMsg struct {
	Cmd   string `json:"cmd" validate:"required"`
	Param string `json:"param" validate:"required"`
	Ok    bool   `json:"ok"`
}

func (c *WSocketLogger) readWebSocket() {
	defer func() {
		SBus.disconnect <- c.fLink
		c.conn.Close() // #nosec G104 -- Unhandle is OK here
		SBus.castWaiter.Done()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("error: %v", err)
				fmt.Println(err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, []byte{'\n'}, []byte{' '}, -1))

		// Read client msg
		clmsg := &clientMsg{}
		readErr := json.Unmarshal(message, clmsg)
		if readErr != nil {
			log.Printf("WSocket 2 err: %v", readErr)
			continue
		}

		clmsg.Ok = false
		// Handle client msg
		if clmsg.Cmd == "join" {
			c.joinExecution(clmsg.Param)
			clmsg.Ok = true
		}

		// Write back to client
		writeErr := c.lockedWriteToWS(clmsg)
		if writeErr != nil {
			log.Printf("WSocket 3 err: %v", writeErr)
			continue
		}
	}
}

func (c *WSocketLogger) lockedWriteToWS(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteJSON(v)
}

func (c *WSocketLogger) lockedSetWriteDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.SetWriteDeadline(t)
}

func (c *WSocketLogger) joinExecution(remoteExecutionUUID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ER.mu.Lock()
	defer ER.mu.Unlock()
	if _, exists := ER.ByClientUUID[c.fLink.ClientUUID]; !exists {
		ER.ByClientUUID[c.fLink.ClientUUID] = make(map[string]bool)
	}
	ER.ByClientUUID[c.fLink.ClientUUID][remoteExecutionUUID] = true
}

func (c *WSocketLogger) canReadExecution(remoteExecutionUUID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	ER.mu.Lock()
	defer ER.mu.Unlock()
	if _, exists := ER.ByClientUUID[c.fLink.ClientUUID]; !exists {
		// no client
		return false
	}
	if active, exists := ER.ByClientUUID[c.fLink.ClientUUID][remoteExecutionUUID]; exists {
		// client into this channel
		if !active {
			// channel is not active, skip log
			return false
		}
		// channel active, log
	} else {
		// client out of this channel, skip log
		return false
	}
	return true
}

// readCastBus read log pipe and write back to websocket
func (c *WSocketLogger) readCastBus() {
	ticker := time.NewTicker(((60 * time.Second) * 9) / 10)

	defer func() {
		ticker.Stop()
		c.conn.Close() // #nosec G104 -- Unhandle is OK here
		SBus.castWaiter.Done()
	}()
	power := true

	for {
		if !power && len(c.fLink.LogChan) <= 0 {
			return
		}
		select {
		case fback, ok := <-c.fLink.CommonChan:
			// No timeout for msg
			if err := c.lockedSetWriteDeadline(time.Time{}); err != nil {
				log.Printf("WSocket 6b err: %v", err)
				return
			}
			if !ok {
				// if there is no more values to receive and channel is closed
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					log.Printf("WSocket 7a err: %v", err)
				}
				return
			}
			if fback.TypeID == BusDataTypeEOF {
				// EOF feedback, close logger
				// entering shutdown mode
				power = false
			}

			if fback.EventID != nil && *fback.EventID == EventRegisteredManager {
				// this event is for httpd, ignore
				continue
			}

			if executionUUID, exists := fback.Extra["join"]; exists {
				c.joinExecution(executionUUID.(string))
			}

			if fback.ExecutionUUID == nil {
				// no remote uuid, skip data
				continue
			}
			if !c.canReadExecution(*fback.ExecutionUUID) {
				continue
			}
			err := c.lockedWriteToWS(fback)
			if err != nil {
				nerr, ok := err.(net.Error)
				if !ok {
					// no net error
					continue
				}
				if !nerr.Timeout() {
					// no timeout err
					continue
				}
				// TODO: here the socket may be broken
				log.Printf("WSocket 8a err: %v", err)
				return
			}
		case fback, ok := <-c.fLink.LogChan:
			// No timeout for msg
			if err := c.lockedSetWriteDeadline(time.Time{}); err != nil {
				log.Printf("WSocket 6b err: %v", err)
				return
			}
			if !ok {
				// if there is no more values to receive and channel is closed
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					log.Printf("WSocket 7b err: %v", err)
				}
				return
			}
			if fback.ExecutionUUID == nil {
				// no remote uuid, skip log
				continue
			}
			if !c.canReadExecution(*fback.ExecutionUUID) {
				continue
			}
			err := c.lockedWriteToWS(fback)
			if err != nil {
				nerr, ok := err.(net.Error)
				if !ok {
					// no net error
					continue
				}
				if !nerr.Timeout() {
					// no timeout err
					continue
				}
				// TODO: here the socket may be broken
				log.Printf("WSocket 8b err: %v", err)
				return
			}

		case <-ticker.C:
			// Timeout 10 for Pong
			if err := c.lockedSetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				log.Printf("WSocket 9 err: %v", err)
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// NewWebSocketLogger handles websocket requests from the peer.
func NewWebSocketLogger(conn *websocket.Conn, clientUUID string) {
	fLink := &BusConsumerLink{
		Name:            "WebSocket",
		ClientUUID:      clientUUID,
		LogChan:         make(chan *BusData, 100),
		CommonChan:      make(chan *BusData, 100),
		AllowEventData:  true,
		AllowStatusData: true,
	}
	logger := &WSocketLogger{conn: conn, fLink: fLink}
	select {
	case SBus.connect <- fLink:
	default:
		// Hey developer!,  what a wonderful day!
	}
	SBus.castWaiter.Add(2) // one for log pipe reader and another for ws reader
	go logger.readCastBus()
	go logger.readWebSocket()
	LogInfo("New WS client connected", nil)
}
