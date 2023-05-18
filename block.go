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
	return "Unsupported"
}

const (
	BlockTypeUnsupported BlockType = 0
	BlockTypeLink        BlockType = 1
	BlockTypeFile        BlockType = 2
	BlockTypeFolder      BlockType = 3
	BlockTypeAssignment  BlockType = 4
	BlockTypeQuiz        BlockType = 5
)

type Block interface {
	Resource
	Type() BlockType
}

type BlockLink interface {
	Url() string
}

type BlockFile interface {
	Block
	DownloadUrl() string
	SizeBytes() uint64
	CreatedAt() time.Time
	LastModifiedAt() time.Time
}

type BlockFolder interface {
	Block
	DownloadZipUrl() string
}

type BlockAssignment interface {
	Block
	SubmissionsAcceptedFrom() (time.Time, bool)
	DeadlineAt() (time.Time, bool)
	StrictDeadlineAt() (time.Time, bool)
}

type BlockQuiz interface {
	Block
	OpensAt() (time.Time, bool)
	ClosesAt() (time.Time, bool)
	Duration() (time.Duration, bool)
}
