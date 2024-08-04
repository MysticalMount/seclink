package db

import "time"

type SSharedLink struct {
	Id        string
	Path      string
	ExpiresAt time.Time
	Ttl       time.Duration
	TtlString string
	Url       string
}
