package goodle

import "time"

type BlockType int8

func (bt BlockType) Name() string {
	switch bt {
	case BlockTypeLink:
		return "Link"
	case BlockTypeFile:
		return "File"
	case BlockTypeFolder:
		return "Folder"
	case BlockTypeAssignment:
		return "Assignment"
	case BlockTypeQuiz:
		return "Quiz"
	}
	return "Unknown"
}

const (
	BlockTypeUnknown    BlockType = 0
	BlockTypeLink       BlockType = 1
	BlockTypeFile       BlockType = 2
	BlockTypeFolder     BlockType = 3
	BlockTypeAssignment BlockType = 4
	BlockTypeQuiz       BlockType = 5
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
