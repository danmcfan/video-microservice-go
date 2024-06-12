package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Video struct {
	ID         string  `json:"id"`
	Filepath   string  `json:"-"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	FrameRate  float64 `json:"frame_rate"`
	FrameCount int     `json:"frame_count"`
}

func CreateVideoTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS videos (
			id TEXT PRIMARY KEY,
			filepath TEXT,
			width INTEGER,
			height INTEGER,
			frame_rate REAL,
			frame_count INTEGER
		);
	`)
	return err
}

func InsertVideo(d *sql.DB, v *Video) error {
	_, err := d.Exec(`
		INSERT INTO videos (id, filepath, width, height, frame_rate, frame_count)
		VALUES (?, ?, ?, ?, ?, ?);
	`, v.ID, v.Filepath, v.Width, v.Height, v.FrameRate, v.FrameCount)
	return err
}

func ListVideos(db *sql.DB) ([]*Video, error) {
	rows, err := db.Query(`
		SELECT id, filepath, width, height, frame_rate, frame_count
		FROM videos;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	videos := []*Video{}
	for rows.Next() {
		v := &Video{}
		err := rows.Scan(&v.ID, &v.Filepath, &v.Width, &v.Height, &v.FrameRate, &v.FrameCount)
		if err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}
	return videos, nil
}

func GetVideo(db *sql.DB, id string) (*Video, error) {
	row := db.QueryRow(`
		SELECT id, filepath, width, height, frame_rate, frame_count
		FROM videos
		WHERE id = ?;
	`, id)

	v := &Video{}
	err := row.Scan(&v.ID, &v.Filepath, &v.Width, &v.Height, &v.FrameRate, &v.FrameCount)
	if err != nil {
		return nil, err
	}
	return v, nil
}
