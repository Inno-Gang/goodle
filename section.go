package goodle

type Section interface {
	Resource
	Blocks() []*Block
}
