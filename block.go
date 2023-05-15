package goodle

import "time"

type BlockType int8

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
	SubmissionsAcceptedFrom() *time.Time
	DeadlineAt() *time.Time
	StrictDeadlineAt() *time.Time
}

type BlockQuiz interface {
	Block
	OpensAt() *time.Time
	ClosesAt() *time.Time
	Duration() *time.Duration
}
