package vbutton

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type VoiceClipDB struct {
	*sql.DB
}

func NewVoiceClipDB(db *sql.DB) *VoiceClipDB {
	return &VoiceClipDB{db}
}

func (db *VoiceClipDB) Close() error {
	return db.DB.Close()
}

func (db *VoiceClipDB) Create() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS voice_clips (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			vtuber_name TEXT NOT NULL,
			agency_name TEXT,
			reference_url TEXT,
			length INTEGER NOT NULL,
			approved_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)

	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			text TEXT NOT NULL,
			UNIQUE (text)
		);
	`)

	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS voice_clip_tags (
			voice_clip_id INTEGER NOT NULL,
			tag_id INTEGER NOT NULL,
			PRIMARY KEY (voice_clip_id, tag_id)
		);
	`)

	return err
}

func (db *VoiceClipDB) InsertVoiceClip(clip *VoiceClip) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO voice_clips (
			title,
			vtuber_name,
			agency_name,
			reference_url,
			length
		) VALUES (?, ?, ?, ?, ?);
	`)
	if err != nil {
		return err
	}

	res, err := stmt.Exec(
		clip.Title,
		clip.VTuberName,
		clip.AgencyName,
		clip.ReferenceURL,
		clip.Length,
	)
	if err != nil {
		return err
	}

	clip.ID, err = res.LastInsertId()
	if err != nil {
		return err
	}

	for _, tag := range clip.Tags {
		tagID, err := db.insertTag(tx, tag)
		if err != nil {
			return err
		}

		err = db.insertVoiceClipTag(tx, clip.ID, tagID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *VoiceClipDB) GetVoiceClipsByVTuber(vtuberName string) ([]*VoiceClip, error) {
	rows, err := db.Query(`
		SELECT
			voice_clips.id,
			voice_clips.title,
			voice_clips.vtuber_name,
			voice_clips.agency_name,
			voice_clips.reference_url,
			voice_clips.length,
			voice_clips.approved_at,
			voice_clips.created_at
		FROM voice_clips
		WHERE voice_clips.vtuber_name = ?
		AND voice_clips.approved_at IS NOT NULL
		ORDER BY voice_clips.created_at DESC;
	`, vtuberName)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var clips []*VoiceClip

	for rows.Next() {
		var clip VoiceClip

		err := rows.Scan(
			&clip.ID,
			&clip.Title,
			&clip.VTuberName,
			&clip.AgencyName,
			&clip.ReferenceURL,
			&clip.Length,
			&clip.ApprovedAt,
			&clip.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		clip.Tags, err = db.GetClipTags(clip.ID)

		if err != nil {
			return nil, err
		}

		clips = append(clips, &clip)
	}

	return clips, nil
}

func (db *VoiceClipDB) GetVoiceClipsByAgency(agency string) ([]*VoiceClip, error) {
	rows, err := db.Query(`
		SELECT
			voice_clips.id,
			voice_clips.title,
			voice_clips.vtuber_name,
			voice_clips.agency_name,
			voice_clips.reference_url,
			voice_clips.length,
			voice_clips.approved_at,
			voice_clips.created_at
		FROM voice_clips
		WHERE voice_clips.agency = ?
		AND voice_clips.approved_at IS NOT NULL
		ORDER BY voice_clips.created_at DESC;
	`, agency)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var clips []*VoiceClip

	for rows.Next() {
		var clip VoiceClip

		err := rows.Scan(
			&clip.ID,
			&clip.Title,
			&clip.VTuberName,
			&clip.AgencyName,
			&clip.ReferenceURL,
			&clip.Length,
			&clip.ApprovedAt,
			&clip.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		clip.Tags, err = db.GetClipTags(clip.ID)

		if err != nil {
			return nil, err
		}

		clips = append(clips, &clip)
	}

	return clips, nil
}

func (db *VoiceClipDB) GetVoiceClipsByTag(tag string) ([]*VoiceClip, error) {
	rows, err := db.Query(`
		SELECT
			voice_clips.id,
			voice_clips.title,
			voice_clips.vtuber_name,
			voice_clips.agency_name,
			voice_clips.reference_url,
			voice_clips.length,
			voice_clips.approved_at,
			voice_clips.created_at
		FROM voice_clips
		INNER JOIN voice_clip_tags ON voice_clip_tags.voice_clip_id = voice_clips.id
		INNER JOIN tags ON tags.id = voice_clip_tags.tag_id
		WHERE tags.text = ?
		AND voice_clips.approved_at IS NOT NULL
		ORDER BY voice_clips.created_at DESC;
	`, tag)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var clips []*VoiceClip

	for rows.Next() {
		var clip VoiceClip

		err := rows.Scan(
			&clip.ID,
			&clip.Title,
			&clip.VTuberName,
			&clip.AgencyName,
			&clip.ReferenceURL,
			&clip.Length,
			&clip.ApprovedAt,
			&clip.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		clip.Tags, err = db.GetClipTags(clip.ID)

		if err != nil {
			return nil, err
		}

		clips = append(clips, &clip)
	}

	return clips, nil
}

func (db *VoiceClipDB) GetRecentVoiceClips(limit int) ([]*VoiceClip, error) {
	rows, err := db.Query(`
		SELECT
			voice_clips.id,
			voice_clips.title,
			voice_clips.vtuber_name,
			voice_clips.agency_name,
			voice_clips.reference_url,
			voice_clips.length,
			voice_clips.approved_at,
			voice_clips.created_at
		FROM voice_clips
		WHERE voice_clips.approved_at IS NOT NULL
		ORDER BY voice_clips.created_at DESC
		LIMIT ?;
	`, limit)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var clips []*VoiceClip

	for rows.Next() {
		var clip VoiceClip

		err := rows.Scan(
			&clip.ID,
			&clip.Title,
			&clip.VTuberName,
			&clip.AgencyName,
			&clip.ReferenceURL,
			&clip.Length,
			&clip.ApprovedAt,
			&clip.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		clip.Tags, err = db.GetClipTags(clip.ID)

		if err != nil {
			return nil, err
		}

		clips = append(clips, &clip)
	}

	return clips, nil
}

func (db *VoiceClipDB) GetClipTags(clipID int64) ([]string, error) {
	rows, err := db.Query(`
		SELECT tags.text
		FROM tags
		INNER JOIN voice_clip_tags ON voice_clip_tags.tag_id = tags.id
		WHERE voice_clip_tags.voice_clip_id = ?;
	`, clipID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tags []string

	for rows.Next() {
		var tag string
		err := rows.Scan(&tag)
		if err != nil {
			return nil, err
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

func (db *VoiceClipDB) DeleteVoiceClip(id int64) error {
	tx, err := db.Begin()

	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		DELETE FROM voice_clips WHERE id = ?;
	`, id)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *VoiceClipDB) SearchClips(
	query sql.NullString,
	vtuber sql.NullString,
	agency sql.NullString,
	tag sql.NullString,
	limit int,
) ([]*VoiceClip, error) {
	var rows *sql.Rows

	rows, err := db.Query(`
		SELECT
			id,
			title,
			vtuber_name,
			agency_name,
			reference_url,
			length,
			approved_at,
			created_at
		FROM voice_clips
		WHERE approved_at IS NOT NULL
		AND id IN (
			SELECT DISTINCT voice_clips.id
			FROM voice_clips
			INNER JOIN voice_clip_tags ON voice_clip_tags.voice_clip_id = voice_clips.id
			INNER JOIN tags ON tags.id = voice_clip_tags.tag_id
			WHERE (
				voice_clips.title LIKE ?
				OR ? IS NULL
			)
			AND (
				voice_clips.vtuber_name = ?
				OR ? IS NULL
			)
			AND (
				voice_clips.agency_name = ?
				OR ? IS NULL
			)
			AND (
				tags.text = ?
				OR ? IS NULL
			)
			ORDER BY voice_clips.created_at DESC
			LIMIT ?
		)
	`, query, query, vtuber, vtuber, agency, agency, tag, tag, limit)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var clips []*VoiceClip

	for rows.Next() {
		var clip VoiceClip

		err := rows.Scan(
			&clip.ID,
			&clip.Title,
			&clip.VTuberName,
			&clip.AgencyName,
			&clip.ReferenceURL,
			&clip.Length,
			&clip.ApprovedAt,
			&clip.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		clip.Tags, err = db.GetClipTags(clip.ID)

		if err != nil {
			return nil, err
		}

		clips = append(clips, &clip)
	}

	return clips, nil
}

func (db *VoiceClipDB) GetTopTags(limit int) ([]string, error) {
	rows, err := db.Query(`
		SELECT tags.text
		FROM tags
		INNER JOIN voice_clip_tags ON voice_clip_tags.tag_id = tags.id
		INNER JOIN voice_clips ON voice_clips.id = voice_clip_tags.voice_clip_id
		WHERE voice_clips.approved_at IS NOT NULL
		GROUP BY tags.text
		ORDER BY COUNT(*) DESC
		LIMIT ?;
	`, limit)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tags []string

	for rows.Next() {
		var tag string
		err := rows.Scan(&tag)
		if err != nil {
			return nil, err
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

func (db *VoiceClipDB) GetTopVTubers(limit int) ([]string, error) {
	rows, err := db.Query(`
		SELECT vtuber_name
		FROM voice_clips
		WHERE approved_at IS NOT NULL
		GROUP BY vtuber_name
		ORDER BY COUNT(*) DESC
		LIMIT ?;
	`, limit)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var vtubers []string

	for rows.Next() {
		var vtuber string
		err := rows.Scan(&vtuber)
		if err != nil {
			return nil, err
		}

		vtubers = append(vtubers, vtuber)
	}

	return vtubers, nil
}

func (db *VoiceClipDB) GetTopAgencies(limit int) ([]string, error) {
	rows, err := db.Query(`
		SELECT agency_name
		FROM voice_clips
		WHERE agency_name IS NOT NULL
		AND approved_at IS NOT NULL
		GROUP BY agency_name
		ORDER BY COUNT(*) DESC
		LIMIT ?;
	`, limit)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var agencies []string

	for rows.Next() {
		var agency string
		err := rows.Scan(&agency)
		if err != nil {
			return nil, err
		}

		agencies = append(agencies, agency)
	}

	return agencies, nil
}

func (db *VoiceClipDB) GetUnapprovedVoiceClips(maxAge time.Duration) ([]*VoiceClip, error) {
	rows, err := db.Query(`
		SELECT
			voice_clips.id,
			voice_clips.title,
			voice_clips.vtuber_name,
			voice_clips.agency_name,
			voice_clips.reference_url,
			voice_clips.length,
			voice_clips.approved_at,
			voice_clips.created_at
		FROM voice_clips
		WHERE voice_clips.approved_at IS NULL
		AND voice_clips.created_at > ?
		ORDER BY voice_clips.created_at DESC;
	`, time.Now().Add(-maxAge))

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var clips []*VoiceClip

	for rows.Next() {
		var clip VoiceClip

		err := rows.Scan(
			&clip.ID,
			&clip.Title,
			&clip.VTuberName,
			&clip.AgencyName,
			&clip.ReferenceURL,
			&clip.Length,
			&clip.ApprovedAt,
			&clip.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		clip.Tags, err = db.GetClipTags(clip.ID)

		if err != nil {
			return nil, err
		}

		clips = append(clips, &clip)
	}

	return clips, nil
}

func (db *VoiceClipDB) UpdateVoiceClip(clip *VoiceClip) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		UPDATE voice_clips
		SET
			title = ?,
			vtuber_name = ?,
			agency_name = ?,
			reference_url = ?,
			length = ?,
			approved_at = ?
		WHERE id = ?;
	`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(
		clip.Title,
		clip.VTuberName,
		clip.AgencyName,
		clip.ReferenceURL,
		clip.Length,
		clip.ApprovedAt,
		clip.ID,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		DELETE FROM voice_clip_tags WHERE voice_clip_id = ?;
	`, clip.ID)
	if err != nil {
		return err
	}

	for _, tag := range clip.Tags {
		tagID, err := db.insertTag(tx, tag)
		if err != nil {
			return err
		}

		err = db.insertVoiceClipTag(tx, clip.ID, tagID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *VoiceClipDB) GetVoiceClip(id int64) (*VoiceClip, error) {
	row := db.QueryRow(`
		SELECT
			voice_clips.id,
			voice_clips.title,
			voice_clips.vtuber_name,
			voice_clips.agency_name,
			voice_clips.reference_url,
			voice_clips.length,
			voice_clips.approved_at,
			voice_clips.created_at
		FROM voice_clips
		WHERE voice_clips.id = ?;
	`, id)

	var clip VoiceClip

	err := row.Scan(
		&clip.ID,
		&clip.Title,
		&clip.VTuberName,
		&clip.AgencyName,
		&clip.ReferenceURL,
		&clip.Length,
		&clip.ApprovedAt,
		&clip.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	clip.Tags, err = db.GetClipTags(clip.ID)

	if err != nil {
		return nil, err
	}

	return &clip, nil
}

func (db *VoiceClipDB) insertTag(tx *sql.Tx, tag string) (int64, error) {
	row := tx.QueryRow(`
		INSERT INTO tags (text) VALUES (?) ON CONFLICT (text) DO UPDATE SET text = excluded.text RETURNING id;
	`, tag)

	var tagID int64
	err := row.Scan(&tagID)
	return tagID, err
}

func (db *VoiceClipDB) insertVoiceClipTag(tx *sql.Tx, clipID, tagID int64) error {
	stmt, err := tx.Prepare(`
		INSERT INTO voice_clip_tags (voice_clip_id, tag_id) VALUES (?, ?);
	`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(clipID, tagID)
	return err
}
