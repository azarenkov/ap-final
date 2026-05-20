package domain

import "time"

type SearchFilter struct {
	Origin          string
	Destination     string
	DepartureAfter  *time.Time
	DepartureBefore *time.Time
	Page            int32
	PageSize        int32
}

func (s *SearchFilter) Normalize() {
	if s.Page < 1 {
		s.Page = 1
	}
	if s.PageSize < 1 || s.PageSize > 100 {
		s.PageSize = 20
	}
}

func (s *SearchFilter) Offset() int32 {
	return (s.Page - 1) * s.PageSize
}
