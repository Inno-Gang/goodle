package goodle

import "time"

//go:generate enumer -type=BlockType -trimprefix=BlockType
type BlockType int8

const (
	BlockTypeUnknown BlockType = iota + 1
	BlockTypeLink
	BlockTypeFile
	BlockTypeFolder
	BlockTypeAssignment
	BlockTypeQuiz
)

type Block interface {
	Id() int
	Type() BlockType
	Title() string
	MoodleUrl() string
}

type BlockLink interface {
	Block
	Url() string
}

type BlockFile interface {
	Block
	DownloadUrl() string
	FileName() string
	SizeBytes() uint
	MimeType() string
	CreatedAt() time.Time
	LastModifiedAt() time.Time
}

type BlockFolder interface {
	Block
}

type BlockAssignment interface {
	Block
	SubmissionsAcceptedFrom() time.Time
	DeadlineAt() time.Time
	StrictDeadlineAt() time.Time
}

type BlockQuiz interface {
	Block
	OpensAt() time.Time
	ClosesAt() time.Time
}
