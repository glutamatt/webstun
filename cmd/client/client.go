package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

func readLine(reader *bufio.Reader) (string, error) {
	isPrefix := true
	line := bytes.NewBuffer([]byte{})
	var linePart []byte
	var err error
	for isPrefix {
		linePart, isPrefix, err = reader.ReadLine()
		if err != nil {
			return "", fmt.Errorf("Error on reader.ReadLine: %v", err)
		}
		n, err := line.Write(linePart)
		if err != nil {
			return "", fmt.Errorf("Error on line.Write: %v", err)
		}
		if n != len(linePart) {
			return "", fmt.Errorf("Error on line.Write written bytes len: %v", err)
		}
	}

	return line.String(), nil
}

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
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		defer close(done)
		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			if messageType == websocket.TextMessage {
				reader := bufio.NewReader(bytes.NewReader(message))
				hash, err := readLine(reader)
				if err != nil {
					log.Printf("HASH: %s \n err: %v\n", hash, err)
					return
				}
				/*
					MESSAGE, _ := ioutil.ReadAll(reader)
					log.Println("MESSAGE ", string(MESSAGE))
				*/

				req, err := http.ReadRequest(reader)
				if err != nil {
					log.Println("http.ReadRequest ERR :", err)
					return
				}
				url := "http://grafana.deez.re" + req.URL.String()

				/*
					body, _ := ioutil.ReadAll(req.Body)
					log.Println("BODY ", string(body))
				*/

				forgedReq, err := http.NewRequest(req.Method, url, req.Body)
				if err != nil {
					log.Println("http.NewRequest ERR :", err)
					return
				}
				for k, v := range req.Header {
					if k != "Host" {
						forgedReq.Header[k] = v
					}
				}

				/*
					body, _ := ioutil.ReadAll(forgedReq.Body)
					log.Println("BODY ", string(body))
				*/

				debug, _ := httputil.DumpRequest(forgedReq, true)
				log.Println("debug ", string(debug))

				res, err := http.DefaultClient.Do(forgedReq)
				if err != nil {
					log.Println("http.DefaultClient.Do ERR :", err)
					return
				}
				dump, err := httputil.DumpResponse(res, true)
				if err != nil {
					log.Println("httputil.DumpResponse ERR :", err)
					return
				}
				if err := c.WriteMessage(websocket.TextMessage, append([]byte(hash+"\n"), dump...)); err != nil {
					log.Println("c.WriteMessage ERR :", err)
					return
				}
			}
		}
	}()

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

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
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
