package restmodels

// Pagination describes an offset/limit window over a result set.
// A zero value (Limit == 0) means "unbounded" and is used for internal
// lookups that must not be truncated (e.g. exact-match single-row fetches).
type Pagination struct {
	Offset int
	Limit  int
}
