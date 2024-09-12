package tracker_test

import (
	"axiomiety/go-bt/data"
	"axiomiety/go-bt/tracker"
	"testing"
)

func TestQueryString(t *testing.T) {
	q := data.TrackerQuery{
		InfoHash: "deadbeef",
		PeerId:   "foo",
		Left:     3,
	}

	expected := "info_hash=deadbeef&peer_id=foo&uploaded=0&downloaded=0&left=3&event="
	qstring := tracker.ToQueryString(&q)
	if qstring != expected {
		t.Errorf("expected %s but got %s", expected, qstring)
	}
}
