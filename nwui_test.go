package nwui

import (
	"fmt"
	"testing"
	"time"
)

func Test_Window(t *testing.T) {
	w := NewWindow("test window", 800, 600)
	button := NewButton("button")
	button.OnClick(func() {
		fmt.Println("button clicked!")
	})
	go func() {
		time.Sleep(time.Second * 10)
		button.SetText("hello!!")
	}()
	w.OnExit(func() {
		fmt.Println("EXIT!")
	})
	w.Show(button)
}
