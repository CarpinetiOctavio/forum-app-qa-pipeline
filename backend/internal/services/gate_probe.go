package services

// GateProbe exists only on a disposable branch: it is new code without a test,
// so a pull request into main shows whether the SonarCloud quality gate fails
// on uncovered new code. The branch is never merged.
func GateProbe(n int) string {
	if n < 0 {
		return "negative"
	}
	if n == 0 {
		return "zero"
	}
	if n%2 == 0 {
		return "even"
	}
	if n > 100 {
		return "large odd"
	}
	return "odd"
}
