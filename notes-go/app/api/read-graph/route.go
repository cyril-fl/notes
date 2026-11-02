package readgraph

import (
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/services/process/counttags"
	"main.go/app/shared/services/process/readneighbors"
)

// Handler gives every tag with how many notes carry it, most used first: GET /api/graph.
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		process := counttags.Process(r.Context(), db)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.List(w, http.StatusOK, process.GetResult())
	}
}

// Neighbors gives a note and the notes sharing its tags: GET /api/notes/{id}/neighbors.
func Neighbors(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := respond.ID(w, r)
		if !ok {
			return
		}

		process := readneighbors.Process(r.Context(), db, id)
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		result := process.GetResult()
		result.Note.Tags = append([]string{}, result.Note.Tags...)
		result.Neighbors = append([]readneighbors.Neighbor{}, result.Neighbors...)
		respond.JSON(w, http.StatusOK, result)
	}
}
