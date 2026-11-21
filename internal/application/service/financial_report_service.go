package service

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/Doratou/Backend/internal/application/adapter"
	"github.com/Doratou/Backend/internal/application/dto"
	"github.com/Doratou/Backend/internal/domain/entity"
	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/exception"
	"github.com/Doratou/Backend/internal/domain/finance"
	domainDB "github.com/Doratou/Backend/internal/domain/repository/DB"
	"github.com/Doratou/Backend/internal/infrastructure/database"
)

const settlementOptimizationTimeout = 2 * time.Second

type financialReportService struct {
	repo    domainDB.FinancialReportRepository
	users   domainDB.UserRepository
	readTrx database.TrxManager
	urls    adapter.AttachmentURLResolver
}

func NewFinancialReportService(repo domainDB.FinancialReportRepository, users domainDB.UserRepository, readTrx database.TrxManager, urls adapter.AttachmentURLResolver) adapter.FinancialReportService {
	return &financialReportService{repo: repo, users: users, readTrx: readTrx, urls: urls}
}

func (s *financialReportService) readGroup(ctx context.Context, userID, groupID int64, read func(context.Context, dto.FinancialReportContext) error) error {
	if userID <= 0 || groupID <= 0 {
		return exception.ErrBadRequest
	}
	return s.readTrx.WithTransaction(ctx, func(txCtx context.Context) error {
		user, err := s.users.FindByID(txCtx, userID)
		if err != nil {
			return err
		}
		if user.Status != enum.UserStatusActive {
			return exception.ErrUserInactive
		}
		group, err := s.repo.FindGroup(txCtx, groupID)
		if err != nil {
			return err
		}
		viewer, err := s.repo.FindViewerMemberID(txCtx, groupID, userID)
		if err != nil {
			return err
		}
		return read(txCtx, dto.FinancialReportContext{GroupID: groupID, GroupStatus: group.Status, ViewerMemberID: viewer, Currency: "TOMAN"})
	})
}

func (s *financialReportService) ExpenseShares(ctx context.Context, userID, groupID int64, owedByMe bool, pagination dto.ExpensePaginationRequest) (*dto.ExpenseShareReport, error) {
	page, limit, offset, err := expensePagination(pagination.Page, pagination.Limit)
	if err != nil {
		return nil, err
	}
	report, edges, members, total, err := s.expenseReport(ctx, userID, groupID, &owedByMe, limit, offset)
	if err != nil {
		return nil, err
	}
	items := make([]dto.ExpenseReportItem, 0, len(edges))
	for _, edge := range edges {
		party := edge.FromMemberID
		if owedByMe {
			party = edge.ToMemberID
		}
		items = append(items, dto.ExpenseReportItem{ExpenseReportEdge: edge, Counterparty: members[party]})
	}
	return &dto.ExpenseShareReport{FinancialReportContext: report, Items: items,
		Pagination: dto.PaginationResponse{Page: page, Limit: limit, Total: total}}, nil
}

func (s *financialReportService) ExpenseGraph(ctx context.Context, userID, groupID int64) (*dto.ExpenseGraphReport, error) {
	report, edges, members, _, err := s.expenseReport(ctx, userID, groupID, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	nodes := make([]dto.FinancialReportMember, 0, len(members))
	for _, member := range members {
		nodes = append(nodes, member)
	}
	slices.SortFunc(nodes, func(a, b dto.FinancialReportMember) int {
		switch {
		case a.MemberID < b.MemberID:
			return -1
		case a.MemberID > b.MemberID:
			return 1
		default:
			return 0
		}
	})
	return &dto.ExpenseGraphReport{FinancialReportContext: report, Nodes: nodes, Edges: edges}, nil
}

func (s *financialReportService) expenseReport(ctx context.Context, userID, groupID int64, owedByMe *bool, limit, offset int) (dto.FinancialReportContext, []dto.ExpenseReportEdge, map[int64]dto.FinancialReportMember, int64, error) {
	var report dto.FinancialReportContext
	var shares []domainDB.ExpenseReportShare
	var members []entity.GroupMember
	var total int64
	err := s.readGroup(ctx, userID, groupID, func(txCtx context.Context, scope dto.FinancialReportContext) error {
		report = scope
		var from, to int64
		if owedByMe != nil {
			if *owedByMe {
				from = scope.ViewerMemberID
			} else {
				to = scope.ViewerMemberID
			}
		}
		var err error
		shares, total, err = s.repo.ListExpenseShares(txCtx, groupID, from, to, limit, offset)
		if err != nil {
			return err
		}
		ids := make(map[int64]bool)
		for _, share := range shares {
			ids[share.FromMemberID], ids[share.ToMemberID] = true, true
		}
		memberIDs := make([]int64, 0, len(ids))
		for id := range ids {
			memberIDs = append(memberIDs, id)
		}
		members, err = s.repo.ListMembers(txCtx, groupID, memberIDs)
		return err
	})
	if err != nil {
		return report, nil, nil, 0, err
	}

	profiles, err := s.memberProfiles(ctx, userID, members)
	if err != nil {
		return report, nil, nil, 0, err
	}
	edges := make([]dto.ExpenseReportEdge, 0, len(shares))
	for _, share := range shares {
		edges = append(edges, dto.ExpenseReportEdge{Transfer: share.Transfer, ShareID: share.ShareID,
			ExpenseID: share.ExpenseID, ExpenseTitle: share.ExpenseTitle, OccurredAt: share.OccurredAt})
	}
	return report, edges, profiles, total, nil
}

func (s *financialReportService) SettlementGraph(ctx context.Context, userID, groupID int64) (*dto.SettlementGraphReport, error) {
	scope, positions, edges, members, err := s.settlementPlan(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}
	profiles, err := s.memberProfiles(ctx, userID, members)
	if err != nil {
		return nil, err
	}
	nodes := make([]dto.SettlementGraphNode, 0, len(positions))
	for _, position := range positions {
		nodes = append(nodes, dto.SettlementGraphNode{FinancialReportMember: profiles[position.MemberID], NetAmountToman: position.NetToman})
	}
	return &dto.SettlementGraphReport{SettlementReportContext: scope, Nodes: nodes, Edges: edges}, nil
}

func (s *financialReportService) MySettlements(ctx context.Context, userID, groupID int64) (*dto.MySettlementReport, error) {
	scope, positions, edges, members, err := s.settlementPlan(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}
	viewer := scope.ViewerMemberID
	report := &dto.MySettlementReport{SettlementReportContext: scope, Outgoing: make([]dto.SettlementReportItem, 0), Incoming: make([]dto.SettlementReportItem, 0)}
	counterparties := make(map[int64]bool)
	for _, edge := range edges {
		if edge.FromMemberID == viewer {
			counterparties[edge.ToMemberID] = true
		} else if edge.ToMemberID == viewer {
			counterparties[edge.FromMemberID] = true
		}
	}
	members = slices.DeleteFunc(members, func(member entity.GroupMember) bool { return !counterparties[member.ID] })
	profiles, err := s.memberProfiles(ctx, userID, members)
	if err != nil {
		return nil, err
	}
	for _, edge := range edges {
		if edge.FromMemberID == viewer {
			report.Outgoing = append(report.Outgoing, dto.SettlementReportItem{Transfer: edge, Counterparty: profiles[edge.ToMemberID]})
			report.TotalToPayToman += edge.AmountToman
		} else if edge.ToMemberID == viewer {
			report.Incoming = append(report.Incoming, dto.SettlementReportItem{Transfer: edge, Counterparty: profiles[edge.FromMemberID]})
			report.TotalToReceiveToman += edge.AmountToman
		}
	}
	for _, position := range positions {
		if position.MemberID == viewer {
			report.NetAmountToman = position.NetToman
			break
		}
	}
	return report, nil
}

func (s *financialReportService) settlementPlan(ctx context.Context, userID, groupID int64) (dto.SettlementReportContext, []finance.Position, []finance.Transfer, []entity.GroupMember, error) {
	var report dto.SettlementReportContext
	var ledger *finance.Ledger
	var members []entity.GroupMember
	var positions []finance.Position
	err := s.readGroup(ctx, userID, groupID, func(txCtx context.Context, scope dto.FinancialReportContext) error {
		var err error
		ledger, err = loadFinancialLedger(txCtx, s.repo, groupID)
		if err != nil {
			return err
		}
		positions, err = ledger.Positions()
		if err != nil {
			return financialCalculationError(err)
		}
		ids := make([]int64, 0, len(positions))
		for _, position := range positions {
			ids = append(ids, position.MemberID)
		}
		members, err = s.repo.ListMembers(txCtx, groupID, ids)
		report = dto.SettlementReportContext{FinancialReportContext: scope, BalanceRevision: ledger.Revision(),
			CanRegisterPayments: scope.GroupStatus != enum.GroupStatusClosed}
		return err
	})
	if err != nil {
		return report, nil, nil, nil, err
	}
	solverCtx, cancel := context.WithTimeout(ctx, settlementOptimizationTimeout)
	defer cancel()
	edges, err := finance.MinimumTransfers(solverCtx, positions)
	if err != nil {
		return report, nil, nil, nil, financialCalculationError(err)
	}
	return report, positions, edges, members, nil
}

func (s *financialReportService) memberProfiles(ctx context.Context, userID int64, members []entity.GroupMember) (map[int64]dto.FinancialReportMember, error) {
	ids := make([]int64, 0, len(members))
	for _, member := range members {
		if member.User != nil && member.User.AvatarAttachmentID != nil {
			ids = append(ids, *member.User.AvatarAttachmentID)
		}
	}
	var urls map[int64]string
	if len(ids) > 0 {
		var err error
		urls, err = s.urls.ResolveAccessURLs(ctx, userID, ids)
		var appErr *exception.AppError
		if err != nil && (!errors.As(err, &appErr) || appErr.Code != exception.ErrStorageUnavailable.Code) {
			return nil, err
		}
	}
	profiles := make(map[int64]dto.FinancialReportMember, len(members))
	for _, member := range members {
		profile := dto.FinancialReportMember{MemberID: member.ID, UserID: member.UserID, IsFormerMember: !activeSettlementMember(&member)}
		if member.User != nil {
			profile.Name = member.User.Name
			if member.User.AvatarAttachmentID != nil {
				if url := urls[*member.User.AvatarAttachmentID]; url != "" {
					profile.AvatarURL = &url
				}
			}
		}
		profiles[member.ID] = profile
	}
	return profiles, nil
}

func loadFinancialLedger(ctx context.Context, repo domainDB.FinancialBalanceRepository, groupID int64) (*finance.Ledger, error) {
	data, err := repo.LoadBalanceData(ctx, groupID)
	if err != nil {
		return nil, err
	}
	ledger, err := finance.NewLedger(data)
	if err != nil {
		return nil, financialCalculationError(err)
	}
	return ledger, nil
}

func financialCalculationError(err error) error {
	switch {
	case errors.Is(err, finance.ErrAmountOutOfRange):
		return exception.ErrFinancialAmountOutOfRange
	case errors.Is(err, context.DeadlineExceeded):
		return exception.ErrSettlementPlanTimeout
	default:
		return exception.Wrap(exception.ErrInternal, err)
	}
}
