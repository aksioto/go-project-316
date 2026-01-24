package presenter

import "time"

type ReportDTO struct {
	RootURL     string    `json:"root_url"`
	Depth       int       `json:"depth"`
	GeneratedAt time.Time `json:"generated_at"`
	Pages       []PageDTO `json:"pages"`
}
