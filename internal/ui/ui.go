package ui

import (
	"image/color"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type UI struct {
	app        fyne.App
	window     fyne.Window
	status     *widget.Label
	indicator  *canvas.Circle
	visualizer *fyne.Container
	bars       []*canvas.Rectangle
	stopAnim   chan bool
}

func New() *UI {
	a := app.New()
	w := a.NewWindow("Voice Input")
	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(250, 120)) // Slightly larger for visualizer
	// w.SetDecorated(false) // Overlay style

	status := widget.NewLabel("Initializing...")
	status.Alignment = fyne.TextAlignCenter

	indicator := canvas.NewCircle(color.RGBA{R: 200, G: 200, B: 200, A: 255})
	indicator.Resize(fyne.NewSize(10, 10))

	// Visualizer Bars
	numBars := 10
	bars := make([]*canvas.Rectangle, numBars)
	barObjects := make([]fyne.CanvasObject, numBars)
	for i := 0; i < numBars; i++ {
		rect := canvas.NewRectangle(color.RGBA{R: 100, G: 100, B: 255, A: 255})
		rect.SetMinSize(fyne.NewSize(15, 5)) // Min height
		bars[i] = rect
		barObjects[i] = rect
	}
	// Use a Grid layout for bars, but we want them to align bottom. 
	// Actually HBox is better, but we need to control height.
	// Let's use a container with a custom layout or just HBox and update MinSize?
	// HBox tries to fill height.
	// Let's use a Grid with 1 row, numBars columns.
	visContainer := container.NewGridWithColumns(numBars, barObjects...)
	
	// Wrap visualizer in a container that gives it some height
	visWrapper := container.NewPadded(visContainer)
	visWrapper.Resize(fyne.NewSize(200, 40))

	content := container.New(layout.NewVBoxLayout(),
		container.NewCenter(indicator),
		status,
		visWrapper,
	)
	w.SetContent(content)
	w.CenterOnScreen()

	return &UI{
		app:        a,
		window:     w,
		status:     status,
		indicator:  indicator,
		visualizer: visContainer,
		bars:       bars,
		stopAnim:   make(chan bool),
	}
}

func (u *UI) startVisualizer() {
	// Stop existing if any (though we usually stop before start)
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		
		for {
			select {
			case <-u.stopAnim:
				// Reset bars
				fyne.Do(func() {
					for _, b := range u.bars {
						b.SetMinSize(fyne.NewSize(15, 5))
						b.FillColor = color.RGBA{R: 100, G: 100, B: 255, A: 255}
						b.Refresh()
					}
				})
				return
			case <-ticker.C:
				fyne.Do(func() {
					for _, b := range u.bars {
						// Random height between 5 and 40
						h := float32(5 + rand.Intn(35))
						b.SetMinSize(fyne.NewSize(15, h))
						// Randomize color slightly for effect
						b.FillColor = color.RGBA{
							R: 100, 
							G: uint8(100 + rand.Intn(155)), 
							B: 255, 
							A: 255,
						}
						b.Refresh()
					}
					u.visualizer.Refresh()
				})
			}
		}
	}()
}

func (u *UI) stopVisualizer() {
	// Non-blocking send
	select {
	case u.stopAnim <- true:
	default:
	}
}

func (u *UI) ShowRecording() {
	fyne.Do(func() {
		u.status.SetText("Recording...")
		u.indicator.FillColor = color.RGBA{R: 255, G: 0, B: 0, A: 255} // Red
		u.indicator.Refresh()
		u.window.Show()
	})
	u.startVisualizer()
}

func (u *UI) ShowTranscribing() {
	u.stopVisualizer()
	fyne.Do(func() {
		u.status.SetText("Transcribing...")
		u.indicator.FillColor = color.RGBA{R: 255, G: 165, B: 0, A: 255} // Orange
		u.indicator.Refresh()
	})
}

func (u *UI) ShowDone() {
	u.stopVisualizer()
	fyne.Do(func() {
		u.status.SetText("Done!")
		u.indicator.FillColor = color.RGBA{R: 0, G: 255, B: 0, A: 255} // Green
		u.indicator.Refresh()
	})
}

func (u *UI) ShowError(err string) {
	u.stopVisualizer()
	fyne.Do(func() {
		u.status.SetText("Error: " + err)
		u.indicator.FillColor = color.RGBA{R: 255, G: 0, B: 0, A: 255} // Red
		u.indicator.Refresh()
	})
}

func (u *UI) Run() {
	u.app.Run()
}

func (u *UI) Quit() {
	u.stopVisualizer()
	u.app.Quit()
}
