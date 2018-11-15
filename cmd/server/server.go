package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/gorilla/websocket"
)

func handler(calls chan<- HttpCall) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ch := make(chan *http.Response)
		select {
		case calls <- HttpCall{req: r, resp: ch}:
			fmt.Fprintf(w, "Req pushed to httpCalls chan")
			resp := <-ch
			io.Copy(w, resp.Body)
		default:
			fmt.Fprintf(w, "ERROR: Unable to push Req to httpCalls chan")
		}
	}
}

type HttpCall struct {
	req  *http.Request
	resp chan<- *http.Response
}

func errorResponse(err error) *http.Response {
	return &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       ioutil.NopCloser(strings.NewReader(err.Error())),
	}
}

func main() {
	var httpCalls = make(chan HttpCall)
	var upgrader = websocket.Upgrader{}

	http.HandleFunc("/_ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error on upgrader.Upgrade: %v\n", err)
			return
		}
		log.Println("backend connected via websocket: ready to accept http calls")

		defer conn.Close()

		go func() {
			for {
				messageType, message, err := conn.ReadMessage()
				if err != nil {
					log.Println("read err:", err)
					return
				}
				if messageType == websocket.TextMessage {
					log.Println(string(message))
				}
			}
		}()

		for {
			log.Println("Waiting for http call...")
			select {
			case httpCall := <-httpCalls:
				log.Println("Incomming call")
				req, err := httputil.DumpRequest(httpCall.req, true)
				if err != nil {
					httpCall.resp <- errorResponse(fmt.Errorf("httputil.DumpRequest error: %v", err))
					continue
				}
				data := append([]byte(fmt.Sprintf("%x\n", sha1.Sum(req))), req...)
				if err = conn.WriteMessage(websocket.TextMessage, data); err != nil {
					httpCall.resp <- errorResponse(fmt.Errorf("conn.WriteMessage error (closing backend websocket connection): %v", err))
					return
				}
				log.Println("conn.WriteMessage OK")
				httpCall.resp <- errorResponse(fmt.Errorf("Yeah the req has been pushed on WS"))
			}
		}
	})

	http.HandleFunc("/", handler(httpCalls))
	log.Println("Let's go !")
	log.Fatal(http.ListenAndServe(":3001", nil))
}
