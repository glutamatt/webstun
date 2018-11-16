package main

import (
	"bufio"
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	u := url.URL{Scheme: "ws", Host: "localhost:3000", Path: "/_ws"}
	log.Printf("connecting to %s", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
	log.Printf("connected to %s", u.String())

	done := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	responses := make(chan []byte)
	signal.Notify(interrupt, os.Interrupt)

	back := "http://192.168.0.18:3000"
	backendURL, err := url.ParseRequestURI(back)
	if err != nil {
		log.Fatal("url.ParseRequestURI %s err : %v", back, err)
	}
	proxy := httputil.NewSingleHostReverseProxy(backendURL)

	go func() {
		defer close(done)
		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			if messageType == websocket.CloseMessage {
				log.Println("Close message from the server", string(message))
				return
			}
			if messageType == websocket.TextMessage {
				go handleRequest(message, responses, proxy)
			}
		}
	}()

	for {
		select {
		case resp := <-responses:
			if err := c.WriteMessage(websocket.TextMessage, resp); err != nil {
				log.Println("c.WriteMessage ERR :", err)
				return
			}
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func handleRequest(message []byte, responses chan []byte, proxy *httputil.ReverseProxy) {
	reader := bufio.NewReader(bytes.NewReader(message))
	hash, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("HASH: %s \n err: %v\n", hash, err)
		return
	}
	req, err := http.ReadRequest(reader)
	if err != nil {
		log.Println("http.ReadRequest ERR :", err)
		return
	}

	rw := httptest.NewRecorder()
	proxy.ServeHTTP(rw, req)
	resp, err := httputil.DumpResponse(rw.Result(), true)

	if err != nil {
		log.Println("httputil.DumpResponse ERR :", err)
		return
	}

	responses <- append([]byte(hash), resp...)
}
