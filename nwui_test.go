package nwui

import (
	"testing"
)

func Test_Window(t *testing.T) {
	var w *Window
	w = &Window{
		Title:  "window",
		Width:  800,
		Height: 600,
		Controls: []interface{}{
			&Button{
				ID:   "btn0",
				Text: "button",
				OnClick: func() {
					println(GetConByID("btn0").(*Button).Text)
				},
			},
		},
	}
	w.Show()
}
