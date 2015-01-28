package nwui

import (
	"fmt"
	"testing"
)

func Test_Window(t *testing.T) {
	w := NewWindow("test window", 800, 600)
	button := NewButton("button")
	button.OnClick(func() {
		fmt.Println("button clicked!")
	})
	w.Show(button)
}
