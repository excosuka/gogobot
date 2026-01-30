package parserService

type ParserService interface {
	Parse(url string) (*ParsedPage, error)
}

type ParsedPage struct{}
