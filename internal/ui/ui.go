package ui

import (
	"image/color"
	"math"
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
	bars       []*canvas.LinearGradient
	stopAnim   chan bool
	currentLevel float64
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
	numBars := 32 // Increased for higher resolution
	bars := make([]*canvas.LinearGradient, numBars)
	barObjects := make([]fyne.CanvasObject, numBars)
	
	for i := 0; i < numBars; i++ {
		// Create a gradient for each bar
		grad := canvas.NewLinearGradient(
			color.RGBA{R: 0, G: 255, B: 255, A: 255},   // Cyan
			color.RGBA{R: 138, G: 43, B: 226, A: 255}, // BlueViolet
			0, // Vertical gradient
		)
		grad.SetMinSize(fyne.NewSize(6, 2)) // Thinner bars
		bars[i] = grad
		barObjects[i] = grad
	}

	// Use a Grid layout with small spacing
	visContainer := container.NewGridWithColumns(numBars, barObjects...)
	
	// Wrap visualizer in a container that gives it some height and centering
	visWrapper := container.NewPadded(visContainer)
	visWrapper.Resize(fyne.NewSize(240, 50))

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
		ticker := time.NewTicker(30 * time.Millisecond) // Faster update for smoothness
		defer ticker.Stop()
		
		t := 0.0
		for {
			select {
			case <-u.stopAnim:
				// Reset bars
				fyne.Do(func() {
					for _, b := range u.bars {
						b.SetMinSize(fyne.NewSize(6, 2))
						b.StartColor = color.RGBA{R: 0, G: 255, B: 255, A: 255}
						b.EndColor = color.RGBA{R: 138, G: 43, B: 226, A: 255}
						b.Refresh()
					}
				})
				return
			case <-ticker.C:
				t += 0.1
				fyne.Do(func() {
					numBars := len(u.bars)
					center := float64(numBars) / 2.0
					
					for i, b := range u.bars {
						// Symmetric wave simulation
						dist := math.Abs(float64(i) - center)
						
						// Combine sine waves for a "tech" feel
						// Base wave
						h1 := math.Sin(t + dist*0.5) 
						// Secondary faster wave
						h2 := math.Sin(t*2.0 - dist*0.8)
						
						// Normalize and scale
						val := (h1 + h2 + 2) / 4.0 // 0.0 to 1.0 approx
						
						// Add some randomness for "noise" but keep it smooth-ish
						noise := (rand.Float64() - 0.5) * 0.2
						val += noise
						
						// Scale by audio level
						// If silence (level < 0.01), keep it very low
						level := u.currentLevel
						if level < 0.05 {
							level = 0.05 // Minimum "alive" hum
						}
						val *= level

						if val < 0.05 { val = 0.05 }
						if val > 1.0 { val = 1.0 }
						
						height := float32(val * 45.0) // Max height 45
						
						b.SetMinSize(fyne.NewSize(6, height))
						
						// Dynamic color based on height
						// Higher bars get more purple/red, lower bars are cyan
						r := uint8(val * 255)
						b.EndColor = color.RGBA{R: r, G: 43, B: 226, A: 255}
						
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

func (u *UI) SetAudioLevel(level float64) {
	u.currentLevel = level
}

func (u *UI) ShowRecording() {
	fyne.Do(func() {
		u.status.SetText("") // Clear text for cleaner UI
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
