package goodle

import "github.com/inno-gang/goodle/richtext"

type Course interface {
	Id() int
	Title() string
	Description() *richtext.RichText
	MoodleUrl() string
}
