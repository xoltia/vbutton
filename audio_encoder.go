package vbutton

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

var (
	DefaultCodec  = "libopus"
	DefaultFormat = "ogg"
)

type readCloser struct {
	io.Reader
	close func() error
}

func (c *readCloser) Close() (err error) {
	if c.Reader == nil {
		return
	}
	err = c.close()
	c.Reader = nil
	return
}

type FFmpegEncoder struct {
	Format   string
	Codec    string
	BitRate  string
	buffPool *sync.Pool
}

func (e *FFmpegEncoder) getBuffer() *bytes.Buffer {
	if e.buffPool == nil {
		e.buffPool = &sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		}
	}

	return e.buffPool.Get().(*bytes.Buffer)
}

func (e *FFmpegEncoder) putBuffer(b *bytes.Buffer) {
	b.Reset()
	e.buffPool.Put(b)
}

func (e *FFmpegEncoder) Encode(r io.Reader) (out io.ReadCloser, err error) {
	if e.Format == "" {
		e.Format = DefaultFormat
	}

	if e.Codec == "" {
		e.Codec = DefaultCodec
	}

	if e.BitRate == "" {
		e.BitRate = "96k"
	}

	cmd := exec.Command(
		"ffmpeg",
		"-i", "pipe:0", // input
		"-f", e.Format, // output format
		"-c:a", e.Codec, // output codec
		"-b:a", e.BitRate, // output bitrate
		"pipe:1", // output
	)

	cmd.Stdin = r
	cmd.Stderr = new(bytes.Buffer)

	var stdout io.Reader

	if stdout, err = cmd.StdoutPipe(); err != nil {
		return nil, err
	}

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	buff := e.getBuffer()

	defer func() {
		if err != nil {
			e.putBuffer(buff)
		}
	}()

	if _, err = io.Copy(buff, stdout); err != nil {
		return
	}

	if err = cmd.Wait(); err != nil {
		if cmd.ProcessState.ExitCode() != 0 {
			err = fmt.Errorf("%w: %s", err, cmd.Stderr.(*bytes.Buffer).String())
		}
		return
	}

	out = &readCloser{
		Reader: buff,
		close: func() error {
			e.putBuffer(buff)
			return nil
		},
	}

	return
}

func (e *FFmpegEncoder) Extension() string {
	if strings.HasPrefix(e.Codec, "lib") {
		return e.Codec[3:] + "." + e.Format
	}

	return e.Codec + "." + e.Format
}

func NewOpusOggEncoder() *FFmpegEncoder {
	return &FFmpegEncoder{
		Format:  "ogg",
		Codec:   "libopus",
		BitRate: "128k",
	}
}

func NewOpusCafEncoder() *FFmpegEncoder {
	return &FFmpegEncoder{
		Format: "mp3",
		Codec:  "libmp3lame",
	}
}
