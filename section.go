package goodle

import "github.com/inno-gang/goodle/richtext"

type Section interface {
	Id() int
	Title() string
	Description() *richtext.RichText
	Blocks() []Block
}
