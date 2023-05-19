package goodle

import "github.com/inno-gang/goodle/richtext"

type Resource interface {
	Id() int
	Title() string
	Description() *richtext.RichText
	MoodleUrl() string
}

var unused_var = 123
