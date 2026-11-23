package grid

// OrderDirection defines ascending or descending sort order across GRID APIs.
type OrderDirection string

const (
	OrderDirectionAsc  OrderDirection = "ASC"
	OrderDirectionDesc OrderDirection = "DESC"
)

// IDFilter matches any of the specified entity IDs.
type IDFilter struct {
	In []string `json:"in,omitempty"`
}

// NullableIDFilter matches a nullable entity ID.
// When present with ID == nil, GRID filters for entities without a relation.
type NullableIDFilter struct {
	ID *string `json:"id"`
}

// StringFilter performs case-insensitive string matching.
type StringFilter struct {
	Contains *string `json:"contains,omitempty"`
	Equals   *string `json:"equals,omitempty"`
}

// BooleanFilter filters by a boolean value.
type BooleanFilter struct {
	Equals *bool `json:"equals,omitempty"`
}

// IntFilter filters by an integer value or inclusive range.
type IntFilter struct {
	Equals *int `json:"equals,omitempty"`
	GTE    *int `json:"gte,omitempty"`
	LTE    *int `json:"lte,omitempty"`
}

// DateTimeFilter filters Central Data results by an ISO-8601 date and time range.
type DateTimeFilter struct {
	GTE *string `json:"gte,omitempty"`
	LTE *string `json:"lte,omitempty"`
}

// NullableDateFilter filters by an ISO-8601 Date range (e.g. "2024-01-01").
type NullableDateFilter struct {
	GTE *string `json:"gte,omitempty"`
	LTE *string `json:"lte,omitempty"`
}
