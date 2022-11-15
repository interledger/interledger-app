package db

import (
	"fmt"

	adminpb "gitlab.com/fynbos/proto/backend/admin/v1"
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

	if p.PageSize <= 0 || p.Page > 50 {
		p.PageSize = 50
	}

	return fmt.Sprintf(" LIMIT %d OFFSET %d  ", p.PageSize, pg*p.PageSize)
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

func FromAdminPB(req *adminpb.PaginationRequest) Pagination {
	pageSize := req.PageSize
	if pageSize > 50 {
		pageSize = 50
	}
	return Pagination{
		Page:     int(req.Page),
		PageSize: int(pageSize),
	}
}

func (p *Pagination) ToPB(resultLen int) *pb.PaginationResponse {
	return &pb.PaginationResponse{
		Page:        int32(p.Page),
		PageSize:    int32(p.PageSize),
		HasNextPage: resultLen == p.PageSize,
	}
}

func (p *Pagination) ToAdminPB(resultLen int) *adminpb.PaginationResponse {
	return &adminpb.PaginationResponse{
		Page:        int32(p.Page),
		PageSize:    int32(p.PageSize),
		HasNextPage: resultLen == p.PageSize,
	}
}
