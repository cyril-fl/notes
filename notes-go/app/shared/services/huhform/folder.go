package huhform

import (
	"charm.land/huh/v2"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listfolders"
)

// PickFolder asks for a folder in a list of paths ("/" filters it) and returns its id ("" = the root).
func PickFolder(folders []listfolders.Folder) (string, error) {
	id := ""
	picker := huh.NewSelect[string]().Title("Folder").Options(folderOptions(folders)...).Value(&id)
	return id, New(picker).Run()
}

// folderOptions shows paths, the root first; the value is the folder id.
func folderOptions(folders []listfolders.Folder) []huh.Option[string] {
	options := []huh.Option[string]{huh.NewOption("/", "")}
	for _, folder := range folders {
		options = append(options, huh.NewOption(folder.Path, folder.ID))
	}
	return options
}
