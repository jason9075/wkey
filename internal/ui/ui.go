package ui

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"time"

	"wkey/internal/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type UI struct {
	app          fyne.App
	window       fyne.Window
	status       *widget.Label
	indicator    *canvas.Circle
	visualizer   *fyne.Container
	bars         []*canvas.LinearGradient
	stopAnim     chan bool
	currentLevel float64
	config       *config.Config
}

func New(cfg *config.Config) *UI {
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
	numBars := cfg.Visual.BarCount
	bars := make([]*canvas.LinearGradient, numBars)

	// Manual layout parameters
	maxH := float32(50)
	barW := float32(200 / float32(numBars)) // Scale bar width based on count
	if barW < 2 {
		barW = 2
	}
	gap := float32(1)

	// Use WithoutLayout for manual positioning
	visContainer := container.NewWithoutLayout()
	visContainer.Resize(fyne.NewSize(float32(numBars)*(barW+gap), maxH))

	startColor := parseHexColor(cfg.Visual.BarColorStart)
	endColor := parseHexColor(cfg.Visual.BarColorEnd)

	for i := 0; i < numBars; i++ {
		// Create a gradient for each bar
		grad := canvas.NewLinearGradient(
			startColor,
			endColor,
			0, // Vertical gradient
		)

		// Initial position at bottom, minimal height
		x := float32(i) * (barW + gap)
		grad.Resize(fyne.NewSize(barW, 2))
		grad.Move(fyne.NewPos(x, maxH-2))

		bars[i] = grad
		visContainer.Add(grad)
	}

	// Wrap visualizer in a container that gives it some height and centering
	visWrapper := container.NewPadded(visContainer)
	visWrapper.Resize(fyne.NewSize(240, 60)) // Adjusted height for padding

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
		config:     cfg,
	}
}

func (u *UI) startVisualizer() {
	// Stop existing if any (though we usually stop before start)
	go func() {
		ticker := time.NewTicker(30 * time.Millisecond) // Faster update for smoothness
		defer ticker.Stop()

		t := 0.0

		// Initialize random offsets for each bar
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		numBars := len(u.bars)
		offsets := make([]float64, numBars)
		for i := 0; i < numBars; i++ {
			offsets[i] = 0.1 + rng.Float64()*1.3 // Random between 0.1 and 1.4
		}

		// Layout constants
		maxH := float32(50)
		barW := float32(200 / float32(numBars))
		if barW < 2 {
			barW = 2
		}
		gap := float32(1)

		startColor := parseHexColor(u.config.Visual.BarColorStart)
		endColor := parseHexColor(u.config.Visual.BarColorEnd)

		for {
			select {
			case <-u.stopAnim:
				// Reset bars
				fyne.Do(func() {
					for i, b := range u.bars {
						x := float32(i) * (barW + gap)
						b.Resize(fyne.NewSize(barW, 2))
						b.Move(fyne.NewPos(x, maxH-2))
						b.StartColor = startColor
						b.EndColor = endColor
						b.Refresh()
					}
				})
				return
			case <-ticker.C:
				t += 0.1 * u.config.Visual.AnimationSpeed
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
						// Third wave for more variation
						h3 := math.Cos(t*1.5 + dist*0.3)

						// Normalize and scale
						val := (h1 + h2 + h3 + 3) / 6.0 // 0.0 to 1.0 approx

						// Add some randomness for "noise" but keep it smooth-ish
						noise := (rng.Float64() - 0.5) * 0.2 // Reduced noise
						val += noise

						// Scale by audio level
						// If silence (level < 0.01), keep it very low
						level := u.currentLevel * 3.0 // Boost sensitivity
						if level > 1.0 {
							level = 1.0
						}
						if level < 0.05 {
							level = 0.05 // Minimum "alive" hum
						}
						val *= level

						// Add a center bias so middle bars are generally higher
						centerBias := 1.0 - (dist / center * 0.5)
						val *= centerBias

						// Apply persistent random offset
						val *= offsets[i]

						// Non-linear scaling to accentuate peaks and suppress lows
						val = val * val

						if val < 0.05 {
							val = 0.05
						}
						if val > 1.0 {
							val = 1.0
						}

						height := float32(val * 45.0) // Max height 45
						if height < 2 {
							height = 2
						}
						if height > maxH {
							height = maxH
						}

						// Manual layout update
						x := float32(i) * (barW + gap)
						b.Resize(fyne.NewSize(barW, height))
						b.Move(fyne.NewPos(x, maxH-height))

						// Apply colors from config
						b.StartColor = startColor
						b.EndColor = endColor

						b.Refresh()
					}
					u.visualizer.Refresh()
				})
			}
		}
	}()
}

func parseHexColor(s string) color.Color {
	c := color.RGBA{A: 255}
	if len(s) == 7 && s[0] == '#' {
		fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	}
	return c
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
		u.status.SetText("Execute again to stop") // Clear text for cleaner UI
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

func (u *UI) ShowDone(text string) {
	u.stopVisualizer()
	fyne.Do(func() {
		u.status.SetText(text)
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

func (u *UI) Hide() {
	fyne.Do(func() {
		u.window.Hide()
	})
}

func (u *UI) Show() {
	fyne.Do(func() {
		u.window.Show()
	})
}

func (u *UI) Run() {
	u.app.Run()
}

func (u *UI) Quit() {
	u.stopVisualizer()
	fyne.Do(func() {
		u.app.Quit()
	})
}
