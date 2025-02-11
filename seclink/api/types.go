package api

import (
	"html/template"
	"seclink/db"
)

type UiData struct {
	Links []db.GetAllLinksRow
	Posts []db.Post
}

type SFile struct {
	Path      string
	TtlString string
}

type Page struct {
	Title                   string
	Slug                    string
	Parent                  string
	Content                 template.HTML
	Description             string
	Order                   int
	Headers                 []string // these are the in page h2 tags
	MetaDescription         string
	MetaPropertyTitle       string
	MetaPropertyDescription string
}
