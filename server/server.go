package server

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//ListenAndServe set the handlers and start the http server
func ListenAndServe(addr string) error {
	var httpCalls = make(chan httpCall)
	http.HandleFunc("/_ws", websocketHandler(httpCalls))
	http.HandleFunc("/", handler(httpCalls))
	return http.ListenAndServe(addr, nil)
}

func handler(calls chan<- httpCall) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ch := make(chan *http.Response)
		defer close(ch)
		timer := time.NewTimer(1 * time.Second)
		defer timer.Stop()
		select {
		case calls <- httpCall{req: r, resp: ch}:
			resp := <-ch
			for k, vv := range resp.Header {
				for _, v := range vv {
					w.Header().Add(k, v)
				}
			}
			w.WriteHeader(resp.StatusCode)
			io.Copy(w, resp.Body)
		case <-timer.C:
			fmt.Fprintf(w, "ERROR timeout: Unable to push Req to httpCalls chan")
		}
	}
}

type httpCall struct {
	req  *http.Request
	resp chan<- *http.Response
}
type dispatcher struct {
	lock  sync.Mutex
	pipes map[string]chan<- *http.Response
}

func (d *dispatcher) Handle(hash string, resp chan<- *http.Response) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.pipes[hash] = resp
}

func (d *dispatcher) Serve(hash string, resp *http.Response) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if call, exist := d.pipes[hash]; exist {
		call <- resp
		delete(d.pipes, hash)
		return
	}
	log.Println("Serve", hash, "doesn't exist !!!")
}

func websocketHandler(calls <-chan httpCall) func(http.ResponseWriter, *http.Request) {
	var upgrader = websocket.Upgrader{}
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error on upgrader.Upgrade: %v\n", err)
			return
		}
		log.Println("proxy connected via websocket: ready to accept http calls")

		defer conn.Close()

		disp := &dispatcher{pipes: map[string]chan<- *http.Response{}}

		go func() {
			for {
				messageType, message, err := conn.ReadMessage()
				if err != nil {
					log.Println("read err:", err)
					return
				}
				if messageType == websocket.TextMessage {
					reader := bufio.NewReader(bytes.NewReader(message))
					hash, err := reader.ReadString('\n')
					if err != nil {
						log.Printf("HASH: %s \n err: %v\n", hash, err)
						continue
					}
					resp, err := http.ReadResponse(reader, nil)
					if err != nil {
						log.Printf("http.ReadResponse err: %v", err)
						continue
					}
					go disp.Serve(strings.TrimRight(hash, "\n"), resp)
				}
			}
		}()

		for {
			select {
			case httpCall := <-calls:
				req, err := httputil.DumpRequest(httpCall.req, true)
				if err != nil {
					httpCall.resp <- errorResponse(fmt.Errorf("httputil.DumpRequest error: %v", err))
					continue
				}
				hash := fmt.Sprintf("%x", sha1.Sum(req))
				data := append([]byte(fmt.Sprintf("%s\n", hash)), req...)
				if err = conn.WriteMessage(websocket.TextMessage, data); err != nil {
					httpCall.resp <- errorResponse(fmt.Errorf("conn.WriteMessage error (closing proxy websocket connection): %v", err))
					return
				}
				go disp.Handle(hash, httpCall.resp)
			}
		}
	}
}
func errorResponse(err error) *http.Response {
	return &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       ioutil.NopCloser(strings.NewReader(err.Error())),
	}
}
