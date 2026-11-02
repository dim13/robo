package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestPlot(t *testing.T) {
	var buf bytes.Buffer
	p := plot(&buf, []byte("IN;SP1;PU0,0;PD400,0,400 400;PR-400,0;PU;PA0,0;XX1;"))
	got := strings.Split(strings.TrimSuffix(buf.String(), "\x03"), "\x03")
	want := []string{"M0,0", "D200,0", "D200,200", "D0,200", "M0,0", "M0,0"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := p.skipped["XX"], 1; got != want {
		t.Errorf("skipped XX: got %d, want %d", got, want)
	}
}

func TestCircle(t *testing.T) {
	var buf bytes.Buffer
	plot(&buf, []byte("PA400,400;CI40,90;"))
	got := strings.Split(strings.TrimSuffix(buf.String(), "\x03"), "\x03")
	want := []string{"M200,200", "M220,200", "D200,220", "D180,200", "D200,180", "D220,200", "M200,200", "M0,0"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLabel(t *testing.T) {
	var buf bytes.Buffer
	p := plot(&buf, []byte("SI1,1.4;LB-\x03;DI0,1;LB-中\x03"))
	got := strings.Split(strings.TrimSuffix(buf.String(), "\x03"), "\x03")
	want := []string{
		"M54.8,120", "D301.4,120", "M356.2,0", // along X
		"M236.2,54.8", "D236.2,301.4", "M356.2,356.2", // along Y
		"M0,0",
	}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := p.missing, 1; got != want {
		t.Errorf("missing: got %d, want %d", got, want)
	}
}
