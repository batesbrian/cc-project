package store

// domain models (maps to db rows)

type CaseType struct {
	ID   int64
	Slug string
	Name string
}

type Motion struct {
	ID         int64
	CaseTypeID int64
	Slug       string
	Name       string
}

type Issue struct {
	ID           int64
	Slug         string
	Name         string
	TemplatePath string
	SortOrder    int
}

// view models (for ui)

type CaseTypeWithMotions struct {
	CaseType CaseType
	Motions  []Motion
}

type MotionWithIssues struct {
	Motion Motion
	Issues []Issue
}
