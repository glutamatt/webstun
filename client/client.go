package client

import (
	"bufio"
	"bytes"
	"fmt"
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

func ConnectWSAndServe(edge, back string) error {
	u, err := url.ParseRequestURI(edge)
	if err != nil {
		return fmt.Errorf("Error parsing edge %s : %v", edge, err)
	}
	backendURL, err := url.ParseRequestURI(back)
	if err != nil {
		return fmt.Errorf("url.ParseRequestURI back %s err : %v", back, err)
	}
	log.Printf("connecting to %s", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("websocket.DefaultDialer.Dial err : %v", err)
	}
	defer c.Close()
	log.Printf("connected to %s", u.String())

	done := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	responses := make(chan []byte)
	signal.Notify(interrupt, os.Interrupt)

	director := httputil.NewSingleHostReverseProxy(backendURL).Director
	// Wrapping original Director because of https://github.com/golang/go/commit/ae315999c2d5514cec17adbd37cf2438e20cbd12#diff-d863507a61be206d112f6e00e6d812a2R68
	proxy := &httputil.ReverseProxy{Director: func(r *http.Request) {
		director(r)
		r.Host = r.URL.Host
	}}

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
				return fmt.Errorf("c.WriteMessage ERR : %v", err)
			}
		case <-done:
			return nil
		case <-interrupt:
			log.Println("interrupt")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				return fmt.Errorf("c.WriteMessage(websocket.CloseMessage) ERR : %v", err)
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return nil
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
