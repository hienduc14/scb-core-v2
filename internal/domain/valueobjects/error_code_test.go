package valueobjects

import "testing"

func TestErrorCodeTerminalBehavior(t *testing.T) {
	if ErrStatusPublish.IsTerminal() {
		t.Fatal("status publish errors should not be terminal")
	}
	if !ErrScanExecution.IsTerminal() {
		t.Fatal("scan execution errors should be terminal")
	}
}
