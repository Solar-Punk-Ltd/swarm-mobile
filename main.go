package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/Solar-Punk-Ltd/swarm-mobile/internal/screens"
)

func main() {
	a := app.NewWithID("com.solarpunk.swarmmobile")

	w := a.NewWindow("Swarm Mobile")
	w.SetMaster()

	w.Resize(fyne.NewSize(390, 422))
	w.SetFixedSize(true)
	w.SetContent(screens.Make(a, w))
	w.ShowAndRun()
	tidyUp("Window Closed")
}

func tidyUp(msg string) {
	fmt.Printf("Swarm Mobile Exited: %s\n", msg)
}
