package routes

import (
	"timeful/server/models"
)

// scannedPosterPayload is the poster-scan request body. The struct stays
// handwritten because its JSON tags define the wire shape, which a GALA Codec
// cannot reproduce.
type scannedPosterPayload struct {
	Url      string           `json:"url" binding:"required"`
	Location *models.Location `json:"location"`
}

func newScannedPosterPayload() scannedPosterPayload {
	return scannedPosterPayload{}
}
