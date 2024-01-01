package vbutton

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"time"
)

type VoiceClipRepository interface {
	InsertVoiceClip(vc *VoiceClip) error
	UpdateVoiceClip(vc *VoiceClip) error
	DeleteVoiceClip(id int64) error
	GetRecentVoiceClips(limit int) ([]*VoiceClip, error)
	GetClipTags(id int64) ([]string, error)
	GetVoiceClip(id int64) (*VoiceClip, error)
	GetVoiceClipsByVTuber(vtuberName string) ([]*VoiceClip, error)
	GetVoiceClipsByAgency(agencyName string) ([]*VoiceClip, error)
	GetVoiceClipsByTag(tag string) ([]*VoiceClip, error)
	GetUnapprovedVoiceClips(age time.Duration) ([]*VoiceClip, error)
	GetTopAgencies(limit int) ([]string, error)
	GetTopVTubers(limit int) ([]string, error)
	GetTopTags(limit int) ([]string, error)
	SearchClips(query, vtuber, agency, tag sql.NullString, limit int) ([]*VoiceClip, error)
}

type FileStorage interface {
	SaveFile(name string, content io.Reader) error
	GetFile(name string) (io.ReadCloser, error)
	DeleteFile(name string) error
}

type AudioEncoder interface {
	Encode(r io.Reader) (io.ReadCloser, error)
	Extension() string
}

type VoiceClipService struct {
	db           VoiceClipRepository
	audioStorage FileStorage
	encoders     []AudioEncoder
}

func NewVoiceClipService(db VoiceClipRepository, audioStorage FileStorage, encoders []AudioEncoder) *VoiceClipService {
	return &VoiceClipService{db: db, audioStorage: audioStorage, encoders: encoders}
}

func (s *VoiceClipService) FileTypes() []string {
	types := make([]string, len(s.encoders))

	for i, encoder := range s.encoders {
		types[i] = encoder.Extension()
	}

	return types
}

func (s *VoiceClipService) SearchClips(query, vtuber, agency, tag sql.NullString, limit int) ([]*VoiceClip, error) {
	return s.db.SearchClips(query, vtuber, agency, tag, limit)
}

func (s *VoiceClipService) CreateVoiceClip(clip *VoiceClip, inAudio io.Reader) error {
	err := s.db.InsertVoiceClip(clip)

	if err != nil {
		err = fmt.Errorf("failed to insert voice clip: %w", err)
		return err
	}

	if len(s.encoders) == 0 {
		err = fmt.Errorf("no encoders available")
		return err
	}

	if len(s.encoders) == 1 {
		audio, err := s.encoders[0].Encode(inAudio)

		if err != nil {
			err = fmt.Errorf("failed to encode audio: %w", err)
			return err
		}

		defer audio.Close()

		err = s.audioStorage.SaveFile(fmt.Sprintf("%d.%s", clip.ID, s.encoders[0].Extension()), audio)

		if err != nil {
			return err
		}

		return nil
	}

	buff := new(bytes.Buffer)
	tee := io.TeeReader(inAudio, buff)

	audio, err := s.encoders[0].Encode(tee)

	if err != nil {
		err = fmt.Errorf("failed to encode audio: %w", err)
		return err
	}

	err = s.audioStorage.SaveFile(fmt.Sprintf("%d.%s", clip.ID, s.encoders[0].Extension()), audio)

	if err != nil {
		audio.Close()
		return err
	}

	audio.Close()

	fmt.Println("len", buff.Len())

	reader := bytes.NewReader(buff.Bytes())

	for _, encoder := range s.encoders[1:] {
		reader.Seek(0, io.SeekStart)
		audio2, err := encoder.Encode(reader)

		if err != nil {
			err = fmt.Errorf("failed to encode audio: %w", err)
			return err
		}

		err = s.audioStorage.SaveFile(fmt.Sprintf("%d.%s", clip.ID, encoder.Extension()), audio2)

		if err != nil {
			audio2.Close()
			return err
		}

		audio2.Close()
	}

	return nil
}

func (s *VoiceClipService) GetVoiceClip(id int64) (vc *VoiceClip, err error) {
	return s.db.GetVoiceClip(id)
}

func (s *VoiceClipService) GetVoiceClipAudio(id int64) (audio io.ReadCloser, err error) {
	if len(s.encoders) == 0 {
		return nil, fmt.Errorf("no encoders available")
	}

	primaryType := s.encoders[0].Extension()
	audio, err = s.audioStorage.GetFile(fmt.Sprintf("%d.%s", id, primaryType))

	if err != nil {
		return nil, err
	}

	return
}

func (s *VoiceClipService) GetRecentVoiceClips(limit int) ([]*VoiceClip, error) {
	return s.db.GetRecentVoiceClips(limit)
}

func (s *VoiceClipService) GetClipTags(id int64) ([]string, error) {
	return s.db.GetClipTags(id)
}

func (s *VoiceClipService) GetVoiceClipsByVTuber(vtuberName string) ([]*VoiceClip, error) {
	return s.db.GetVoiceClipsByVTuber(vtuberName)
}

func (s *VoiceClipService) GetVoiceClipsByAgency(agencyName string) ([]*VoiceClip, error) {
	return s.db.GetVoiceClipsByAgency(agencyName)
}

func (s *VoiceClipService) GetVoiceClipsByTag(tag string) ([]*VoiceClip, error) {
	return s.db.GetVoiceClipsByTag(tag)
}

func (s *VoiceClipService) UpdateVoiceClip(clip *VoiceClip) error {
	return s.db.UpdateVoiceClip(clip)
}

func (s *VoiceClipService) DeleteVoiceClip(id int64) error {
	err := s.db.DeleteVoiceClip(id)

	if err != nil {
		return err
	}

	for _, encoder := range s.encoders {
		ext := encoder.Extension()
		err = s.audioStorage.DeleteFile(fmt.Sprintf("%d.%s", id, ext))

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *VoiceClipService) GetUnapprovedVoiceClips(age time.Duration) ([]*VoiceClip, error) {
	return s.db.GetUnapprovedVoiceClips(age)
}

func (s *VoiceClipService) GetTopAgencies(limit int) ([]string, error) {
	return s.db.GetTopAgencies(limit)
}

func (s *VoiceClipService) GetTopVTubers(limit int) ([]string, error) {
	return s.db.GetTopVTubers(limit)
}

func (s *VoiceClipService) GetTopTags(limit int) ([]string, error) {
	return s.db.GetTopTags(limit)
}
