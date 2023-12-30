package vbutton

import (
	"fmt"
	"io"
)

type VoiceClipRepository interface {
	InsertVoiceClip(vc *VoiceClip) error
	GetRecentVoiceClips(limit int) ([]*VoiceClip, error)
	GetClipTags(id int64) ([]string, error)
	GetVoiceClip(id int64) (*VoiceClip, error)
	GetVoiceClipsByVTuber(vtuberName string) ([]*VoiceClip, error)
	GetVoiceClipsByAgency(agencyName string) ([]*VoiceClip, error)
	GetVoiceClipsByTag(tag string) ([]*VoiceClip, error)
}

type FileStorage interface {
	SaveFile(name string, content io.Reader) error
	GetFile(name string) (io.Reader, error)
}

type AudioEncoder interface {
	Encode(r io.Reader) (io.Reader, error)
	Extension() string
}

type VoiceClipService struct {
	db           VoiceClipRepository
	audioStorage FileStorage
	audioEncoder AudioEncoder
}

func NewVoiceClipService(db VoiceClipRepository, audioStorage FileStorage, encoder AudioEncoder) *VoiceClipService {
	return &VoiceClipService{db: db, audioStorage: audioStorage, audioEncoder: encoder}
}

func (s *VoiceClipService) CreateVoiceClip(clip *VoiceClip, audio io.Reader) error {
	err := s.db.InsertVoiceClip(clip)

	if err != nil {
		err = fmt.Errorf("failed to insert voice clip: %w", err)
		return err
	}

	audio, err = s.audioEncoder.Encode(audio)

	if err != nil {
		err = fmt.Errorf("failed to encode audio: %w", err)
		return err
	}

	return s.audioStorage.SaveFile(fmt.Sprintf("%d.%s", clip.ID, s.audioEncoder.Extension()), audio)
}

func (s *VoiceClipService) GetVoiceClip(id int64) (vc *VoiceClip, err error) {
	return s.db.GetVoiceClip(id)
}

func (s *VoiceClipService) GetVoiceClipAudio(id int64) (audio io.Reader, err error) {
	return s.audioStorage.GetFile(fmt.Sprintf("%d.%s", id, s.audioEncoder.Extension()))
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
