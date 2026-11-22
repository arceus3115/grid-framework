package grid

// Central Data enums documented in GRID-API-DATA.md.

type SeriesOrderBy string

const (
	SeriesOrderByID                 SeriesOrderBy = "ID"
	SeriesOrderByStartTimeScheduled SeriesOrderBy = "StartTimeScheduled"
	SeriesOrderByUpdatedAt          SeriesOrderBy = "UpdatedAt"
)

type SeriesType string

const (
	SeriesTypeCompetitive SeriesType = "COMPETITIVE"
	SeriesTypeEsports     SeriesType = "ESPORTS"
	SeriesTypeLoopfeed    SeriesType = "LOOPFEED"
	SeriesTypeScrim       SeriesType = "SCRIM"
)

type SeriesWorkflowStatus string

const (
	SeriesWorkflowStatusPublished SeriesWorkflowStatus = "PUBLISHED"
	SeriesWorkflowStatusDeleted   SeriesWorkflowStatus = "DELETED"
)

type TeamWorkflowStatus string

const (
	TeamWorkflowStatusPublished TeamWorkflowStatus = "PUBLISHED"
	TeamWorkflowStatusDeleted   TeamWorkflowStatus = "DELETED"
)

type PlayerType string

const (
	PlayerTypeEsports     PlayerType = "ESPORTS"
	PlayerTypeCompetitive PlayerType = "COMPETITIVE"
)

type ServiceLevel string

const (
	ServiceLevelFull    ServiceLevel = "FULL"
	ServiceLevelLimited ServiceLevel = "LIMITED"
	ServiceLevelNone    ServiceLevel = "NONE"
)

type TournamentType string

const (
	TournamentTypeLeague      TournamentType = "LEAGUE"
	TournamentTypeUnspecified TournamentType = "UNSPECIFIED"
)

type TournamentVenueType string

const (
	TournamentVenueTypeHybrid  TournamentVenueType = "HYBRID"
	TournamentVenueTypeLAN     TournamentVenueType = "LAN"
	TournamentVenueTypeOnline  TournamentVenueType = "ONLINE"
	TournamentVenueTypeUnknown TournamentVenueType = "UNKNOWN"
)

type ContentCatalogEntityType string

const (
	ContentCatalogEntityTypeCharacter ContentCatalogEntityType = "CHARACTER"
	ContentCatalogEntityTypeItem      ContentCatalogEntityType = "ITEM"
	ContentCatalogEntityTypeMap       ContentCatalogEntityType = "MAP"
	ContentCatalogEntityTypeStructure ContentCatalogEntityType = "STRUCTURE"
)

type ContentCatalogVersionOrderField string

const (
	ContentCatalogVersionOrderFieldPublishedOn ContentCatalogVersionOrderField = "PUBLISHED_ON"
	ContentCatalogVersionOrderFieldTitle       ContentCatalogVersionOrderField = "TITLE"
)

// Central Data Filter Inputs.

type TitleFilter struct {
	Private *BooleanFilter `json:"private,omitempty"`
}

type ProductServiceLevelFilter struct {
	ProductName  string       `json:"productName,omitempty"`
	ServiceLevel ServiceLevel `json:"serviceLevel,omitempty"`
}

type SeriesTournamentFilter struct {
	EndDate         *NullableDateFilter `json:"endDate,omitempty"`
	ID              *IDFilter           `json:"id,omitempty"`
	IncludeChildren *BooleanFilter      `json:"includeChildren,omitempty"`
	Name            *StringFilter       `json:"name,omitempty"`
	NameShortened   *StringFilter       `json:"nameShortened,omitempty"`
	StartDate       *NullableDateFilter `json:"startDate,omitempty"`
}

type SeriesPlayerFilter struct {
	Age *IntFilter `json:"age,omitempty"`
}

type SeriesFilter struct {
	LivePlayerIDs        *IDFilter                  `json:"livePlayerIds,omitempty"`
	Players              *SeriesPlayerFilter        `json:"players,omitempty"`
	Private              *BooleanFilter             `json:"private,omitempty"`
	ProductServiceLevels *ProductServiceLevelFilter `json:"productServiceLevels,omitempty"`
	StartTimeScheduled   *DateTimeFilter            `json:"startTimeScheduled,omitempty"`
	TeamID               *string                    `json:"teamId,omitempty"`
	TeamIDs              *IDFilter                  `json:"teamIds,omitempty"`
	TitleID              *string                    `json:"titleId,omitempty"`
	TitleIDs             *IDFilter                  `json:"titleIds,omitempty"`
	Tournament           *SeriesTournamentFilter    `json:"tournament,omitempty"`
	TournamentID         *string                    `json:"tournamentId,omitempty"`
	TournamentIDs        *IDFilter                  `json:"tournamentIds,omitempty"`
	Type                 *SeriesType                `json:"type,omitempty"`
	Types                []SeriesType               `json:"types,omitempty"`
	UpdatedAt            *DateTimeFilter            `json:"updatedAt,omitempty"`
	WorkflowStatuses     []SeriesWorkflowStatus     `json:"workflowStatuses,omitempty"`
}

type TeamFilter struct {
	Name             *StringFilter        `json:"name,omitempty"`
	NameShortened    *StringFilter        `json:"nameShortened,omitempty"`
	OrganizationID   *string              `json:"organizationId,omitempty"`
	Private          *BooleanFilter       `json:"private,omitempty"`
	TitleID          *string              `json:"titleId,omitempty"`
	UpdatedAt        *DateTimeFilter      `json:"updatedAt,omitempty"`
	WorkflowStatuses []TeamWorkflowStatus `json:"workflowStatuses,omitempty"`
}

type NationalityFilter struct {
	Code *StringFilter `json:"code,omitempty"`
	Name *StringFilter `json:"name,omitempty"`
}

type PlayerRoleTitleFilter struct {
	ID *IDFilter `json:"id,omitempty"`
}

type PlayerPlayerRoleFilter struct {
	Name  *StringFilter          `json:"name,omitempty"`
	Title *PlayerRoleTitleFilter `json:"title,omitempty"`
}

type PlayerRoleFilter struct {
	Name  *StringFilter          `json:"name,omitempty"`
	Title *PlayerRoleTitleFilter `json:"title,omitempty"`
}

type PlayerFilter struct {
	Age          *IntFilter              `json:"age,omitempty"`
	FullName     *StringFilter           `json:"fullName,omitempty"`
	Nationality  *NationalityFilter      `json:"nationality,omitempty"`
	Nickname     *StringFilter           `json:"nickname,omitempty"`
	Private      *BooleanFilter          `json:"private,omitempty"`
	Roles        *PlayerPlayerRoleFilter `json:"roles,omitempty"`
	TeamIDFilter *NullableIDFilter       `json:"teamIdFilter,omitempty"`
	TitleID      *string                 `json:"titleId,omitempty"`
	Types        []PlayerType            `json:"types,omitempty"`
	UpdatedAt    *DateTimeFilter         `json:"updatedAt,omitempty"`
}

type TournamentTitleFilter struct {
	ID *IDFilter `json:"id,omitempty"`
}

type TournamentFilter struct {
	EndDate              *NullableDateFilter        `json:"endDate,omitempty"`
	HasChildren          *BooleanFilter             `json:"hasChildren,omitempty"`
	HasParent            *BooleanFilter             `json:"hasParent,omitempty"`
	Name                 *StringFilter              `json:"name,omitempty"`
	NameShortened        *StringFilter              `json:"nameShortened,omitempty"`
	Private              *BooleanFilter             `json:"private,omitempty"`
	ProductServiceLevels *ProductServiceLevelFilter `json:"productServiceLevels,omitempty"`
	StartDate            *NullableDateFilter        `json:"startDate,omitempty"`
	Title                *TournamentTitleFilter     `json:"title,omitempty"`
	TitleID              *string                    `json:"titleId,omitempty"`
	UpdatedAt            *DateTimeFilter            `json:"updatedAt,omitempty"`
}

type OrganizationFilter struct {
	Name      *StringFilter   `json:"name,omitempty"`
	Private   *BooleanFilter  `json:"private,omitempty"`
	UpdatedAt *DateTimeFilter `json:"updatedAt,omitempty"`
}

type ContentCatalogEntityFilter struct {
	EntityType []ContentCatalogEntityType `json:"entityType,omitempty"`
	Name       *StringFilter              `json:"name,omitempty"`
	Private    *BooleanFilter             `json:"private,omitempty"`
	TitleID    *string                    `json:"titleId,omitempty"`
	UpdatedAt  *DateTimeFilter            `json:"updatedAt,omitempty"`
}

type ContentCatalogVersionFilter struct {
	Name    *StringFilter  `json:"name,omitempty"`
	Private *BooleanFilter `json:"private,omitempty"`
	TitleID *string        `json:"titleId,omitempty"`
}

type ContentCatalogVersionOrder struct {
	Field     ContentCatalogVersionOrderField `json:"field"`
	Direction OrderDirection                  `json:"direction"`
}

// Central Data Domain Models.

type DataProvider struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type ExternalEntity struct {
	ID string `json:"id"`
}

type ExternalLink struct {
	DataProvider   DataProvider   `json:"dataProvider"`
	ExternalEntity ExternalEntity `json:"externalEntity"`
}

type Title struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	NameShortened string `json:"nameShortened,omitempty"`
	LogoURL       string `json:"logoUrl,omitempty"`
	Private       *bool  `json:"private,omitempty"`
}

type SeriesFormat struct {
	ID            string `json:"id,omitempty"`
	Name          string `json:"name"`
	NameShortened string `json:"nameShortened"`
}

type Organization struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Private bool   `json:"private,omitempty"`
}

type Nationality struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type PlayerRole struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Private bool   `json:"private,omitempty"`
	Title   *Title `json:"title,omitempty"`
}

type Team struct {
	ID             string             `json:"id"`
	Name           string             `json:"name"`
	NameShortened  *string            `json:"nameShortened,omitempty"`
	ColorPrimary   string             `json:"colorPrimary,omitempty"`
	ColorSecondary string             `json:"colorSecondary,omitempty"`
	LogoURL        string             `json:"logoUrl,omitempty"`
	Rating         *float64           `json:"rating,omitempty"`
	Private        bool               `json:"private,omitempty"`
	Title          *Title             `json:"title,omitempty"`
	Organization   *Organization      `json:"organization,omitempty"`
	ExternalLinks  []ExternalLink     `json:"externalLinks,omitempty"`
	UpdatedAt      string             `json:"updatedAt,omitempty"`
	WorkflowStatus TeamWorkflowStatus `json:"workflowStatus,omitempty"`
}

type TeamParticipant struct {
	BaseInfo       *Team `json:"baseInfo,omitempty"`
	ScoreAdvantage int   `json:"scoreAdvantage,omitempty"`
}

type Player struct {
	ID            string         `json:"id"`
	Nickname      string         `json:"nickname"`
	FullName      *string        `json:"fullName,omitempty"`
	Age           *int           `json:"age,omitempty"`
	ImageURL      string         `json:"imageUrl,omitempty"`
	Private       bool           `json:"private,omitempty"`
	Type          PlayerType     `json:"type,omitempty"`
	Nationality   []Nationality  `json:"nationality,omitempty"`
	Roles         []PlayerRole   `json:"roles,omitempty"`
	Team          *Team          `json:"team,omitempty"`
	Title         *Title         `json:"title,omitempty"`
	ExternalLinks []ExternalLink `json:"externalLinks,omitempty"`
	UpdatedAt     string         `json:"updatedAt,omitempty"`
}

type Tournament struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	NameShortened  string              `json:"nameShortened,omitempty"`
	LogoURL        string              `json:"logoUrl,omitempty"`
	StartDate      *string             `json:"startDate,omitempty"`
	EndDate        *string             `json:"endDate,omitempty"`
	Private        bool                `json:"private,omitempty"`
	TournamentType TournamentType      `json:"tournamentType,omitempty"`
	VenueType      TournamentVenueType `json:"venueType,omitempty"`
	Parent         *Tournament         `json:"parent,omitempty"`
	Children       []Tournament        `json:"children,omitempty"`
	Titles         []Title             `json:"titles,omitempty"`
	Teams          []Team              `json:"teams,omitempty"`
	ExternalLinks  []ExternalLink      `json:"externalLinks,omitempty"`
	UpdatedAt      string              `json:"updatedAt,omitempty"`
}

type ProductServiceLevel struct {
	ProductName  string       `json:"productName"`
	ServiceLevel ServiceLevel `json:"serviceLevel"`
}

type Series struct {
	ID                   string                `json:"id"`
	StartTimeScheduled   string                `json:"startTimeScheduled,omitempty"`
	Type                 SeriesType            `json:"type,omitempty"`
	Private              bool                  `json:"private,omitempty"`
	UpdatedAt            string                `json:"updatedAt,omitempty"`
	WorkflowStatus       SeriesWorkflowStatus  `json:"workflowStatus,omitempty"`
	Title                *Title                `json:"title,omitempty"`
	Format               *SeriesFormat         `json:"format,omitempty"`
	Tournament           *Tournament           `json:"tournament,omitempty"`
	Teams                []TeamParticipant     `json:"teams,omitempty"`
	Players              []Player              `json:"players,omitempty"`
	ProductServiceLevels []ProductServiceLevel `json:"productServiceLevels,omitempty"`
	ExternalLinks        []ExternalLink        `json:"externalLinks,omitempty"`
}

type ContentCatalogVersion struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PublishedOn string `json:"publishedOn,omitempty"`
	Private     *bool  `json:"private,omitempty"`
	Title       *Title `json:"title,omitempty"`
}

type ContentCatalogEntity struct {
	ID                    string                 `json:"id"`
	Name                  string                 `json:"name"`
	ImageURL              string                 `json:"imageUrl,omitempty"`
	Private               *bool                  `json:"private,omitempty"`
	UpdatedAt             string                 `json:"updatedAt,omitempty"`
	ContentCatalogVersion *ContentCatalogVersion `json:"contentCatalogVersion,omitempty"`
}
