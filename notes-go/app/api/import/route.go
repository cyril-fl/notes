package importzip

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/uptrace/bun"
	"main.go/app/api/services/respond"
	"main.go/app/shared/services/process/importnotes"
)

const (
	// maxZip limits the upload, maxContent what it unpacks to: a zip of text, these are far beyond real notes.
	maxZip     = 50 << 20
	maxContent = 200 << 20
)

// Handler adds the .md files of a zip as notes, its directories as folders, into a folder (the root without
// parent), all or nothing: POST /api/import?parent= with the zip as body.
// @Summary  Import markdown files from a zip
// @Tags     transfer
// @Accept   application/zip
// @Produce  json
// @Param    parent query string false "folder uuid, the root when left out"
// @Param    body body []byte true "zip of .md files"
// @Success  200 {object} importnotes.Result
// @Failure  400 {object} respond.Failure
// @Failure  409 {object} respond.Failure
// @Failure  413 {object} respond.Failure
// @Failure  500 {object} respond.Failure
// @Router   /api/import [post]
func Handler(db bun.IDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parent := r.URL.Query().Get("parent")
		if !respond.ValidIDs(w, parent) {
			return
		}

		data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxZip))
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			respond.JSON(w, http.StatusRequestEntityTooLarge, respond.Failure{Error: "request body too large"})
			return
		}
		if err != nil {
			respond.Error(w, err)
			return
		}

		archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			respond.JSON(w, http.StatusBadRequest, respond.Failure{Error: "not a zip file"})
			return
		}
		var size uint64
		for _, file := range archive.File {
			size += file.UncompressedSize64
		}
		if size > maxContent {
			respond.JSON(w, http.StatusRequestEntityTooLarge, respond.Failure{Error: "zip content too large"})
			return
		}

		process := importnotes.Process(r.Context(), db, importnotes.Input{Dir: archive, ParentID: parent})
		process.Run()
		if process.HasErr() {
			respond.Error(w, process.GetErr())
			return
		}
		respond.JSON(w, http.StatusOK, process.GetResult())
	}
}
