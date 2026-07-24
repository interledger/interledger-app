package db

import (
	"fmt"
	"strings"

	adminpb "github.com/interledger/interledger-app/go/proto/backend/admin/v1"
	pb "github.com/interledger/interledger-app/go/proto/backend/v1"
)

type Pagination struct {
	PageToken string
	PageSize  int
	Search    string
	Filter    WalletFilter
}

type WalletFilter struct {
	FirstName     string
	LastName      string
	WalletAddress string
	ProviderID    string
	WalletIDs     []string
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

// FromListWalletsPB converts a ListWalletsRequest the same way FromAdminPB
// converts a PaginationRequest (same page-size cap), plus the optional
// WalletSearchFilter. Whitespace-only filter values are treated as empty
// (empty fields must never restrict results).
func FromListWalletsPB(req *adminpb.ListWalletsRequest) Pagination {
	pageSize := req.GetPageSize()
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 50
	}

	p := Pagination{
		PageToken: req.GetPageToken(),
		PageSize:  int(pageSize),
		Search:    req.GetSearch(),
	}

	if f := req.GetFilter(); f != nil {
		p.Filter = WalletFilter{
			FirstName:     strings.TrimSpace(f.GetFirstName()),
			LastName:      strings.TrimSpace(f.GetLastName()),
			WalletAddress: strings.TrimSpace(f.GetWalletAddress()),
			ProviderID:    strings.TrimSpace(f.GetProviderId()),
		}
	}

	return p
}
