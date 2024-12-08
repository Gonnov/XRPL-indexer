package main

import (
    "log"
    "net/url"
    "os"
    "os/signal"
    "time"

    "github.com/gorilla/websocket"
)

func setupInterruptHandler() chan os.Signal {
    interrupt := make(chan os.Signal, 1)
    signal.Notify(interrupt, os.Interrupt)
    return interrupt
}

func connectWebSocket(u url.URL) (*websocket.Conn, chan struct{}) {
    c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
    if err != nil {
        log.Fatal("dial:", err)
    }
    done := make(chan struct{})
    return c, done
}

func readMessages(c *websocket.Conn, done chan struct{}) {
    defer close(done)
    for {
        _, message, err := c.ReadMessage()
        if err != nil {
            log.Println("read:", err)
            return
        }
        log.Printf("recv: %s", message)
    }
}

func writeMessages(c *websocket.Conn, done chan struct{}, interrupt chan os.Signal) {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-done:
            return
        case t := <-ticker.C:
            err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
            if err != nil {
                log.Println("write:", err)
                return
            }
        case <-interrupt:
            log.Println("interrupt")
            closeConnection(c, done)
            return
        }
    }
}

func closeConnection(c *websocket.Conn, done chan struct{}) {
    err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
    if err != nil {
        log.Println("write close:", err)
        return
    }
    select {
    case <-done:
    case <-time.After(time.Second):
    }
}

func WebsocketManagement() {
    interrupt := setupInterruptHandler()

    // Choose one of the following URLs
    u := url.URL{Scheme: "wss", Host: "xrplcluster.com", Path: "/"}

    log.Printf("connecting to %s", u.String())

    c, done := connectWebSocket(u)
    defer c.Close()

    go readMessages(c, done)

    writeMessages(c, done, interrupt)
}

