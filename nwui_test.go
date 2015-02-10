package nwui

import (
	"fmt"
	"testing"
	"time"
)

func Test_Window(t *testing.T) {
	w := &Window{
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

	go func() {
		time.Sleep(time.Second * 5)
		btn := GetConByID("btn0").(*Button)
		btn.SetText("text!")
		fmt.Println(btn.Text)
	}()

	Show(w)
}
