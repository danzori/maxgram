package message

import (
	"cmp"
	"path"
	"strings"
)

type AttachmentKind string

const (
	KindPhoto     AttachmentKind = "photo"
	KindVideo     AttachmentKind = "video"
	KindVideoNote AttachmentKind = "video_note"
	KindVoice     AttachmentKind = "voice"
	KindAudio     AttachmentKind = "audio"
	KindFile      AttachmentKind = "file"
	KindSticker   AttachmentKind = "sticker"
	KindLocation  AttachmentKind = "location"
	KindContact   AttachmentKind = "contact"
	KindShare     AttachmentKind = "share"
	KindUnknown   AttachmentKind = "unknown"
)

type Location struct {
	Latitude  float64
	Longitude float64
}

type Contact struct {
	Name  string
	Phone string
	VCard string
}

type Attachment struct {
	Kind    AttachmentKind
	RawType string

	URL          string
	ThumbnailURL string
	FileID       int64
	Token        string

	FileName string
	Size     int64
	Duration int
	Width    int
	Height   int

	Location *Location
	Contact  *Contact
}

func (a Attachment) NeedsDownload() bool {
	return a.Kind != KindLocation && a.Kind != KindContact &&
		a.Kind != KindShare && a.Kind != KindUnknown
}

func (a Attachment) Resolvable() bool {
	return a.URL == "" && a.FileID != 0 && a.Kind == KindFile
}

func (a Attachment) DisplayName() string {
	if a.FileName != "" {
		return a.FileName
	}

	base, ext := a.defaultName()

	return base + cmp.Or(a.extension(), ext)
}

func (a Attachment) defaultName() (base, ext string) {
	switch a.Kind {
	case KindPhoto:
		return "photo", ".jpg"
	case KindVideo:
		return "video", ".mp4"
	case KindVideoNote:
		return "video_note", ".mp4"
	case KindVoice:
		return "voice", ".ogg"
	case KindAudio:
		return "audio", ".m4a"
	case KindSticker:
		return "sticker", ".webp"
	default:
		return "file", ""
	}
}

func (a Attachment) extension() string {
	return path.Ext(strings.SplitN(a.URL, "?", 2)[0])
}
