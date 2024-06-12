package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Frame struct {
	ID       string
	Filepath string
	FrameNo  int
	VideoID  string
}

func CreateFrameTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS frames (
			id TEXT PRIMARY KEY,
			filepath TEXT,
			frame_no INTEGER,
			video_id TEXT
		);
	`)
	return err
}

func CreateFrameIndex(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_video_id_frame_no ON frames (video_id, frame_no);
	`)
	return err
}

func InsertFrame(db *sql.DB, f *Frame) error {
	_, err := db.Exec(`
		INSERT INTO frames (id, filepath, frame_no, video_id)
		VALUES (?, ?, ?, ?);
	`, f.ID, f.Filepath, f.FrameNo, f.VideoID)
	return err
}

func GetFrame(db *sql.DB, videoID string, frameNo int) (*Frame, error) {
	row := db.QueryRow(`
		SELECT id, filepath, frame_no, video_id
		FROM frames
		WHERE video_id = ? AND frame_no = ?; 
	`, videoID, frameNo)

	f := &Frame{}
	err := row.Scan(&f.ID, &f.Filepath, &f.FrameNo, &f.VideoID)
	if err != nil {
		return nil, err
	}
	return f, nil
}
