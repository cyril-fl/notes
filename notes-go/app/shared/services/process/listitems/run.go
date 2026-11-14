package listitems

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	plumprimitive "github.com/cyril-fl/plum/primitive"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/driver/pgdriver"
)

// invalidText is the Postgres SQLSTATE raised when the id is not a uuid.
const invalidText = "22P02"

var (
	ErrInvalidID   = errors.New("not a uuid")
	ErrInvalidKind = errors.New("invalid kind")
	ErrInvalidPage = errors.New("invalid limit or offset")
)

type run struct {
	plumprimitive.Process
	cache

	ctx      context.Context
	db       bun.IDB // *bun.DB or bun.Tx: tests run inside a rolled-back transaction
	parentID string
	kind     domain.Kind
	limit    int // 0: every child
	offset   int
}

type cache struct {
	parentID string
	order    []string // every child id, sorted
	result   []domain.Item
	total    int
}

// Process lists the direct children of a folder; an empty parentID lists the root, an empty kind both kinds.
func Process(ctx context.Context, db bun.IDB, parentID string, kind domain.Kind) *run {
	return &run{
		ctx:      ctx,
		db:       db,
		parentID: parentID,
		kind:     kind,
	}
}

// Page keeps limit children from offset, in the same order; GetTotal still counts them all.
func (p *run) Page(limit, offset int) *run {
	p.limit, p.offset = limit, offset
	return p
}

func (p *run) Run() {
	p.checkKind()
	p.checkPage()
	p.makeParentID()
	p.loadOrder()
	p.loadItems()
}

// Getters & Setters

func (p *run) GetResult() []domain.Item {
	return p.cache.result
}

func (p *run) GetTotal() int {
	return p.cache.total
}

// Methods

func (p *run) checkKind() {
	if p.kind != "" && p.kind != domain.FOLDER && p.kind != domain.NOTE {
		p.SetErr(fmt.Errorf("%w: %s", ErrInvalidKind, p.kind))
	}
}

func (p *run) checkPage() {
	if p.limit < 0 || p.offset < 0 {
		p.SetErr(fmt.Errorf("%w: %d, %d", ErrInvalidPage, p.limit, p.offset))
	}
}

func (p *run) makeParentID() {
	p.cache.parentID = strings.TrimSpace(p.parentID)
}

// loadOrder sorts the children folders first, then notes, by title (sealed: ordered here); soft_delete hides the
// trashed ones. The titles come from the cached paths (none decrypted), and the order itself is cached while
// users.items_version stays the same: sorting 48,000 notes takes 140 ms, redone on every page otherwise.
func (p *run) loadOrder() {
	if p.HasErr() {
		return
	}

	// The version before the items, as in domain.Paths: a write committed between the two only makes the next call rebuild.
	userID := domain.UserID(p.ctx)
	var version int64
	if err := p.db.NewSelect().Model((*domain.User)(nil)).Column("items_version").Where("id = ?", userID).Scan(p.ctx, &version); err != nil {
		p.SetErr(err)
		return
	}
	cached, ok := orders.Load(userID)
	if !ok || cached.(*cachedOrders).version != version {
		cached = &cachedOrders{version: version, ids: map[orderKey][]string{}}
		orders.Store(userID, cached)
	}
	account, key := cached.(*cachedOrders), orderKey{p.cache.parentID, p.kind}
	account.mu.Lock()
	p.cache.order, ok = account.ids[key]
	account.mu.Unlock()
	if ok {
		return
	}

	// By id under the same title: the order stays the same from one page to the next.
	var order []domain.Item
	query := p.db.NewSelect().Model(&order).Column("id", "kind").OrderExpr("kind = ? DESC, id", domain.FOLDER)
	if err := p.children(query).Scan(p.ctx); err != nil {
		p.SetErr(toDomainErr(err))
		return
	}
	paths, err := domain.Paths(p.ctx, p.db)
	if err != nil {
		p.SetErr(err)
		return
	}
	// A child's path is its parent's, then "/" and its title; the root's is "".
	prefix := len(paths[p.cache.parentID]) + 1
	domain.SortFold(order, func(i domain.Item) string {
		if path := paths[i.ID]; len(path) >= prefix {
			return path[prefix:]
		}
		return "" // under a trashed folder: no path
	}, func(a, b domain.Item) int {
		return boolOrder(b.Kind == domain.FOLDER) - boolOrder(a.Kind == domain.FOLDER)
	})

	p.cache.order = make([]string, len(order))
	for i, item := range order {
		p.cache.order[i] = item.ID
	}
	account.mu.Lock()
	account.ids[key] = p.cache.order
	account.mu.Unlock()
}

// loadItems reads only the page of the order: a folder may hold 48,000 notes. readonly marks the locked notes,
// for their padlock in the folder; has_folders tells the sidebar a folder opens without reading its children.
func (p *run) loadItems() {
	if p.HasErr() {
		return
	}

	order := p.cache.order
	p.cache.total = len(order)
	paged := p.limit > 0
	if paged {
		order = order[min(p.offset, len(order)):min(p.offset+p.limit, len(order))]
	}
	if len(order) == 0 {
		return
	}

	var items []domain.Item
	query := p.children(p.db.NewSelect().Model(&items)).
		Column("id", "parent_id", "created_at", "updated_at", "kind", "title").
		ColumnExpr("COALESCE((SELECT mn.readonly FROM metadata_notes AS mn WHERE mn.item_id = i.id), false) AS readonly").
		ColumnExpr("EXISTS (SELECT 1 FROM items AS c WHERE c.parent_id = i.id AND c.kind = ? AND c.deleted_at IS NULL) AS has_folders", domain.FOLDER)
	if paged {
		query.Where("id IN (?)", bun.In(order))
	}
	if err := query.Scan(p.ctx); err != nil {
		p.SetErr(err)
		return
	}

	byID := make(map[string]domain.Item, len(items))
	for _, item := range items {
		byID[item.ID] = item
	}
	p.cache.result = make([]domain.Item, 0, len(order))
	for _, id := range order {
		if full, ok := byID[id]; ok { // trashed between the two reads
			p.cache.result = append(p.cache.result, full)
		}
	}
}

// children filters a query on items to the asked children of the account.
func (p *run) children(query *bun.SelectQuery) *bun.SelectQuery {
	query.Where("user_id = ?", domain.UserID(p.ctx))
	if p.cache.parentID == "" {
		query.Where("parent_id IS NULL")
	} else {
		query.Where("parent_id = ?", p.cache.parentID)
	}
	if p.kind != "" {
		query.Where("kind = ?", p.kind)
	}
	return query
}

// orders keeps, by account id, the sorted children ids of the folders listed since items_version last changed:
// a new version drops them all, so an account holds at most about twice its items. Ids only, no title.
var orders sync.Map

type cachedOrders struct {
	version int64
	mu      sync.Mutex
	ids     map[orderKey][]string // shared, do not modify
}

type orderKey struct {
	parentID string
	kind     domain.Kind
}

// Helpers

// toDomainErr turns a parent id that is not a uuid into ErrInvalidID.
func toDomainErr(err error) error {
	var pgErr pgdriver.Error
	if errors.As(err, &pgErr) && pgErr.Field('C') == invalidText {
		return fmt.Errorf("%w: %s", ErrInvalidID, pgErr.Field('M'))
	}
	return err
}

func boolOrder(b bool) int {
	if b {
		return 1
	}
	return 0
}
