package respond

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"main.go/app/shared/services/process/createnote"
	"main.go/app/shared/services/process/edittag"
	"main.go/app/shared/services/process/moveitems"
	"main.go/app/shared/services/process/readitem"
)

func TestError(t *testing.T) {
	tests := map[string]struct {
		err    error
		status int
		body   string
	}{
		"bad request": {createnote.ErrNoContent, http.StatusBadRequest, "no note content given"},
		"not found":   {fmt.Errorf("%w: %q", readitem.ErrNotFound, "x"), http.StatusNotFound, "item not found"},
		"conflict":    {edittag.ErrAlreadyExists, http.StatusConflict, "tag already exists"},
		"cycle":       {moveitems.ErrCycle, http.StatusConflict, "cannot move"},
		"unknown":     {errors.New("dial postgres://user:secret@db"), http.StatusInternalServerError, "internal error"},
	}

	for name, test := range tests {
		w := httptest.NewRecorder()
		Error(w, test.err)
		if w.Code != test.status || !strings.Contains(w.Body.String(), test.body) {
			t.Errorf("%s: %d %s, want %d with %q", name, w.Code, w.Body, test.status, test.body)
		}
		if strings.Contains(w.Body.String(), "secret") {
			t.Errorf("%s: the error leaks to the client: %s", name, w.Body)
		}
	}
}

func TestListEmpty(t *testing.T) {
	w := httptest.NewRecorder()
	List[string](w, http.StatusOK, nil)
	if strings.TrimSpace(w.Body.String()) != "[]" {
		t.Errorf("body = %q, want []", w.Body)
	}
}

func TestDecode(t *testing.T) {
	tests := map[string]struct {
		body   string
		ok     bool
		status int
	}{
		"valid":         {`{"title":"a"}`, true, http.StatusOK},
		"unknown field": {`{"titel":"a"}`, false, http.StatusBadRequest},
		"not json":      {`nope`, false, http.StatusBadRequest},
		"empty":         {``, false, http.StatusBadRequest},
		"too large":     {`{"title":"` + strings.Repeat("a", maxBody) + `"}`, false, http.StatusRequestEntityTooLarge},
	}

	for name, test := range tests {
		var dest struct {
			Title string `json:"title"`
		}
		w := httptest.NewRecorder()
		ok := Decode(w, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body)), &dest)
		if ok != test.ok || w.Code != test.status {
			t.Errorf("%s: ok = %v, status %d; want %v, %d", name, ok, w.Code, test.ok, test.status)
		}
	}
}

func TestID(t *testing.T) {
	valid := "3f2a4b1c-0d9e-4a8b-9c7d-1e2f3a4b5c6d"
	for id, want := range map[string]bool{valid: true, "nope": false, "": false, valid + "x": false} {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.SetPathValue("id", id)
		w := httptest.NewRecorder()
		if _, ok := ID(w, r); ok != want {
			t.Errorf("ID(%q) = %v, want %v", id, ok, want)
		}
		if !want && w.Code != http.StatusNotFound {
			t.Errorf("ID(%q): status %d, want 404", id, w.Code)
		}
	}
}

func TestValidIDs(t *testing.T) {
	valid := "3f2a4b1c-0d9e-4a8b-9c7d-1e2f3a4b5c6d"
	if !ValidIDs(httptest.NewRecorder(), valid, "") {
		t.Error("a uuid and an empty id must pass")
	}
	w := httptest.NewRecorder()
	if ValidIDs(w, valid, "nope") || w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}
