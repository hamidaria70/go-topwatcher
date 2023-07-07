package main

import (
	"testing"
	"topwatcher/cmd"
)

func TestIsException(t *testing.T) {
	got := cmd.IsException("deployment1", "pod1", []string{"deployment1", "deployment2", "deployment3"})
	want := false

	if got != want {
		t.Errorf("got %v, wanted %v", got, want)
	}
}
