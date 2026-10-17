package natsort

import (
	"sort"
	"strings"
	"testing"
)

func TestLess(t *testing.T) {
	in := []string{"10", "2", "a10", "A2", "1", "PG", "b", "VA 36", "VA 4", "07", "7", "Clé Gymnase", "clé 3"}
	sort.SliceStable(in, func(i, j int) bool { return Less(in[i], in[j]) })
	want := "1 2 7 07 10 A2 a10 b clé 3 Clé Gymnase PG VA 4 VA 36"
	if got := strings.Join(in, " "); got != want {
		t.Errorf("ordre naturel:\n got %s\nwant %s", got, want)
	}
}
