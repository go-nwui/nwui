# nwui

[![Gitter](https://img.shields.io/badge/GITTER-JOIN%20CHAT%20%E2%86%92-brightgreen.svg?style=flat)](https://gitter.im/go-nwui/nwui?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge) [![Go Walker](https://img.shields.io/badge/Go%20Walker-API%20Documentation-green.svg?style=flat)](https://gowalker.org/github.com/go-nwui/nwui) [![GoDoc](https://img.shields.io/badge/GoDoc-API%20Documentation-blue.svg?style=flat)](http://godoc.org/github.com/go-nwui/nwui)

node-webkit UI for Go

## Screenshot

![screenshot](screenshot.png)

## Example

创建一个包含按钮的窗口：

```go
&Window{
	Title:  "window",
	Width:  800,
	Height: 600,
	OnExit: func() {
		fmt.Println("exit")
	},
	Controls: []interface{}{
		&Button{
			ID:   "btn0",
			Text: "button",
			OnClick: func() {
				text := GetConByID("btn0").(*Button).Text
				fmt.Println(text, "clicked!")
			},
		},
	},
}
```

使用可以参见test文件

## 自带控件列表

* 普通的窗口控件
* 普通的按钮控件
* 还没移植完成的自定义窗口控件

## TODO

- [x] 基础框架
- [ ] 自动调用`nw.js`
- [ ] 更多控件
  - [x] 按钮 
  - [ ] 单选框
  - [ ] 多选框
  - [ ] 单行输入框
  - [ ] 多行输入框
  - [ ] 表格
  - [ ] 横向框架
  - [ ] 纵向框架
  - [ ] more
- [ ] 默认控件和框架分离
- [ ] 更多主题

## 控件编写指南

**本段是为想自己写nwui控件者编写的，如果只是普通的使用可以直接无视**

首先你要会一点js+css+html，以及强大的理解能力（没办法，作者写的太乱）

先上Button控件的例子：

```go
type Button struct {
	ID      string
	Text    string
	OnClick func()
	sender  chan EventMsg
}

func (b *Button) Init(sender chan EventMsg) Controls {
	static := ConStatic{
		Name: "Button",
		CSS: `
.button {
    border: 4px solid #304ffe;
    color: white;
    background: #304ffe;
    padding: 6px 12px;
}
.button:hover {
    background: white;
    color: #304ffe;
}
.button:active {
    color: #fff;
    background: #304ffe;
    box-shadow: 1px 2px 7px rgba(0, 0, 0, 0.3) inset;
}`,
		JavaScript: `
function ButtonSetText(id,text) {
	var button = document.getElementById(id);
	button.textContent = text;
}
(function() {
	var buttons = document.getElementsByClassName('button');
	for (var i = 0; i < buttons.length; i++) {
		var button = buttons[i]
		button.onclick = function(){
			send(button.id, "ButtonOnClick", "");
		};
	}
})();`,
	}

	if b.ID == "" {
		b.ID = NewControlID()
	}
	events := make(map[string]func(v string))
	b.sender = sender

	html := "<button id=\"" + b.ID + "\"class=\"button\">" + b.Text + "</button>"
	if b.OnClick != nil {
		// 如果用户使用了OnClick事件
		// 那么添加事件
		events["ButtonOnClick"] = func(v string) {
			b.OnClick()
		}
	}
	return Controls{
		b.ID: Control{
			V:      b,
			HTML:   html,
			Static: static,
			Events: events,
		},
	}
}

// 设置按钮文字
func (b *Button) SetText(text string) {
	// 这里的判断是防止控件还没有初始化
	// sender还未赋值用户就调用
	if b.sender != nil {
		// ButtonSetText 为需要调用的js函数
		b.sender <- EventMsg{b.ID, "ButtonSetText", text}
	}
	b.Text = text
}

```

以下为nwui中的结构体定义

```go
type Controls map[string]Control

type Control struct {
	V      interface{}               // 控件的结构体指针，用于初始化时的处理
	HTML   string                    // 控件的HTML代码
	Events map[string]func(v string) // 控件事件列表，key为事件名称
	Static ConStatic                 // 控件的静态数据，css和js等
}

type ConStatic struct {
	Name       string // 控件名称，不能和其他控件重复
	CSS        string // 控件的css
	JavaScript string // 控件的js
}

type EventMsg struct {
	ID    string `json:"id"`    // 控件的ID
	Event string `json:"event"` // 事件名称
	Value string `json:"value"` // 想发送信息的内容，复杂内容推荐用json编码
}

```

控件必须实现一个`Init(sender chan EventMsg) Controls`方法，nwui会用反射来调用并传入`sender`

`sender`是给前端发送消息的一个chan，具体使用可以参加上面源码

因为一个控件内可能会包含其他的控件，所以返回一个`Controls`而不是`Control`

当你编写的控件要包含其他控件时，比如写一个`Group`控件

可以获取子控件的`HTML`字段，然后按照需求添加到自己的`HTML`字段中（比如需要特殊排列啦什么的）。然后要记得清空子控件的`HTML`字段以防止重复添加（css和javascript之所以单独放在`ConStatic`是因为它们都是可重复使用的。`ConStatic`的`Name`字段也是用于区分控件类型的，所以`Name`不要随便填写）。return时记得在`Controls`里加上子控件（key为控件ID）

在写控件的js和css时要注意，必须写成对所有这种控件都可用的（增加代码复用率）。具体可以看上面`Button`控件的实现（根据class来选择控件啦什么的）

### 窗口级别的控件编写

窗口级别的控件Init方法比较特殊，比普通控件多两个返回值

```go
Init(sender chan EventMsg) (Controls, NWUI, chan bool)
```

`NWUI`是nwui窗口的配置，各个字段的意义可以直接看`nw.js`的`package.json`

当`chan bool`接收到消息时，会关闭窗口

### 看不懂？

作者的表达能力有限（或者说超烂），直接看源码吧……

## 参与开发

跪求帮忙开发，哪怕只增加注释或者增加TODO都可以

因为需要完成的控件太多所以作者一个人实在忙不过来