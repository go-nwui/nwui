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
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

// 用于生成随机数
var r = rand.New(rand.NewSource(time.Now().UnixNano()))

var temp = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>%v</title>
<style>%v</style>
</head>
<body>
%v
<script>
var wsURL = "ws://%v/ws";
var socket = new WebSocket(wsURL);
socket.onmessage = function(evt) {
	var data = JSON.parse(evt.data);
	alert(data["type"]);
	alert(data["value"]);
};
%v
</script>
</body>
</html>`

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
			html      string
			js        string = w.theme.JavaScript
			allEvents        = make(map[string]func(v string))
			message          = make(chan []byte)
			upgrader         = websocket.Upgrader{
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
			html += v.genHTML()
			js += v.genJavaScript()
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
						return
					}

					if messageType == websocket.TextMessage {
						var msg wsMsg
						err = json.Unmarshal(p, &msg)
						if err != nil {
							log.Println(err)
							return
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
		http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(rw, temp, w.title, w.theme.CSS, html, "127.0.0.1:8080", js)
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
	return &button{
		id:     NewControlID(),
		text:   text,
		events: make(map[string]func(v string)),
	}
}

type Button interface {
	Control
	OnClick(f func())
	SetText(text string)
	GetText() string
}

// 按钮控件
type button struct {
	id         string
	text       string
	javascript string
	events     map[string]func(v string)
	send       func(f, v string)
}

func (b *button) getEvents() map[string]func(v string) {
	return b.events
}

func (b *button) setSendFunc(f func(f, v string)) {
	b.send = f
}

func (b *button) genHTML() string {
	return "<input id=\"" + b.id + "\" type=\"button\" value=\"" + b.text + "\"/> "
}

func (b *button) genJavaScript() string {
	return b.javascript
}

// 按钮被点击触发的事件
func (b *button) OnClick(f func()) {
	// 当收到b.id+"OnClick"事件时
	// 会执行函数f
	b.events[b.id+"OnClick"] = func(v string) {
		f()
	}
	b.javascript += `
(function() {
	var button = document.getElementById('` + b.id + `');
	button.onclick = function(){
		socket.send('{"type": "` + b.id + `OnClick","value": ""}')
	};
})();`
}

// 设置按钮文字
func (b *button) SetText(text string) {
	// b.send会去调用javascript里名为
	// b.id+"SetText"的函数
	// 并传入参数text
	b.send(b.id+"SetText", text)
	b.text = text
}

// 获取按钮文字
func (b *button) GetText() string {
	return b.text
}

type Theme struct {
	CSS        string
	JavaScript string
}

type Control interface {
	getEvents() map[string]func(v string)
	setSendFunc(func(f, v string))

	genHTML() string
	genJavaScript() string
}

type wsMsg struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
