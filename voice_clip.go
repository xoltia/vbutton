package vbutton

import (
	"database/sql"
	"time"
)

type VoiceClip struct {
	ID           int64
	Title        string
	VTuberName   string
	AgencyName   sql.NullString
	Tags         []string
	ReferenceURL sql.NullString
	Length       time.Duration
	ApprovedAt   sql.NullTime
	CreatedAt    time.Time
}
