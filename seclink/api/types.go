package api

import "seclink/db"

type SUiData struct {
	SharedLinks []db.Link
	Files       []SFile
}

type SFile struct {
	Path      string
	TtlString string
}
