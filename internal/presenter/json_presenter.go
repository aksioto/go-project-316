package presenter

import (
	"encoding/json"

	"code/internal/domain"
)

type JSONPresenter struct {
	indent bool
}

func NewJSONPresenter(indent bool) *JSONPresenter {
	return &JSONPresenter{
		indent: indent,
	}
}

func (p *JSONPresenter) Present(report domain.Report) ([]byte, error) {
	dto := MapReport(report)

	if p.indent {
		return json.MarshalIndent(dto, "", "  ")
	}
	return json.Marshal(dto)
}
