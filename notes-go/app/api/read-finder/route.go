package readfinder

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listitems"
	"github.com/uptrace/bun"
)

// Handler lists the direct children of a folder, of the root without parent: GET /api/items?parent=&kind=&limit=&offset=.
// X-Total-Count counts them all, whatever the page.
// @Summary  List the direct children of a folder, or of the root
// @Tags     items
// @Produce  json
// @Param    parent query string  false "folder uuid, the root when left out"
// @Param    kind   query string  false "folder or note, both when left out"
// @Param    limit  query integer false "children in the page, all when left out or 0"
// @Param    offset query integer false "children skipped before the page, folders first then by title"
// @Success  200 {array} domain.Item
// @Header   200 {integer} X-Total-Count "every child, not only the page"
// @Failure  400 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/items [get]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parent := r.URL.Query().Get("parent")
		if !respond.ValidIDs(w, parent) {
			return
		}

		limit, errLimit := number(r.URL.Query().Get("limit"))
		offset, errOffset := number(r.URL.Query().Get("offset"))
		if errLimit != nil || errOffset != nil {
			respond.Error(w, fmt.Errorf("%w: not a number", listitems.ErrInvalidPage))
			return
		}

		process := listitems.Process(r.Context(), db, parent, domain.Kind(r.URL.Query().Get("kind"))).Page(limit, offset)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		w.Header().Set("X-Total-Count", strconv.Itoa(process.GetTotal()))
		respond.List(w, http.StatusOK, process.GetResult())
	}
}

// number reads a query number, 0 when left out.
func number(value string) (int, error) {
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}
