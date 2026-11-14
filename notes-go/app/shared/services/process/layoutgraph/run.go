package layoutgraph

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/forcelayout"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readgraph"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readneighbors"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

// saved is what domain.GraphLayout keeps, sealed: the positions in the simulation's units, for the graph of signature.
type saved struct {
	Signature string                       `json:"signature"`
	Positions map[string]forcelayout.Point `json:"positions"`
}

type run struct {
	plumprimitive.Process
	cache

	ctx    context.Context
	db     bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	noteID string
}

type cache struct {
	result    readgraph.Result
	notes     []forcelayout.Note
	links     []forcelayout.Link
	saved     *saved
	positions map[string]forcelayout.Point
	changed   bool
}

// Process lays out a graph on the server, for every device. noteID "": every note, from the account's saved layout,
// computed again only when the graph changed (and then from the old one: only what changed moves), saved for the
// next visit. Otherwise the neighborhood of that note, from scratch.
func Process(ctx context.Context, db bun.IDB, noteID string) *run {
	return &run{
		ctx:    ctx,
		db:     db,
		noteID: noteID,
	}
}

func (p *run) Run() {
	p.loadGraph()
	p.loadSaved()
	p.makeLayout()
	p.saveLayout()
	p.applyLayout()
}

// Getters & Setters

func (p *run) GetResult() readgraph.Result {
	return p.cache.result
}

// Methods

func (p *run) loadGraph() {
	if p.noteID == "" {
		process := readgraph.Process(p.ctx, p.db, true)
		process.Run()
		p.cache.result = process.GetResult()
		p.SetErr(process.GetErr())
	} else {
		process := readneighbors.Process(p.ctx, p.db, p.noteID)
		process.Run()
		p.cache.result = neighborhood(process.GetResult())
		p.SetErr(process.GetErr())
	}

	for _, note := range p.cache.result.Notes {
		p.cache.notes = append(p.cache.notes, forcelayout.Note{ID: note.ID, Tags: note.Tags})
	}
	for _, link := range p.cache.result.Links {
		p.cache.links = append(p.cache.links, forcelayout.Link{Source: link.Source, Target: link.Target})
	}
}

// loadSaved reads the account's layout. A cache: one that no longer opens is computed again.
func (p *run) loadSaved() {
	if p.HasErr() || p.noteID != "" {
		return
	}

	var row domain.GraphLayout
	err := p.db.NewSelect().Model(&row).Where("user_id = ?", domain.UserID(p.ctx)).Scan(p.ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return
	}
	if err != nil {
		p.SetErr(err)
		return
	}
	data, err := secret.OpenBytes(row.UserID, row.Data)
	var layout saved
	if err == nil && json.Unmarshal(data, &layout) == nil {
		p.cache.saved = &layout
	}
}

// makeLayout keeps the saved layout of the same graph, or settles a changed one from it.
func (p *run) makeLayout() {
	if p.HasErr() {
		return
	}

	signature := forcelayout.Signature(p.cache.notes, p.cache.links)
	if p.cache.saved != nil && p.cache.saved.Signature == signature {
		p.cache.positions = p.cache.saved.Positions
		return
	}

	var seed map[string]forcelayout.Point
	if p.cache.saved != nil {
		seed = p.cache.saved.Positions
	}
	positions, err := forcelayout.Simulate(p.cache.notes, p.cache.links, seed)
	p.cache.positions = positions
	p.cache.changed = true
	p.cache.saved = &saved{Signature: signature, Positions: positions}
	p.SetErr(err)
}

// saveLayout keeps the whole graph's new layout for the next visit, from any device.
func (p *run) saveLayout() {
	if p.HasErr() || p.noteID != "" || !p.cache.changed {
		return
	}

	data, err := json.Marshal(p.cache.saved)
	if err != nil {
		p.SetErr(err)
		return
	}
	user := domain.UserID(p.ctx)
	if data, err = secret.SealBytes(user, data); err != nil {
		p.SetErr(err)
		return
	}
	_, err = p.db.NewInsert().Model(&domain.GraphLayout{UserID: user, Data: data}).
		On("CONFLICT (user_id) DO UPDATE").Set("data = EXCLUDED.data").Exec(p.ctx)
	p.SetErr(err)
}

// applyLayout puts each node at its place in % of the view, and each spring's pull on its tag or link.
func (p *run) applyLayout() {
	if p.HasErr() {
		return
	}

	fitted := forcelayout.Fit(p.cache.positions)
	edges := forcelayout.Edges(p.cache.notes, p.cache.links)
	k := 0
	p.cache.result.TagPositions = map[string]forcelayout.Point{}
	for i := range p.cache.result.Notes {
		note := &p.cache.result.Notes[i]
		note.X, note.Y = round(fitted[note.ID].X), round(fitted[note.ID].Y)
		note.TagStrengths = make([]float64, len(note.Tags))
		for j, tag := range note.Tags {
			note.TagStrengths[j] = round(edges[k].Strength)
			k++
			at := fitted[forcelayout.TagNode(tag)]
			p.cache.result.TagPositions[tag] = forcelayout.Point{X: round(at.X), Y: round(at.Y)}
		}
	}
	for i := range p.cache.result.Links {
		p.cache.result.Links[i].Strength = round(edges[k].Strength)
		k++
	}
}

// Helpers

// round keeps 3 decimals: a thousandth of the view stays under a pixel at the deepest zoom, and the JSON of a big graph
// loses a quarter of its weight.
func round(v float64) float64 {
	return math.Round(v*1000) / 1000
}

// neighborhood is the graph around one note: the note with its tags, the notes sharing a tag with it (with those
// tags only), the notes of its [[links]] either way, and those links.
func neighborhood(near readneighbors.Result) readgraph.Result {
	id := near.Note.Item.ID
	center := readgraph.Note{ID: id, Title: near.Note.Item.Title, Path: near.Note.Path, Tags: near.Note.Tags}
	result := readgraph.Result{Notes: []readgraph.Note{center}}
	seen := map[string]bool{id: true}
	for _, neighbor := range near.Neighbors {
		seen[neighbor.ID] = true
		result.Notes = append(result.Notes, readgraph.Note{ID: neighbor.ID, Title: neighbor.Title, Path: neighbor.Path, Tags: neighbor.Tags})
	}
	linked := map[string]bool{}
	for _, note := range append(append([]readneighbors.Linked{}, near.Links...), near.Backlinks...) {
		if note.ID == id || linked[note.ID] {
			continue
		}
		if !seen[note.ID] {
			seen[note.ID] = true
			result.Notes = append(result.Notes, readgraph.Note{ID: note.ID, Title: note.Title, Path: note.Path})
		}
		linked[note.ID] = true
		result.Links = append(result.Links, readgraph.Link{Source: id, Target: note.ID})
	}
	return result
}
