package robo

import "testing"

func TestUnitString(t *testing.T) {
	testCases := []struct {
		u    Unit
		want string
	}{
		{u: 0, want: "0"},
		{u: 100, want: "100"},
		{u: -100, want: "-100"},
		{u: 1143.797, want: "1143.8"},
		{u: 10.25, want: "10.25"},
		{u: 10.256, want: "10.26"},
		{u: -1.5, want: "-1.5"},
		{u: -0.4, want: "0"},
		{u: -0.5, want: "-1"},
		{u: -0.004, want: "0"},
	}
	for _, tc := range testCases {
		t.Run(tc.want, func(t *testing.T) {
			if got, want := tc.u.String(), tc.want; got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}
