/*
 Copyright 2015 Bluek404

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package nwui

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

// 用于生成随机数
var r = rand.New(rand.NewSource(time.Now().UnixNano()))

// 生成控件ID
func NewControlID() string { return strconv.FormatInt(r.Int63(), 36) }

// 创建新窗口
func NewWindow(title string, x, y uint) Window {
	w := Window{
		title: title,
		exit:  make(chan bool),
	}
	w.size.x = x
	w.size.y = y
	return w
}

type Window struct {
	title string
	theme Theme
	size  struct {
		x uint
		y uint
	}
	controls []Control
	exit     chan bool
	onExit   func() bool
}

func (w *Window) UseTheme(t Theme) {
	w.theme = t
}

func (w *Window) Show(con ...Control) {
	go func() {
		var (
			allEvents map[string]func(v string)
			message   = make(chan []byte)
			upgrader  = websocket.Upgrader{
				ReadBufferSize:  1024,
				WriteBufferSize: 1024,
			}
			sendFunc = func(f, v string) {
				msg, err := json.Marshal(&wsMsg{
					Type:  f,
					Value: v,
				})
				if err != nil {
					log.Println(err)
				}
				message <- msg
			}
		)

		for _, v := range con {
			v.setSendFunc(sendFunc)
			for e, f := range v.getEvents() {
				allEvents[e] = f
			}
		}

		http.HandleFunc("/ws", func(rw http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(rw, r, nil)
			if err != nil {
				log.Println(err)
				return
			}

			// 用于接收事件
			go func() {
				for {
					messageType, p, err := conn.ReadMessage()
					if err != nil {
						log.Println(err)
					}

					if messageType == websocket.TextMessage {
						var msg wsMsg
						err = json.Unmarshal(p, &msg)
						if err != nil {
							log.Println(err)
						}
						// 判断事件是否为内置事件
						// 进行相应处理
						switch msg.Type {
						case "exit":
							if w.onExit != nil {
								if w.onExit() {
									w.exit <- true
								}
							}
						}
						// 执行事件所绑定的函数
						f, ok := allEvents[msg.Type]
						if ok {
							f(msg.Value)
						} else {
							log.Println("unfind event:", msg.Type)
						}
					}
				}
			}()

			for {
				// 发送消息给前端
				msg := <-message
				err = conn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					log.Println(err)
				}
			}
		})
		http.HandleFunc("/", func(http.ResponseWriter, *http.Request) {
		})
		err := http.ListenAndServe("localhost:8080", nil)
		if err != nil {
			panic(err)
		}
	}()
	<-w.exit
}

func (w *Window) OnExit(f func() bool) {
	w.onExit = f
}

func NewButton(text string) Button {
	return Button{
		id:     NewControlID(),
		text:   text,
		events: make(map[string]func(v string)),
	}
}

// 按钮控件
type Button struct {
	id     string
	text   string
	events map[string]func(v string)
	send   func(f, v string)
}

func (b *Button) getEvents() map[string]func(v string) {
	return b.events
}

func (b *Button) setSendFunc(f func(f, v string)) {
	b.send = f
}

// 按钮被点击触发的事件
func (b *Button) OnClick(f func()) {
	// 当收到b.id+"OnClick"事件时
	// 会执行函数f
	b.events[b.id+"OnClick"] = func(v string) {
		f()
	}
}

// 设置按钮文字
func (b *Button) SetText(text string) {
	// b.send会去调用javascript里名为
	// b.id+"SetText"的函数
	// 并传入参数text
	b.send(b.id+"SetText", text)
	b.text = text
}

// 获取按钮文字
func (b *Button) GetText() string {
	return b.text
}

type Theme struct {
	CSS        string
	JavaScript string
}

type Control interface {
	getEvents() map[string]func(v string)
	setSendFunc(func(f, v string))
}

type wsMsg struct {
	Type  string
	Value string
}
