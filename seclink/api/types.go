package api

type SUiData struct {
	SharedLinks []SSharedLink
	Files       []SFile
}

type SSharedLink struct {
	Path      string
	Url       string
	TtlString string
}

type SFile struct {
	Path      string
	TtlString string
}
