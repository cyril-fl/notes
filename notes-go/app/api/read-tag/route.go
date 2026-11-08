package readtag

import (
	"net/http"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listtag"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readtag"
	"github.com/uptrace/bun"
)

// Handler shows a tag and the notes carrying it: GET /api/tags/{value}.
// @Summary  Read a tag and the notes carrying it
// @Tags     tags
// @Produce  json
// @Param    value path string true "tag"
// @Success  200 {object} readtag.Result
// @Failure  404 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/tags/{value} [get]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		process := readtag.Process(r.Context(), db, r.PathValue("value"))
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		tag := process.GetResult()
		tag.Notes = append([]readtag.Note{}, tag.Notes...)
		respond.JSON(w, http.StatusOK, tag)
	}
}

// List gives every tag: GET /api/tags.
// @Summary  List every tag
// @Tags     tags
// @Produce  json
// @Success  200 {array} domain.Tag
// @Failure  500 {object} respond.Failure
// @Router   /api/tags [get]
func List(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		process := listtag.Process(r.Context(), db)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.List(w, http.StatusOK, process.GetResult())
	}
}
