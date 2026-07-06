// Package options holds form options that aren't stored in database
package options

var FormOpts = Options{
	Counties:  Counties,
	Fonts:     Fonts,
	Citations: Citations,
}

var Fonts = OptSet[Font]{
	title: "Font",
	name:  "font",
	opts: []Font{
		{Slug: "times", Name: "Times New Roman"},
		{Slug: "bookman", Name: "Bookman Old Style"},
	},
	def: "times",
}

var Counties = OptSet[County]{
	title: "County",
	name:  "county",
	opts: []County{
		{Slug: "sar", Name: "Sarasota", City: "Sarasota"},
		{Slug: "man", Name: "Manatee", City: "Bradenton"},
		{Slug: "des", Name: "DeSoto", City: "Arcadia"},
	},
	def: "sar",
}

var Citations = OptSet[Citation]{
	title: "Citation Style",
	name:  "citations",
	opts: []Citation{
		{Slug: "i", Name: "Italic", WordVal: "i"},
		{Slug: "u", Name: "Underline", WordVal: "u"},
	},
	def: "i",
}

type Options struct {
	Counties  OptSet[County]
	Fonts     OptSet[Font]
	Citations OptSet[Citation]
}

type Font struct {
	Slug string
	Name string
}

func (f Font) Key() string {
	return f.Slug
}

func (f Font) Label() string {
	return f.Name
}

type County struct {
	Slug string
	Name string
	City string
}

func (c County) Key() string {
	return c.Slug
}

func (c County) Label() string {
	return c.Name
}

type Citation struct {
	Slug    string
	Name    string
	WordVal string
}

func (c Citation) Key() string {
	return c.Slug
}

func (c Citation) Label() string {
	return c.Name
}

// Option exposes slugs (key) and labels for composing HTML forms from different
// option structs
type Option interface {
	Key() string
	Label() string
}

// OptSet groups all options for a given form field
type OptSet[T Option] struct {
	title string
	name  string
	opts  []T
	def   string
}

// Resolve checks that a given slug exists in the set
func (s OptSet[T]) Resolve(k string) (T, bool) {
	for _, opt := range s.opts {
		if opt.Key() == k {
			return opt, true
		}
	}

	var zero T
	return zero, false
}

func (s OptSet[T]) All() []T {
	return s.opts
}

// Default returns the default value for the set
func (s OptSet[T]) Default() T {
	opt, _ := s.Resolve(s.def)
	return opt
}

func (s OptSet[T]) IsDefault(k string) bool {
	return s.def == k
}

// Title returns the display title for the set
func (s OptSet[T]) Title() string {
	return s.title
}

func (s OptSet[T]) HTMLName() string {
	return s.name
}
