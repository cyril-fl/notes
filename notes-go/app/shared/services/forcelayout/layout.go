// Package forcelayout places the notes of a graph and their tags, as d3-force did in the browser: a repulsion between
// every node (Barnes-Hut), springs along the edges, collisions on a small graph, a pull to the center. Deterministic:
// the same graph gives the same map.
package forcelayout

import (
	"encoding/hex"
	"hash/fnv"
	"math"
	"runtime"
	"sync"

	"gonum.org/v1/gonum/spatial/barneshut"
	"gonum.org/v1/gonum/spatial/r2"
)

// Measured on a synthetic account: half the ticks of d3's full cooling, the same clusters, half the time.
const (
	ticks         = 150
	margin        = 8
	collideRadius = 30
	// Collisions keep the titles of a small graph apart, all shown at once; past it they cost more than every other
	// force and the zoom spreads the titles anyway.
	collideLimit = 500
	// Barnes-Hut approximation of the repulsion, coarser than d3's 0.9: a third faster, the same clusters.
	theta        = 1.5
	charge       = -300
	linkDistance = 2 * collideRadius
	// A tag on every note says nothing of them, but must still hold them.
	minWeight = 0.05
	// In % of the view per unit of the simulation: keeps a small graph from spreading to the corners.
	maxScale = 0.35
	// From a saved layout, only what changed settles: a little heat, a few ticks.
	warmAlpha = 0.1
	warmTicks = 40
	// d3's cooling and friction.
	velocityDecay = 0.6
)

var alphaDecay = 1 - math.Pow(0.001, 1.0/300)

// Note is a node of the graph with its tags, each tag a node too.
type Note struct {
	ID   string
	Tags []string
}

// Link is a [[link]] between two notes.
type Link struct {
	Source string
	Target string
}

// Point is a place, in the simulation's units or in % of the view.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Edge is a spring: a note to one of its tags, or a [[link]].
type Edge struct {
	Source   string
	Target   string
	Strength float64
	Distance float64
}

// TagNode is the node of a tag: notes have uuids, no clash.
func TagNode(tag string) string { return "#" + tag }

// Edges gives the springs of the graph, a note to each of its tags (notes in order, their tags in order), then the [[links]].
//   - A tag pulls by how rare it is among the n notes, log(n / its notes) / log(n): a tag on half of them says little.
//     Shared among the tags of the note, so a note with many tags is not pulled harder than one with a single tag.
//   - Its notes sit at √(its notes) collide radii: the room a disk of them needs, so a big tag spreads instead of piling up.
//   - A [[link]] is a choice, not a vote: short and strong, weakened only on a note with many of them.
func Edges(notes []Note, links []Link) []Edge {
	carriers := map[string]int{}
	for _, note := range notes {
		for _, tag := range note.Tags {
			carriers[tag]++
		}
	}
	linked := map[string]int{}
	for _, link := range links {
		linked[link.Source]++
		linked[link.Target]++
	}

	n := float64(len(notes))
	weight := func(count int) float64 {
		if n <= 1 {
			return 1
		}
		return max(math.Log(n/float64(count))/math.Log(n), minWeight)
	}
	var edges []Edge
	for _, note := range notes {
		for _, tag := range note.Tags {
			count := carriers[tag]
			edges = append(edges, Edge{
				Source:   note.ID,
				Target:   TagNode(tag),
				Strength: weight(count) / float64(len(note.Tags)),
				Distance: max(collideRadius*math.Sqrt(float64(count)), linkDistance),
			})
		}
	}
	for _, link := range links {
		edges = append(edges, Edge{
			Source:   link.Source,
			Target:   link.Target,
			Strength: 1 / float64(min(linked[link.Source], linked[link.Target])),
			Distance: linkDistance,
		})
	}
	return edges
}

// Signature is a fingerprint of the notes, their tags and links (FNV-1a): a saved layout serves as is only for the same graph.
func Signature(notes []Note, links []Link) string {
	hash := fnv.New64a()
	for _, note := range notes {
		hash.Write([]byte(note.ID + ":"))
		for _, tag := range note.Tags {
			hash.Write([]byte(tag + ","))
		}
		hash.Write([]byte(";"))
	}
	for _, link := range links {
		hash.Write([]byte(link.Source + ">" + link.Target + ";"))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// node is a body of the simulation, a barneshut.Particle2 of mass 1.
type node struct {
	id     string
	x, y   float64
	vx, vy float64
	placed bool
}

func (n *node) Coord2() r2.Vec { return r2.Vec{X: n.x, Y: n.y} }
func (n *node) Mass() float64  { return 1 }

// Simulate places the notes and their tags, in the simulation's units: see Edges for the springs.
// The pull to the center weakens as the graph grows (1/√ of its nodes), or a big one shrinks into a ball.
// seed: a layout saved before; the nodes it lacks start next to a neighbor it has, and the whole only settles a little.
func Simulate(notes []Note, links []Link, seed map[string]Point) (map[string]Point, error) {
	nodes := make([]*node, 0, len(notes))
	index := map[string]*node{}
	add := func(id string) {
		if index[id] == nil {
			index[id] = &node{id: id}
			nodes = append(nodes, index[id])
		}
	}
	for _, note := range notes {
		add(note.ID)
	}
	for _, note := range notes {
		for _, tag := range note.Tags {
			add(TagNode(tag))
		}
	}
	edges := Edges(notes, links)
	gravity := math.Sqrt(10 / math.Max(float64(len(nodes)), 10))

	if seed != nil {
		near := map[string]Point{}
		for _, edge := range edges {
			if at, ok := seed[edge.Source]; ok {
				if _, done := near[edge.Target]; !done {
					near[edge.Target] = at
				}
			}
			if at, ok := seed[edge.Target]; ok {
				if _, done := near[edge.Source]; !done {
					near[edge.Source] = at
				}
			}
		}
		for _, n := range nodes {
			if at, ok := seed[n.id]; ok {
				n.x, n.y, n.placed = at.X, at.Y, true
			} else if at, ok := near[n.id]; ok {
				// Beside its neighbor, not on it: two nodes at the same point have no direction to push apart.
				n.x, n.y, n.placed = at.X+collideRadius, at.Y+collideRadius, true
			}
		}
	}
	// d3's phyllotaxis for the nodes not placed yet: spread around the center, never two at the same point.
	for i, n := range nodes {
		if !n.placed {
			radius, angle := 10*math.Sqrt(0.5+float64(i)), float64(i)*math.Pi*(3-math.Sqrt(5))
			n.x, n.y = radius*math.Cos(angle), radius*math.Sin(angle)
		}
	}

	s := simulation{nodes: nodes, edges: edges, index: index, random: 1}
	s.bias()
	alpha, count := 1.0, ticks
	if seed != nil {
		alpha, count = warmAlpha, warmTicks
	}
	for range count {
		alpha += -alpha * alphaDecay
		if err := s.charge(alpha); err != nil {
			return nil, err
		}
		s.link(alpha)
		if len(nodes) <= collideLimit {
			s.collide()
		}
		for _, n := range nodes {
			n.vx -= n.x * 0.04 * gravity * alpha
			n.vy -= n.y * 0.08 * gravity * alpha
			n.vx *= velocityDecay
			n.vy *= velocityDecay
			n.x += n.vx
			n.y += n.vy
		}
	}

	positions := make(map[string]Point, len(nodes))
	for _, n := range nodes {
		positions[n.id] = Point{n.x, n.y}
	}
	return positions, nil
}

// Fit scales a simulated layout into % of the view, around its center.
func Fit(positions map[string]Point) map[string]Point {
	if len(positions) == 0 {
		return map[string]Point{}
	}
	minX, maxX, minY, maxY := math.Inf(1), math.Inf(-1), math.Inf(1), math.Inf(-1)
	for _, at := range positions {
		minX, maxX, minY, maxY = min(minX, at.X), max(maxX, at.X), min(minY, at.Y), max(maxY, at.Y)
	}
	fit := func(low, high float64) func(float64) float64 {
		scale := 0.0
		if high > low {
			scale = min((100-2*margin)/(high-low), maxScale)
		}
		return func(v float64) float64 { return 50 + (v-low-(high-low)/2)*scale }
	}
	x, y := fit(minX, maxX), fit(minY, maxY)
	fitted := make(map[string]Point, len(positions))
	for id, at := range positions {
		fitted[id] = Point{x(at.X), y(at.Y)}
	}
	return fitted
}

type simulation struct {
	nodes  []*node
	edges  []Edge
	index  map[string]*node
	biases []float64
	random uint64
}

// jiggle is d3's tiny random nudge for two nodes at the same point, from its seeded LCG.
func (s *simulation) jiggle() float64 {
	s.random = (1664525*s.random + 1013904223) % 4294967296
	return (float64(s.random)/4294967296 - 0.5) * 1e-6
}

// bias splits the pull of an edge between its ends: the end with more edges moves less.
func (s *simulation) bias() {
	count := map[string]int{}
	for _, edge := range s.edges {
		count[edge.Source]++
		count[edge.Target]++
	}
	s.biases = make([]float64, len(s.edges))
	for i, edge := range s.edges {
		s.biases[i] = float64(count[edge.Source]) / float64(count[edge.Source]+count[edge.Target])
	}
}

// charge pushes every node away from the others, in 1/distance as d3's forceManyBody; the far ones as groups.
// The nodes read the tree only, so they are pushed in parallel.
func (s *simulation) charge(alpha float64) error {
	particles := make([]barneshut.Particle2, len(s.nodes))
	for i, n := range s.nodes {
		particles[i] = n
	}
	plane, err := barneshut.NewPlane(particles)
	if err != nil {
		return err
	}

	push := func(_, _ barneshut.Particle2, _, m2 float64, v r2.Vec) r2.Vec {
		l := v.X*v.X + v.Y*v.Y
		if l == 0 {
			return r2.Vec{}
		}
		l = max(l, math.Sqrt(l)) // d3's distanceMin of 1
		return r2.Scale(m2/l, v)
	}
	forces := make([]r2.Vec, len(s.nodes))
	var wait sync.WaitGroup
	workers := runtime.GOMAXPROCS(0)
	for w := range workers {
		wait.Go(func() {
			for i := w; i < len(s.nodes); i += workers {
				forces[i] = plane.ForceOn(s.nodes[i], theta, push)
			}
		})
	}
	wait.Wait()
	for i, n := range s.nodes {
		n.vx += forces[i].X * charge * alpha
		n.vy += forces[i].Y * charge * alpha
	}
	return nil
}

// link pulls the ends of every edge toward its distance, by its strength, as d3's forceLink.
func (s *simulation) link(alpha float64) {
	for i, edge := range s.edges {
		source, target := s.index[edge.Source], s.index[edge.Target]
		if source == nil || target == nil {
			continue
		}
		x := target.x + target.vx - source.x - source.vx
		if x == 0 {
			x = s.jiggle()
		}
		y := target.y + target.vy - source.y - source.vy
		if y == 0 {
			y = s.jiggle()
		}
		l := math.Sqrt(x*x + y*y)
		l = (l - edge.Distance) / l * alpha * edge.Strength
		x, y = x*l, y*l
		b := s.biases[i]
		target.vx -= x * b
		target.vy -= y * b
		source.vx += x * (1 - b)
		source.vy += y * (1 - b)
	}
}

// collide pushes apart two nodes closer than two radii, as d3's forceCollide.
// Every pair, no tree: it only runs up to collideLimit nodes.
func (s *simulation) collide() {
	const r = 2 * collideRadius
	for i, a := range s.nodes {
		for _, b := range s.nodes[i+1:] {
			x := a.x + a.vx - b.x - b.vx
			y := a.y + a.vy - b.y - b.vy
			l := x*x + y*y
			if l >= r*r {
				continue
			}
			if x == 0 {
				x = s.jiggle()
				l += x * x
			}
			if y == 0 {
				y = s.jiggle()
				l += y * y
			}
			l = math.Sqrt(l)
			l = (r - l) / l
			x, y = x*l*0.5, y*l*0.5
			a.vx += x
			a.vy += y
			b.vx -= x
			b.vy -= y
		}
	}
}
