package main

import (
    "errors"
    "net/http"
    "strings"
    "sync"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
)

const (
    messageMaxBytes = 10000
    wsReadWait = 75 * time.Second
    wsWriteWait = 10 * time.Second
    wsPingEvery = 25 * time.Second
)

type safeSocket struct {
    mu sync.Mutex
    conn *websocket.Conn
}

func (s *safeSocket) write(v any) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    _ = s.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
    return s.conn.WriteJSON(v)
}

func validMessageBody(body string) error {
    if strings.TrimSpace(body) == "" { return errors.New("message body is required") }
    if len(body) > messageMaxBytes { return errors.New("message body is too large") }
    return nil
}

func websocketOriginAllowed(c *gin.Context, frontend string) bool {
    if frontend == "" { return true }
    origin := c.GetHeader("Origin")
    return origin == "" || origin == frontend
}

func prepareSocket(ws *websocket.Conn) {
    ws.SetReadLimit(1 << 20)
    _ = ws.SetReadDeadline(time.Now().Add(wsReadWait))
    ws.SetPongHandler(func(string) error {
        return ws.SetReadDeadline(time.Now().Add(wsReadWait))
    })
}

func rejectWebSocketOrigin(c *gin.Context, frontend string) bool {
    if !websocketOriginAllowed(c, frontend) {
        c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error":"origin not allowed"})
        return true
    }
    return false
}
