package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const simple = "02-03-2013"

func TestFilter(t *testing.T) {
	tt := []struct {
		name string
		in   []*Listing
		out  []*Listing
		date time.Time
	}{
		{
			name: "Test filters out dates that arenot the current date",
			in: []*Listing{{
				Title: "one",
				Link:  "foobar",
				Date:  time.Date(2021, 02, 05, 0, 0, 0, 0, time.UTC),
			}, {
				Title: "two",
				Link:  "foobar2",
				Date:  time.Date(2021, 02, 04, 0, 0, 0, 0, time.UTC),
			}},
			out: []*Listing{{
				Title: "one",
				Link:  "foobar",
				Date:  time.Date(2021, 02, 05, 0, 0, 0, 0, time.UTC),
			}},
			date: time.Date(2021, 02, 05, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestClock(tc.date)
			out := filterListings(c, tc.in)
			assert.Equal(t, tc.out, out)
		})
	}
}

type testClock struct {
	now time.Time
}

func newTestClock(t time.Time) *testClock {
	return &testClock{now: t}
}

func (t *testClock) Now() time.Time {
	return t.now
}
