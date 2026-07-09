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
	Name       string
}

type IGroup struct {
	ID       int64
	MotionID int64
	Slug     string
	Name     string
}

type Issue struct {
	ID           int64
	GroupID      int64
	Name         string
	TemplatePath string
	SortOrder    int
}

// view models (for ui)

type CTMotions struct {
	CaseType CaseType
	Motions  []Motion
}

type MGroups struct {
	Motion Motion
	Groups []GIssues
}

type GIssues struct {
	Group  IGroup
	Issues []Issue
}
