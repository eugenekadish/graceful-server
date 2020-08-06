package api

import (
	"testing"
)

func TestTruth(t *testing.T) {

	var proof = true
	if proof != true {
		t.Fatalf("expected proof: %v got %v", true, proof)
	}
}
