package api

import "seclink/db"

type SUiData struct {
	Links []db.GetAllLinksRow
	Posts []db.Post
}

type SFile struct {
	Path      string
	TtlString string
}
