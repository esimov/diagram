package gui

import (
	"math"
	"time"

	"gioui.org/layout"
)

type EasingFormula int

const (
	EaseInOut EasingFormula = iota
	EaseInSine
	EaseOutSine
	EaseInOutSine
	EaseInOutBack
	EaseInOutQuad
	EaseInElastic
	EaseOutElastic
	EaseInOutElastic
	EaseBezier
)

type Easing interface {
	Update(gtx layout.Context) float64
	Animate(formula EasingFormula, t float64) float64
}

var _ Easing = (*Animation)(nil)

type Animation struct {
	StartTime time.Time
	Delta     time.Duration
	Duration  time.Duration
}

// Update updates the time passed from the initial invocation
// of time.Now until the time set as duration is not reached.
func (a *Animation) Update(gtx layout.Context) float64 {
	delta := gtx.Now.Sub(a.StartTime)

	if delta < a.Duration {
		a.Delta += delta
		if a.Delta > a.Duration {
			a.Delta = a.Duration
		}
	}

	// Calculate the time passed from the first invocation of the time.Now.
	return float64(a.Delta) / float64(a.Duration)
}

func (a *Animation) IsScrolling() bool {
	return a.Delta != a.Duration
}

// Animate applies the In-Out-Back easing formula to a specific event.
func (a *Animation) Animate(formula EasingFormula, t float64) float64 {
	switch formula {
	case EaseInOut:
		if t < 0.5 {
			return 2 * t * t
		} else {
			return 1 - math.Pow(-2*t+2, 2)/2
		}
	case EaseInSine:
		return 1 - math.Cos((t*math.Phi)/2)
	case EaseOutSine:
		return math.Sin((t * math.Phi) / 2)
	case EaseInOutSine:
		return -(math.Cos((t*math.Phi)-1) / 4)
	case EaseInOutBack:
		s := 1.70158
		t *= 2
		if t < 1 {
			s *= 1.525
			return 0.5 * (t * t * ((s+1)*t - s))
		} else {
			t -= 2
			s *= 1.525
			return 0.5 * (t*t*((s+1)*t+s) + 2)
		}
	case EaseInOutQuad:
		if t <= 0.5 {
			return 2.0 * t * t
		} else {
			t -= 0.5
			return 2.0*t*(1.0-t) + 0.5
		}
	case EaseInElastic:
		if t <= 0 {
			return 0
		}
		if t >= 1 {
			return 1
		}
		const c4 = (2 * math.Pi) / 3
		return -math.Pow(2, 10*(t-1)) * math.Sin((t-1.075)*c4)
	case EaseOutElastic:
		if t <= 0 {
			return 0
		}
		if t >= 1 {
			return 1
		}

		c4 := (2 * math.Pi) / 3
		return -math.Pow(2, 10*(t-1)) * math.Sin((t-1.075)*c4)
	case EaseInOutElastic:
		if t <= 0 {
			return 0
		}
		if t >= 1 {
			return 1
		}

		const c5 = (2 * math.Pi) / 4.5

		t *= 2
		if t < 1 {
			return -0.5 * math.Pow(2, 10*(t-1)) * math.Sin((t-1.1125)*c5)
		}
		return 0.5*math.Pow(2, -10*(t-1))*math.Sin((t-1.1125)*c5) + 1
	case EaseBezier:
		return t * t * (3.0 - 2.0*t)
	default:
		return 1
	}
}
