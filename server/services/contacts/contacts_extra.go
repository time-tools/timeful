package contacts

import "timeful/server/errs"

// The People API response shapes stay handwritten because their json tags
// define the wire shape that Decode depends on.
type contactsPersonName struct {
	FamilyName string `json:"familyName"`
	GivenName  string `json:"givenName"`
}

type contactsPersonPhoto struct {
	Url string `json:"url"`
}

type contactsPersonEmail struct {
	Value string `json:"value"`
}

type contactsPerson struct {
	Names          []contactsPersonName  `json:"names"`
	Photos         []contactsPersonPhoto `json:"photos"`
	EmailAddresses []contactsPersonEmail `json:"emailAddresses"`
}

type contactsResult struct {
	Person contactsPerson `json:"person"`
}

type contactsSearchResponse struct {
	Results []contactsResult     `json:"results"`
	Error   *errs.GoogleAPIError `json:"error"`
}

type contactsDirectoryResponse struct {
	People []contactsPerson     `json:"people"`
	Error  *errs.GoogleAPIError `json:"error"`
}
