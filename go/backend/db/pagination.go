package db

import (
	"fmt"

	pb "gitlab.com/fynbos/proto/backend/v1"
)

type Pagination struct {
	Page     int
	PageSize int
}

func (p *Pagination) SQL() string {
	pg := p.Page
	if pg > 0 {
		pg -= 1
	}

	return fmt.Sprintf(" LIMIT %d OFFSET %d ORDER BY created_at ASC ", p.PageSize, pg*p.PageSize)
}

func PaginationFromPB(req *pb.PaginationRequest) Pagination {
	pageSize := req.PageSize
	if pageSize > 50 {
		pageSize = 50
	}
	return Pagination{
		Page:     int(req.Page),
		PageSize: int(pageSize),
	}
}

func (p *Pagination) PaginationToPB(resultLen int) *pb.PaginationResponse {
	return &pb.PaginationResponse{
		Page:        int32(p.Page),
		PageSize:    int32(p.PageSize),
		HasNextPage: resultLen == p.PageSize,
	}
}
