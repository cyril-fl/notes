package readversions

import (
	"net/http"

	"github.com/cyril-fl/notes/notes-go/app/api/services/respond"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/keepconflict"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listversions"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readversion"
	"github.com/uptrace/bun"
)

// Handler lists the older texts of a note, without the texts: GET /api/notes/{id}/versions.
// @Summary  List the versions of a note
// @Tags     notes
// @Produce  json
// @Param    id path string true "uuid"
// @Success  200 {array} domain.NoteVersion
// @Failure  404 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/notes/{id}/versions [get]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := respond.ID(w, r)
		if !ok {
			return
		}

		process := listversions.Process(r.Context(), db, id)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.List(w, http.StatusOK, process.GetResult())
	}
}

// Read answers one version of a note with its text: GET /api/notes/{id}/versions/{version} (#99).
// @Summary  Read a version of a note
// @Tags     notes
// @Produce  json
// @Param    id path string true "uuid"
// @Param    version path string true "uuid"
// @Success  200 {object} domain.NoteVersion
// @Failure  400 {object} respond.Failure
// @Failure  404 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/notes/{id}/versions/{version} [get]
func Read(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := respond.ID(w, r)
		if !ok {
			return
		}
		version := r.PathValue("version")
		if !respond.ValidIDs(w, version) {
			return
		}

		process := readversion.Process(r.Context(), db, id, version)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.JSON(w, http.StatusOK, process.GetResult())
	}
}

// Body is the text an edit could not save.
type Body struct {
	Content string `json:"content"`
}

// Keep keeps a text refused by a 409 as a version marked conflict: POST /api/notes/{id}/versions (#78).
// It answers 204 when the text is the note's: nothing to keep.
// @Summary  Keep a text refused by a conflict as a version
// @Tags     notes
// @Accept   json
// @Produce  json
// @Param    id path string true "uuid"
// @Param    body body Body true "body"
// @Success  201 {object} domain.NoteVersion
// @Success  204
// @Failure  400 {object} respond.Failure
// @Failure  404 {object} respond.Failure
// @Failure  413 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/notes/{id}/versions [post]
func Keep(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := respond.ID(w, r)
		if !ok {
			return
		}

		var body Body
		if !respond.Decode(w, r, &body) {
			return
		}

		process := keepconflict.Process(r.Context(), db, id, body.Content)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		if process.GetResult().ID == "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		respond.JSON(w, http.StatusCreated, process.GetResult())
	}
}
