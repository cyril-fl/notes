package respond

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/edittag"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/moveitems"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
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

func TestGzip(t *testing.T) {
	for _, encoding := range []string{"", "gzip, deflate, br"} {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Accept-Encoding", encoding)
		w := httptest.NewRecorder()
		Gzip(w, r, http.StatusOK, []string{"a"})

		body := io.Reader(w.Body)
		if encoding != "" {
			zr, err := gzip.NewReader(w.Body)
			if err != nil {
				t.Fatalf("%q: not gzip: %v", encoding, err)
			}
			body = zr
		}
		data, _ := io.ReadAll(body)
		if strings.TrimSpace(string(data)) != `["a"]` || w.Header().Get("Vary") != "Accept-Encoding" {
			t.Errorf("%q: body = %q, vary = %q", encoding, data, w.Header().Get("Vary"))
		}
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

func TestPage(t *testing.T) {
	items := []int{0, 1, 2, 3, 4}
	tests := map[string]struct {
		query  string
		ok     bool
		result []int
	}{
		"every item":   {"", true, items},
		"first page":   {"?limit=2", true, []int{0, 1}},
		"middle page":  {"?limit=2&offset=2", true, []int{2, 3}},
		"last page":    {"?limit=2&offset=4", true, []int{4}},
		"past the end": {"?limit=2&offset=9", true, []int{}},
		"huge numbers": {"?limit=9223372036854775807&offset=3", true, []int{3, 4}},
		"negative":     {"?limit=-1", false, nil},
		"not a number": {"?offset=a", false, nil},
	}

	for name, test := range tests {
		w := httptest.NewRecorder()
		result, ok := Page(w, httptest.NewRequest(http.MethodGet, "/"+test.query, nil), items)
		if ok != test.ok || fmt.Sprint(result) != fmt.Sprint(test.result) {
			t.Errorf("%s: %v, %v; want %v, %v", name, result, ok, test.result, test.ok)
		}
		if ok && w.Header().Get("X-Total-Count") != "5" {
			t.Errorf("%s: X-Total-Count = %q, want 5", name, w.Header().Get("X-Total-Count"))
		}
		if !ok && w.Code != http.StatusBadRequest {
			t.Errorf("%s: status %d, want 400", name, w.Code)
		}
	}
}
