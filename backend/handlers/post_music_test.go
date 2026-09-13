package handlers

import (
	"math"
	"testing"
)

func TestValidatePostMusicWindowAcceptsDefaultWindow(t *testing.T) {
	if err := validatePostMusicWindow(postMusicInput{StartSec: 0, EndSec: 0, Volume: 1}, 0); err != nil {
		t.Fatalf("default music window rejected: %v", err)
	}
}

func TestValidatePostMusicWindowRejectsNegativeRange(t *testing.T) {
	cases := []postMusicInput{
		{StartSec: -1, EndSec: 0, Volume: 1},
		{StartSec: 0, EndSec: -1, Volume: 1},
	}
	for _, input := range cases {
		if err := validatePostMusicWindow(input, 120); err == nil {
			t.Fatalf("expected negative music window to be rejected: %+v", input)
		}
	}
}

func TestValidatePostMusicWindowRejectsInvalidOrder(t *testing.T) {
	for _, input := range []postMusicInput{
		{StartSec: 10, EndSec: 10, Volume: 1},
		{StartSec: 20, EndSec: 10, Volume: 1},
	} {
		if err := validatePostMusicWindow(input, 120); err == nil {
			t.Fatalf("expected invalid music order to be rejected: %+v", input)
		}
	}
}

func TestValidatePostMusicWindowRejectsOverlongWindow(t *testing.T) {
	input := postMusicInput{StartSec: 0, EndSec: maxPostMusicDuration + 0.01, Volume: 1}
	if err := validatePostMusicWindow(input, 0); err == nil {
		t.Fatal("expected music window above 10 minutes to be rejected")
	}
}

func TestValidatePostMusicWindowRejectsTrackOverflow(t *testing.T) {
	cases := []postMusicInput{
		{StartSec: 120, EndSec: 0, Volume: 1},
		{StartSec: 90, EndSec: 121, Volume: 1},
	}
	for _, input := range cases {
		if err := validatePostMusicWindow(input, 120); err == nil {
			t.Fatalf("expected track-duration overflow to be rejected: %+v", input)
		}
	}
}

func TestValidatePostMusicWindowRejectsInvalidVolume(t *testing.T) {
	for _, volume := range []float64{-0.01, 1.01} {
		if err := validatePostMusicWindow(postMusicInput{StartSec: 0, EndSec: 0, Volume: volume}, 120); err == nil {
			t.Fatalf("expected invalid music volume %v to be rejected", volume)
		}
	}
}

func TestValidatePostMusicWindowRejectsNonFiniteValues(t *testing.T) {
	cases := []postMusicInput{
		{StartSec: math.NaN(), Volume: 1},
		{EndSec: math.Inf(1), Volume: 1},
		{Volume: math.Inf(-1)},
	}
	for _, input := range cases {
		if err := validatePostMusicWindow(input, 120); err == nil {
			t.Fatalf("expected non-finite music value to be rejected: %+v", input)
		}
	}

	if err := validatePostMusicWindow(postMusicInput{Volume: 1}, math.NaN()); err == nil {
		t.Fatal("expected non-finite duration to be rejected")
	}
	if err := validatePostMusicWindow(postMusicInput{Volume: 1}, math.Inf(1)); err == nil {
		t.Fatal("expected infinite duration to be rejected")
	}
}

func TestValidatePostMusicWindowAllowsTrackBoundary(t *testing.T) {
	input := postMusicInput{StartSec: 60, EndSec: 120, Volume: 0.75}
	if err := validatePostMusicWindow(input, 120); err != nil {
		t.Fatalf("expected exact track boundary to be accepted: %v", err)
	}
}
