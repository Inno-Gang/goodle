package richtext

type RichText struct {
	text     string
	entities []Entity
}

type Entity struct {
	start uint
	end   uint
}

type EntityBold struct{}
type EntityItalic struct{}

func ParseHtml(rawHtml string) (*RichText, error) {
	//TODO implement HTML parsing
	return &RichText{
		text:     rawHtml,
		entities: make([]Entity, 0),
	}, nil
}
