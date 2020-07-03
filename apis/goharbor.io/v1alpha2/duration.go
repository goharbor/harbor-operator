package v1alpha2

import "time"

// +kubebuilder:validation:Pattern="([0-9]+h)?([0-9]+m)?([0-9]+s)?([0-9]+ms)?([0-9]+us)?([0-9]+µs)?([0-9]+ns)?"
// PositiveDuration represents a positive duration of time.
type PositiveDuration string

func (d PositiveDuration) Duration() (time.Duration, error) {
	return time.ParseDuration(string(d))
}

func NewPositiveDuration(duration time.Duration) PositiveDuration {
	return PositiveDuration(duration.String())
}
