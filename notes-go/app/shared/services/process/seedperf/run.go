package seedperf

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/hashtag"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/adduser"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createfolder"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createtag"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
)

// The shape of the account, like the real one that made the app slow (#138): one huge folder, a few small ones,
// 1 to 4 tags per note on a Zipf law (the first on about a quarter of the notes), a [[link]] here and there.
const (
	smallFolders = 30
	bigShare     = 0.96 // of the notes in the big folder
	tagCount     = 150
	zipf         = 0.8 // a tag of rank r is drawn in proportion to 1 / r^zipf
	linkShare    = 0.05
	minContent   = 200
	maxContent   = 5000
	batch        = 1000
	seed         = 138 // the same account on every base
)

// Input is the account to create and how many notes it gets.
type Input struct {
	Name     string
	Password string
	Notes    int
}

type run struct {
	plumprimitive.Process
	cache

	ctx   context.Context
	db    bun.IDB
	input Input
}

type cache struct {
	exists  bool
	ctx     context.Context // acting for the new account
	tx      bun.Tx          // begun with ctx, inside the transaction that created the account
	folders []string        // the big one first
	tags    []string        // by rank, most drawn first
	tagIDs  map[string]string
	notes   []note
}

type note struct {
	item    domain.Item
	title   string
	content string
	tags    []string
	links   []int
}

// Process creates an account filled with generated notes, to measure the app at scale; nothing when the account
// exists. Everything goes in one transaction, written as the app writes it but in bulk: creating the notes one by one
// would resolve the [[links]] of the whole account each time (#145).
func Process(ctx context.Context, db bun.IDB, input Input) *run {
	return &run{
		ctx:   ctx,
		db:    db,
		input: input,
	}
}

func (p *run) Run() {
	p.checkAccount()
	if p.HasErr() || p.cache.exists {
		return
	}

	tx, err := p.db.BeginTx(p.ctx, nil)
	if err != nil {
		p.SetErr(err)
		return
	}
	// Only on a failure: after a commit, rolling a savepoint back (the caller's transaction, in tests) would abort it.
	defer func() {
		if p.HasErr() {
			_ = tx.Rollback()
		}
	}()

	p.createAccount(tx)
	p.createFolders()
	p.createTags()
	p.makeNotes()
	p.insertNotes()
	p.insertContents()
	if p.HasErr() {
		return
	}
	p.SetErr(p.cache.tx.Commit())
	if !p.HasErr() {
		p.SetErr(tx.Commit())
	}
}

// Getters & Setters

// GetResult tells whether the account was created, and with how many notes.
func (p *run) GetResult() (bool, int) {
	return !p.cache.exists, len(p.cache.notes)
}

// Methods

func (p *run) checkAccount() {
	exists, err := p.db.NewSelect().Model((*domain.User)(nil)).Where("lower(name) = lower(?)", p.input.Name).Exists(p.ctx)
	p.cache.exists = exists
	p.SetErr(err)
}

// createAccount makes the account in tx, then begins the rest with a ctx acting for it: a bun.Tx hands its RunInTx
// functions the ctx it began with.
func (p *run) createAccount(tx bun.Tx) {
	add := adduser.Process(p.ctx, tx, adduser.Input{Name: p.input.Name, Password: p.input.Password})
	add.Run()
	if add.HasErr() {
		p.SetErr(add.GetErr())
		return
	}
	id := add.GetResult().ID
	if err := secret.Load(p.ctx, tx, id); err != nil {
		p.SetErr(err)
		return
	}
	p.cache.ctx = domain.WithUser(p.ctx, id)
	savepoint, err := tx.BeginTx(p.cache.ctx, nil)
	p.cache.tx = savepoint
	p.SetErr(err)
}

func (p *run) createFolders() {
	if p.HasErr() {
		return
	}

	titles := []string{"Archives"}
	for i := 1; i <= smallFolders; i++ {
		titles = append(titles, fmt.Sprintf("Dossier %d", i))
	}
	for _, title := range titles {
		folder := createfolder.Process(p.cache.ctx, p.cache.tx, createfolder.Input{Title: title})
		folder.Run()
		if folder.HasErr() {
			p.SetErr(folder.GetErr())
			return
		}
		p.cache.folders = append(p.cache.folders, folder.GetResult().ID)
	}
}

func (p *run) createTags() {
	if p.HasErr() {
		return
	}

	for rank := range tagCount {
		tag := words[rank%len(words)]
		if rank >= len(words) {
			tag = fmt.Sprintf("%s-%d", tag, rank/len(words))
		}
		p.cache.tags = append(p.cache.tags, tag)
	}
	create := createtag.Process(p.cache.ctx, p.cache.tx, p.cache.tags)
	create.Run()
	if create.HasErr() {
		p.SetErr(create.GetErr())
		return
	}

	hashes, err := domain.BlindTags(p.cache.ctx, p.cache.tags)
	if err != nil {
		p.SetErr(err)
		return
	}
	byHash := map[string]string{}
	for _, tag := range create.GetResult() {
		byHash[tag.ValueHash] = tag.ID
	}
	p.cache.tagIDs = map[string]string{}
	for i, tag := range p.cache.tags {
		p.cache.tagIDs[tag] = byHash[hashes[i]]
	}
}

// makeNotes draws every note; the same seed gives the same account.
func (p *run) makeNotes() {
	if p.HasErr() {
		return
	}

	random := rand.New(rand.NewPCG(seed, seed))
	// Cumulated: a draw is a search in it.
	weights := make([]float64, tagCount)
	for rank := range weights {
		weights[rank] = 1 / math.Pow(float64(rank+1), zipf)
		if rank > 0 {
			weights[rank] += weights[rank-1]
		}
	}
	now := time.Now()
	userID := domain.UserID(p.cache.ctx)

	p.cache.notes = make([]note, p.input.Notes)
	for i := range p.cache.notes {
		folder := p.cache.folders[0]
		if random.Float64() > bigShare {
			folder = p.cache.folders[1+random.IntN(smallFolders)]
		}
		n := &p.cache.notes[i]
		n.title = sentence(random, 2+random.IntN(7))
		n.item = domain.Item{
			UserID:    userID,
			ParentID:  &folder,
			Kind:      domain.NOTE,
			CreatedAt: now.Add(-time.Duration(random.Int64N(int64(3 * 365 * 24 * time.Hour)))),
		}
		for range 1 + random.IntN(4) {
			tag := p.cache.tags[draw(random, weights)]
			if !slices.Contains(n.tags, tag) {
				n.tags = append(n.tags, tag)
			}
		}
		if random.Float64() < linkShare {
			for range 1 + random.IntN(2) {
				if target := random.IntN(len(p.cache.notes)); target != i {
					n.links = append(n.links, target)
				}
			}
		}
		// Log-uniform: mostly short notes, a few long ones.
		size := minContent * math.Pow(maxContent/minContent, random.Float64())
		var text strings.Builder
		for text.Len() < int(size) {
			text.WriteString(sentence(random, 6+random.IntN(9)) + ". ")
			if random.IntN(5) == 0 {
				text.WriteString("\n\n")
			}
		}
		n.content = "# " + n.title + "\n\n" + strings.TrimSpace(text.String())
	}
}

// insertNotes writes the items, sealed, by batches: the ids come back for the links.
func (p *run) insertNotes() {
	if p.HasErr() {
		return
	}

	for start := 0; start < len(p.cache.notes); start += batch {
		notes := p.cache.notes[start:min(start+batch, len(p.cache.notes))]
		items := make([]domain.Item, len(notes))
		for i := range notes {
			title, hash, err := domain.SealTitle(p.cache.ctx, notes[i].title)
			if err != nil {
				p.SetErr(err)
				return
			}
			items[i] = notes[i].item
			items[i].Title, items[i].TitleHash = title, hash
		}
		if err := p.cache.tx.NewInsert().Model(&items).Returning("id").Scan(p.cache.ctx); err != nil {
			p.SetErr(err)
			return
		}
		for i := range notes {
			notes[i].item.ID = items[i].ID
		}
	}
}

// insertContents writes the texts, their [[links]] resolved to the ids and their #tags, and the rows that index them.
func (p *run) insertContents() {
	if p.HasErr() {
		return
	}

	var links []domain.NoteLink
	for start := 0; start < len(p.cache.notes); start += batch {
		notes := p.cache.notes[start:min(start+batch, len(p.cache.notes))]
		var contents []domain.MetadataNote
		var tags []domain.NoteTag
		for i := range notes {
			n := &notes[i]
			content := n.content
			seen := map[string]bool{}
			for _, target := range n.links {
				to := p.cache.notes[target]
				if seen[to.item.ID] {
					continue
				}
				seen[to.item.ID] = true
				content += fmt.Sprintf("\n\nVoir aussi [%s](note:%s).", to.title, to.item.ID)
				links = append(links, domain.NoteLink{SourceID: n.item.ID, TargetID: to.item.ID})
			}
			sealed, err := secret.Seal(domain.UserID(p.cache.ctx), hashtag.Sync(content, n.tags))
			if err != nil {
				p.SetErr(err)
				return
			}
			contents = append(contents, domain.MetadataNote{Metadata: domain.Metadata{ItemID: n.item.ID}, Content: sealed})
			for _, tag := range n.tags {
				tags = append(tags, domain.NoteTag{TagID: p.cache.tagIDs[tag], NoteID: n.item.ID})
			}
		}
		for _, rows := range []any{&contents, &tags} {
			if _, err := p.cache.tx.NewInsert().Model(rows).Exec(p.cache.ctx); err != nil {
				p.SetErr(err)
				return
			}
		}
	}

	// Once every text is in: a link may point to a note of a later batch.
	for start := 0; start < len(links); start += batch {
		rows := links[start:min(start+batch, len(links))]
		if _, err := p.cache.tx.NewInsert().Model(&rows).Exec(p.cache.ctx); err != nil {
			p.SetErr(err)
			return
		}
	}
}

// Helpers

// draw picks an index in proportion to its weight, from the cumulated weights.
func draw(random *rand.Rand, cumulated []float64) int {
	return sort.SearchFloat64s(cumulated, random.Float64()*cumulated[len(cumulated)-1])
}

// sentence is n words drawn, the first capitalized.
func sentence(random *rand.Rand, n int) string {
	drawn := make([]string, n)
	for i := range drawn {
		drawn[i] = words[random.IntN(len(words))]
	}
	text := []rune(strings.Join(drawn, " "))
	text[0] = unicode.ToUpper(text[0])
	return string(text)
}

var words = strings.Fields(`projet idée recette voyage lecture réunion code jardin film santé travail famille
	maison budget sport musique cuisine livre article école vacances course marché ville train montagne mer
	rivière forêt chat chien enfant parent ami collègue client facture contrat rapport plan liste note
	question réponse problème solution méthode outil serveur base donnée fichier dossier image photo vidéo
	carte route semaine mois année matin soir nuit printemps été automne hiver pluie soleil neige vent
	histoire science physique chimie langue mot phrase page chapitre auteur titre version test erreur
	tâche objectif priorité rendez-vous appel message courrier compte banque achat vente prix stock
	commande livraison atelier bureau salle cours leçon exercice examen conférence podcast série épisode
	jeu partie équipe match score stratégie analyse mesure résultat graphe tableau schéma modèle`)
