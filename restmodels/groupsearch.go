package restmodels

// GroupSearchResult is a single match from the group name fuzzy search
// endpoint. Relevance is a similarity score in [0,1] (1 being an exact
// match), formatted as a string.
type GroupSearchResult struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Relevance string `json:"relevance"`
}
