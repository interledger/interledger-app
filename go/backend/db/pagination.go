package db

import (
	"fmt"

	adminpb "github.com/interledger/interledger-app/go/proto/backend/admin/v1"
	pb "github.com/interledger/interledger-app/go/proto/backend/v1"
)

type Pagination struct {
	PageToken string
	PageSize  int
	Search    string
}

func (p *Pagination) SQL() string {
	if p.PageSize <= 0 || p.PageSize > 50 {
		p.PageSize = 50
	}

	return fmt.Sprintf(" LIMIT %d ", p.PageSize+1)
}

func PaginationFromPB(req *pb.PaginationRequest) Pagination {
	pageSize := req.GetPageSize()
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 50
	}

	return Pagination{
		PageToken: req.GetPageToken(),
		PageSize:  int(pageSize),
	}
}

func FromAdminPB(req *adminpb.PaginationRequest) Pagination {
	pageSize := req.GetPageSize()
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 50
	}

	return Pagination{
		PageToken: req.GetPageToken(),
		PageSize:  int(pageSize),
		Search:    req.GetSearch(),
	}
}
