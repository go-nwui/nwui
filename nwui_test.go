package nwui

import (
	"fmt"
	"testing"
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
	w.Show()
}
