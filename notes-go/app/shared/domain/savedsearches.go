package domain

import (
	"encoding/json"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
)

// SavedSearch is a smart folder: the criteria are replayed by runsavedsearch, never stored as a result.
type SavedSearch struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	TagIDs  []string `json:"tag_ids,omitempty" binding:"optional"`
	TagMode string   `json:"tag_mode,omitempty" binding:"optional"` // "any" (default) or "all", one of TagModes

	FolderIDs  []string `json:"folder_ids,omitempty" binding:"optional"`
	FolderMode string   `json:"folder_mode,omitempty" binding:"optional"` // "include" or "exclude", one of FolderModes

	CreatedFrom *time.Time `json:"created_from,omitempty" binding:"optional"`
	CreatedTo   *time.Time `json:"created_to,omitempty" binding:"optional"`
	EditedFrom  *time.Time `json:"edited_from,omitempty" binding:"optional"`
	EditedTo    *time.Time `json:"edited_to,omitempty" binding:"optional"`

	Text string `json:"text,omitempty" binding:"optional"`
}

// TagModes and FolderModes are the choices of SavedSearch, the first one its default.
var (
	TagModes    = []string{"any", "all"}
	FolderModes = []string{"include", "exclude"}
)

// SavedSearches are kept sealed, as one JSON string in their jsonb column: write Seal's.
type SavedSearches []SavedSearch

// Seal gives the saved searches sealed by the key of the account, as the JSON string to write.
func (s SavedSearches) Seal(userID string) (string, error) {
	if s == nil {
		s = SavedSearches{}
	}
	data, err := json.Marshal([]SavedSearch(s))
	if err != nil {
		return "", err
	}
	sealed, err := secret.Seal(userID, string(data))
	if err != nil {
		return "", err
	}
	data, err = json.Marshal(sealed)
	return string(data), err
}

// Scan also reads a JSON array in clear: the column's default, [].
func (s *SavedSearches) Scan(src any) error {
	var data []byte
	switch src := src.(type) {
	case string:
		data = []byte(src)
	case []byte:
		data = src
	}
	var sealed string
	if json.Unmarshal(data, &sealed) == nil {
		text, err := secret.Open(sealed)
		if err != nil {
			return err
		}
		data = []byte(text)
	}
	return json.Unmarshal(data, (*[]SavedSearch)(s))
}
