package grid

// Composable GraphQL Selection presets for Central Data entities.
// Callers can use these directly, combine them with Merge(...), or extend them with .With(...).

var (
	TitleSelectionBasic = Scalars("id", "name", "nameShortened")
	TitleSelectionFull  = TitleSelectionBasic.With(Scalar("private"))

	SeriesFormatSelectionBasic = Scalars("id", "name", "nameShortened")

	OrganizationSelectionBasic = Scalars("id", "name")
	OrganizationSelectionFull  = OrganizationSelectionBasic.With(Scalar("private"))

	PlayerRoleSelectionBasic = Scalars("id", "name")
	PlayerRoleSelectionFull  = PlayerRoleSelectionBasic.With(
		Scalar("private"),
		FieldSelect("title", TitleSelectionBasic),
	)

	ExternalLinksSelection = Field("externalLinks",
		Field("dataProvider", Scalar("name")),
		Field("externalEntity", Scalar("id")),
	)

	TeamSelectionBasic = Scalars("id", "name", "nameShortened")
	TeamSelectionFull  = TeamSelectionBasic.With(
		Scalar("colorPrimary"),
		Scalar("colorSecondary"),
		Scalar("logoUrl"),
		Scalar("private"),
		Scalar("updatedAt"),
		FieldSelect("title", TitleSelectionBasic),
		FieldSelect("organization", OrganizationSelectionBasic),
	)

	PlayerSelectionBasic = Scalars("id", "nickname")
	PlayerSelectionFull  = PlayerSelectionBasic.With(
		Scalar("updatedAt"),
		FieldSelect("roles", PlayerRoleSelectionBasic),
		FieldSelect("team", TeamSelectionBasic),
		FieldSelect("title", TitleSelectionBasic),
	)

	TournamentSelectionBasic = Scalars("id", "name", "nameShortened")
	TournamentSelectionFull  = TournamentSelectionBasic.With(
		Scalar("logoUrl"),
		Scalar("startDate"),
		Scalar("endDate"),
		Scalar("tournamentType"),
		Scalar("venueType"),
		Scalar("updatedAt"),
		FieldSelect("titles", TitleSelectionBasic),
	)

	SeriesSelectionBasic       = Scalars("id", "startTimeScheduled", "type")
	SeriesSelectionWithMatchup = SeriesSelectionBasic.With(
		FieldSelect("title", TitleSelectionBasic),
		FieldSelect("format", SeriesFormatSelectionBasic),
		FieldSelect("tournament", TournamentSelectionBasic),
		Field("teams",
			Scalar("scoreAdvantage"),
			FieldSelect("baseInfo", TeamSelectionBasic),
		),
	)
	SeriesSelectionFull = SeriesSelectionWithMatchup.With(
		Scalar("private"),
		Scalar("updatedAt"),
		Scalar("workflowStatus"),
		FieldSelect("players", PlayerSelectionBasic),
		Field("productServiceLevels", Scalar("productName"), Scalar("serviceLevel")),
		ExternalLinksSelection,
	)

	DataProviderSelectionBasic = Scalars("name", "description")

	ContentCatalogVersionSelectionBasic = Scalars("id", "name", "publishedOn")
	ContentCatalogEntitySelectionBasic  = Scalars("id", "name", "imageUrl", "updatedAt")
)

// Default typed return Forms for Central Data queries.

func DefaultTitleForm() Form[Title] {
	return NewForm[Title](TitleSelectionFull)
}

func DefaultTitlesForm() Form[[]Title] {
	return NewForm[[]Title](TitleSelectionFull)
}

func DefaultSeriesFormatListForm() Form[[]SeriesFormat] {
	return NewForm[[]SeriesFormat](SeriesFormatSelectionBasic)
}

func DefaultSeriesForm() Form[Series] {
	return NewForm[Series](SeriesSelectionWithMatchup)
}

func DefaultTeamForm() Form[Team] {
	return NewForm[Team](TeamSelectionFull)
}

func DefaultPlayerForm() Form[Player] {
	return NewForm[Player](PlayerSelectionFull)
}

func DefaultPlayerRoleForm() Form[PlayerRole] {
	return NewForm[PlayerRole](PlayerRoleSelectionFull)
}

func DefaultPlayerRolesForm() Form[[]PlayerRole] {
	return NewForm[[]PlayerRole](PlayerRoleSelectionFull)
}

func DefaultTournamentForm() Form[Tournament] {
	return NewForm[Tournament](TournamentSelectionFull)
}

func DefaultOrganizationForm() Form[Organization] {
	return NewForm[Organization](OrganizationSelectionFull)
}

func DefaultDataProvidersForm() Form[[]DataProvider] {
	return NewForm[[]DataProvider](DataProviderSelectionBasic)
}

func DefaultContentCatalogVersionForm() Form[ContentCatalogVersion] {
	return NewForm[ContentCatalogVersion](ContentCatalogVersionSelectionBasic.With(
		FieldSelect("title", TitleSelectionBasic),
	))
}

func DefaultContentCatalogEntityForm() Form[ContentCatalogEntity] {
	return NewForm[ContentCatalogEntity](ContentCatalogEntitySelectionBasic.With(
		FieldSelect("contentCatalogVersion", ContentCatalogVersionSelectionBasic),
	))
}
