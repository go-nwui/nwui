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
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var (
	// 用于生成随机数
	r    = rand.New(rand.NewSource(time.Now().UnixNano()))
	cons = make(map[string]interface{})
	temp = `<!DOCTYPE html>
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
	eval(data["event"]+"('"+data["value"]+"')")
}
window.onunload = function() {
	socket.send(JSON.stringify({"event": "exit","value": ""}));
}
%v
</script>
</body>
</html>`
	defaultTheme = Theme{
		CSS: `button {
    border: 4px solid #304ffe;
    color: white;
    background: #304ffe;
    padding: 6px 12px;
}
button:hover {
    background: white;
    color: #304ffe;
}
button:active {
    color: #fff;
    background: #304ffe;
    box-shadow: 1px 2px 7px rgba(0, 0, 0, 0.3) inset;
}
.frame {
    position: absolute;
    left: 0px;
    top: 0px;
    width: 100%;
    height: 32px;
    background-color: #424242;
    -webkit-app-region: drag;
}
.frame .title {
    color: white;
    position: absolute;
    left: 12px;
    width: 80%;
    margin-top: 6px;
    margin-bottom: 6px;
    font-size: 11pt;
}
.frame button#close {
    position: absolute;
    left: auto;
    right: 12px;
    width: auto;
    font-size: 11pt;
    -webkit-app-region: no-drag;
}
.main {
    margin-top: 40px;
}`,
	}
)

func printInfo(v ...interface{}) {
	log.Println(append([]interface{}{"[nwui][Info]"}, v...)...)
}

func printError(v ...interface{}) {
	log.Println(append([]interface{}{"[nwui][Error]"}, v...)...)
}

// 生成控件ID
func NewControlID() string { return "_" + strconv.FormatInt(r.Int63(), 36) }

func GetConByID(id string) interface{} {
	return nil
}

// nwui窗口
type Window struct {
	Title     string
	Width     int
	Height    int
	MaxWidth  int
	MaxHeight int
	MinWidth  int
	MinHeight int
	Controls  []interface{}
	OnExit    func()
	theme     Theme
	exit      chan bool
}

// 使用主题（CSS+JavaScript）
func (w *Window) UseTheme(t Theme) {
	w.theme = t
}

// 显示窗口
// 必须在全部控件设置完毕后才能调用
func (w *Window) Show() {
	w.exit = make(chan bool)

	// 查找可用端口
	var port string
	for i := 7072; i <= 65536; i++ {
		p := strconv.Itoa(i)
		ln, err := net.Listen("tcp", "localhost:"+p)
		if err != nil {
			continue
		} else {
			port = p
			ln.Close()
			break
		}
	}
	if port == "" {
		printError("no port can use")
		os.Exit(1)
	}

	var (
		html      string
		js        string = w.theme.JavaScript
		allEvents        = make(map[string]func(v string))
		upgrader         = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
		sender = make(chan EventMsg)
	)
	// 初始化控件
	for _, v := range w.Controls {
		// 返回html string, javascript string, events map[string]func(v string)
		v := reflect.ValueOf(v).MethodByName("Init").Call([]reflect.Value{reflect.ValueOf(sender)})

		html += v[0].String()
		js += v[1].String()
		for _, vv := range v[2].MapKeys() {
			allEvents[vv.String()] = v[2].MapIndex(vv).Interface().(func(v string))
		}
	}
	go func() {
		http.HandleFunc("/ws", func(rw http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(rw, r, nil)
			if err != nil && err != io.EOF {
				printError(err)
				return
			}

			// 用于接收事件
			go func() {
				for {
					messageType, p, err := conn.ReadMessage()
					if err != nil {
						printError(err)
						return
					}

					if messageType == websocket.TextMessage {
						var msg EventMsg
						err = json.Unmarshal(p, &msg)
						if err != nil {
							printError(err)
							return
						}
						// 判断事件是否为内置事件
						// 进行相应处理
						switch msg.Event {
						case "exit":
							if w.OnExit != nil {
								w.OnExit()
							}
							w.exit <- true
							return
						}
						// 执行事件所绑定的函数
						f, ok := allEvents[msg.Event]
						if ok {
							f(msg.Value)
						} else {
							printError("unfind event:", msg.Event)
						}
					}
				}
			}()

			for {
				// 发送消息给前端
				m := <-sender
				msg, err := json.Marshal(&m)
				if err != nil {
					printError(err)
					return
				}
				err = conn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					printError(err)
				}
			}
		})
		http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(rw, temp, w.Title, w.theme.CSS, html, "localhost:"+port, js)
			r.Body.Close()
		})
		printInfo("running on localhost:" + port)
		err := http.ListenAndServe("localhost:"+port, nil)
		if err != nil {
			panic(err)
		}
	}()
	<-w.exit
}

/*
// 创建一个新的自定义窗口边框
// 一个 Window 中只能有一个 Frame
func NewFrame(title string, con ...Control) Frame {
	return &frame{
		title:    title,
		controls: con,
		events:   make(map[string]func(v string)),
	}
}

// 窗口边框
type Frame interface {
	Control
}

type frame struct {
	title    string
	js       string
	controls []Control
	events   map[string]func(v string)
	send     func(f, v string)
}

func (f *frame) getEvents() map[string]func(v string) {
	return f.events
}

func (f *frame) setSendFunc(fc func(f, v string)) {
	f.send = fc
}

func (f *frame) genHTML() string {
	// 转换内部控件
	var html string
	for _, v := range f.controls {
		v.setSendFunc(f.send)
		html += v.genHTML()
		f.js += v.genJavaScript()
		for e, fc := range v.getEvents() {
			f.events[e] = fc
		}
	}
	return `<div class="frame">
<a class="title">` + f.title + `</a>
<button id="close">x</button>
</div>
<div class="main">
` + html + `
</div>`
}

func (f *frame) genJavaScript() string {
	return f.js + `
(function() {
	document.getElementById("close").onclick = function(){
		open(' ', '_self').close();
	}
})();`
}
*/

type Button struct {
	Control
	ID      string
	Text    string
	OnClick func()
	sender  chan EventMsg
}

func (b *Button) Init(sender chan EventMsg) (string, string, map[string]func(v string)) {
	if b.ID == "" {
		b.ID = NewControlID()
	}
	events := make(map[string]func(v string))
	b.sender = sender
	js := `
function ` + b.ID + `SetText(text) {
	var button = document.getElementById('` + b.ID + `');
	button.textContent = text;
}`
	html := "<button id=\"" + b.ID + "\"/>" + b.Text + "</button>"
	if b.OnClick != nil {
		// 如果用户使用了OnClick事件
		// 那么添加事件和js
		events[b.ID+"OnClick"] = func(v string) {
			b.OnClick()
		}
		js += `
(function() {
	var button = document.getElementById('` + b.ID + `');
	button.onclick = function(){
		socket.send(JSON.stringify({"event": "` + b.ID + `OnClick","value": ""}));
	};
})();`
	}
	return html, js, events
}

// 设置按钮文字
func (b *Button) SetText(text string) {
	// b.send会去调用javascript里名为
	// b.id+"SetText"的函数
	// 并传入参数text
	if b.sender != nil {
		b.sender <- EventMsg{b.ID + "SetText", text}
	}
	b.Text = text
}

// nwui主题
type Theme struct {
	CSS        string
	JavaScript string
}

// nwui控件
type Control struct {
	getEvents     func() map[string]func(v string)
	setSendFunc   func(func(f, v string))
	genHTML       func() string
	genJavaScript func() string
}

type EventMsg struct {
	Event string `json:"event"`
	Value string `json:"value"`
}
