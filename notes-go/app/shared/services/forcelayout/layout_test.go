package forcelayout

import (
	"fmt"
	"maps"
	"math"
	"slices"
	"testing"
)

func note(id string, tags ...string) Note { return Note{ID: id, Tags: tags} }

func layout(t *testing.T, notes []Note, links []Link) map[string]Point {
	t.Helper()
	positions, err := Simulate(notes, links, nil)
	if err != nil {
		t.Fatal(err)
	}
	return Fit(positions)
}

func distance(positions map[string]Point, a, b string) float64 {
	return math.Hypot(positions[a].X-positions[b].X, positions[a].Y-positions[b].Y)
}

func pull(edges []Edge, source, target string) Edge {
	for _, edge := range edges {
		if edge.Source == source && edge.Target == target {
			return edge
		}
	}
	return Edge{}
}

func TestEmptyAndSingle(t *testing.T) {
	if got := layout(t, nil, nil); len(got) != 0 {
		t.Errorf("no note: %v", got)
	}
	if got := layout(t, []Note{note("a")}, nil)["a"]; got != (Point{50, 50}) {
		t.Errorf("one note at %v, want the center", got)
	}
}

func TestTagsAreNodes(t *testing.T) {
	got := slices.Sorted(maps.Keys(layout(t, []Note{note("a", "go"), note("b", "go")}, nil)))
	if want := []string{"#go", "a", "b"}; !slices.Equal(got, want) {
		t.Errorf("nodes %v, want %v", got, want)
	}
}

func TestRareTagPullsHarder(t *testing.T) {
	notes := []Note{note("a", "rare", "common"), note("b", "rare", "common")}
	for _, id := range "cdefgh" {
		notes = append(notes, note(string(id), "common"))
	}
	edges := Edges(notes, nil)
	if pull(edges, "a", "#rare").Strength <= pull(edges, "a", "#common").Strength {
		t.Error("the rare tag does not pull harder")
	}
	// c has the common tag alone: the whole of its weight, where a splits it in two.
	if c, a := pull(edges, "c", "#common").Strength, pull(edges, "a", "#common").Strength; math.Abs(c-2*a) > 1e-12 {
		t.Errorf("c pulled %v, a %v", c, a)
	}
	if Edges([]Note{note("a", "all"), note("b", "all")}, nil)[0].Strength <= 0 {
		t.Error("a tag on every note no longer holds them")
	}
	if got := Edges([]Note{note("a", "go")}, nil)[0].Strength; got != 1 {
		t.Errorf("a lone note pulled %v, want 1", got)
	}
}

func TestBigTagKeepsNotesFurther(t *testing.T) {
	var notes []Note
	for i := range 30 {
		tags := []string{"big"}
		if i < 2 {
			tags = append(tags, "small")
		}
		notes = append(notes, note(fmt.Sprintf("n%d", i), tags...))
	}
	edges := Edges(notes, nil)
	if pull(edges, "n0", "#big").Distance <= pull(edges, "n0", "#small").Distance {
		t.Error("the big tag is not further away")
	}
}

func TestSmallTagAndLinkGatherCloser(t *testing.T) {
	var notes []Note
	for i := range 40 {
		notes = append(notes, note(fmt.Sprintf("n%d", i), "big"))
	}
	notes[0].Tags = append(notes[0].Tags, "small")
	notes[1].Tags = append(notes[1].Tags, "small")
	positions := layout(t, notes, []Link{{"n2", "n3"}})
	spread := distance(positions, "n10", "n20")
	if got := distance(positions, "n0", "n1"); got >= spread {
		t.Errorf("small tag %v, big %v", got, spread)
	}
	if got := distance(positions, "n2", "n3"); got >= spread {
		t.Errorf("linked %v, big %v", got, spread)
	}
}

func TestNotesGatherAroundTheirTag(t *testing.T) {
	groups := map[string][]string{"go": {"a", "b", "c"}, "cuisine": {"d", "e", "f"}}
	var notes []Note
	for _, tag := range []string{"go", "cuisine"} {
		for _, id := range groups[tag] {
			notes = append(notes, note(id, tag))
		}
	}
	positions := layout(t, notes, nil)
	for tag, ids := range groups {
		other := map[string]string{"go": "cuisine", "cuisine": "go"}[tag]
		for _, id := range ids {
			if distance(positions, id, TagNode(tag)) >= distance(positions, id, TagNode(other)) {
				t.Errorf("%s nearer %s", id, other)
			}
		}
	}
}

func TestLinkAlonePullsTogether(t *testing.T) {
	positions := layout(t, []Note{note("a"), note("b"), note("c")}, []Link{{"a", "b"}})
	for id, at := range positions {
		if math.IsNaN(at.X) || math.IsNaN(at.Y) || math.IsInf(at.X, 0) || math.IsInf(at.Y, 0) {
			t.Errorf("%s at %v", id, at)
		}
	}
	if distance(positions, "a", "b") >= distance(positions, "a", "c") {
		t.Error("the link does not pull a and b together")
	}
}

func TestEveryNoteInTheView(t *testing.T) {
	var notes []Note
	for i := range 80 {
		notes = append(notes, note(fmt.Sprintf("n%d", i), "go"))
	}
	positions := layout(t, notes, []Link{{"n0", "n1"}})
	if len(positions) != 81 {
		t.Fatalf("%d positions, want 81", len(positions))
	}
	for id, at := range positions {
		if at.X < margin || at.X > 100-margin || at.Y < margin || at.Y > 100-margin {
			t.Errorf("%s at %v, out of the view", id, at)
		}
	}
}

// Past collideLimit the collisions stop: the graph must still settle and stay in the view.
func TestBigGraph(t *testing.T) {
	var notes []Note
	for i := range 2000 {
		notes = append(notes, note(fmt.Sprintf("n%d", i), fmt.Sprintf("t%d", i%7), fmt.Sprintf("u%d", i%31)))
	}
	positions := layout(t, notes, []Link{{"n0", "n1"}, {"n1", "n2"}})
	if len(positions) != 2000+7+31 {
		t.Fatalf("%d positions", len(positions))
	}
	for id, at := range positions {
		if !(at.X >= margin && at.X <= 100-margin && at.Y >= margin && at.Y <= 100-margin) {
			t.Fatalf("%s at %v", id, at)
		}
	}
}

func TestSameNotesSameMap(t *testing.T) {
	notes, links := []Note{note("a", "go"), note("b", "go"), note("c")}, []Link{{"a", "c"}}
	if a, b := layout(t, notes, links), layout(t, notes, links); !maps.Equal(a, b) {
		t.Errorf("two maps: %v, %v", a, b)
	}
}

func TestSignature(t *testing.T) {
	notes := []Note{note("a", "go"), note("b")}
	if Signature(notes, nil) != Signature([]Note{note("a", "go"), note("b")}, nil) {
		t.Error("the same graph gives two signatures")
	}
	if Signature(notes, nil) == Signature([]Note{note("a", "web"), note("b")}, nil) {
		t.Error("a changed tag keeps the signature")
	}
	if Signature(notes, nil) == Signature(notes, []Link{{"a", "b"}}) {
		t.Error("a new link keeps the signature")
	}
}

func TestWarmStart(t *testing.T) {
	var notes []Note
	for _, id := range "abc" {
		notes = append(notes, note(string(id), "go"))
	}
	for _, id := range "def" {
		notes = append(notes, note(string(id), "cuisine"))
	}
	saved, err := Simulate(notes, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	warm, err := Simulate(append(notes, note("g", "cuisine")), nil, saved)
	if err != nil {
		t.Fatal(err)
	}
	before, after := Fit(saved), Fit(warm)
	for _, id := range "abcdef" {
		if math.Hypot(before[string(id)].X-after[string(id)].X, before[string(id)].Y-after[string(id)].Y) >= 10 {
			t.Errorf("%c moved", id)
		}
	}
	if distance(warm, "g", "#cuisine") >= distance(warm, "g", "#go") {
		t.Error("the new note did not join its tag")
	}
}
