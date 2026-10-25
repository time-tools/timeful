package routes

import pgstore "timeful/server/postgres"

// archiveEventInput is the archive request body. The struct stays handwritten
// because its JSON tags define the wire shape, which a GALA Codec cannot
// reproduce.
type archiveEventInput struct {
	Archive *bool `json:"archive" binding:"required"`
}

func newArchiveEventInput() archiveEventInput {
	return archiveEventInput{}
}

// replaceEvent copies one event value through its pointer. The helper keeps the
// .gala file off an assignment through a dereferenced pointer.
func replaceEvent(target *pgstore.Event, source *pgstore.Event) {
	*target = *source
}
