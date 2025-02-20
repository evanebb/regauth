package handlers

import (
	"net/http"
	"strconv"
)

type Pagination[D ~[]E, E any] struct {
	Data      D
	Total     int
	Page      int
	PrevPage  int
	NextPage  int
	FirstPage int
	LastPage  int
	PageSize  int
	Start     int
	End       int
	HasPrev   bool
	HasNext   bool
}

func PaginateRequest[D ~[]E, E any](r *http.Request, data D, defaultPageSize int) Pagination[D, E] {
	// This is gross, don't look here, I am not a front-end developer :D
	p := Pagination[D, E]{}
	p.Total = len(data)

	p.Page = getIntOrDefault(r.URL.Query().Get("page"), 1)
	p.PrevPage = p.Page - 1
	p.NextPage = p.Page + 1

	p.PageSize = getIntOrDefault(r.URL.Query().Get("pageSize"), defaultPageSize)

	p.FirstPage = 1
	lastPage := p.Total / p.PageSize
	if p.Total%p.PageSize > 0 {
		lastPage++
	}
	p.LastPage = lastPage

	p.Start = (p.Page - 1) * p.PageSize
	end := p.Start + p.PageSize
	if end > p.Total {
		end = p.Total
	}
	p.End = end

	p.HasPrev = p.Page != 1
	p.HasNext = p.Page < p.LastPage

	p.Data = data[p.Start:p.End]

	return p
}

func getIntOrDefault(val string, defaultVal int) int {
	if val == "" {
		return defaultVal
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}

	return intVal
}
