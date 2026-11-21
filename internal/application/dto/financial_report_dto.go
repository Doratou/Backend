package dto

import (
	"time"

	"github.com/Doratou/Backend/internal/domain/enum"
	"github.com/Doratou/Backend/internal/domain/finance"
)

type FinancialReportContext struct {
	GroupID        int64            `json:"group_id"`
	GroupStatus    enum.GroupStatus `json:"group_status"`
	ViewerMemberID int64            `json:"viewer_member_id"`
	Currency       string           `json:"currency"`
}

type FinancialReportMember struct {
	MemberID       int64   `json:"member_id"`
	UserID         int64   `json:"user_id"`
	Name           string  `json:"name"`
	AvatarURL      *string `json:"avatar_url"`
	IsFormerMember bool    `json:"is_former_member"`
}

type ExpenseReportEdge struct {
	finance.Transfer
	ShareID      int64      `json:"share_id"`
	ExpenseID    int64      `json:"expense_id"`
	ExpenseTitle string     `json:"expense_title"`
	OccurredAt   *time.Time `json:"occurred_at"`
}

type ExpenseReportItem struct {
	ExpenseReportEdge
	Counterparty FinancialReportMember `json:"counterparty"`
}

type ExpenseShareReport struct {
	FinancialReportContext
	Items      []ExpenseReportItem `json:"items"`
	Pagination PaginationResponse  `json:"pagination"`
}

type ExpenseGraphReport struct {
	FinancialReportContext
	Nodes []FinancialReportMember `json:"nodes"`
	Edges []ExpenseReportEdge     `json:"edges"`
}

type SettlementReportContext struct {
	FinancialReportContext
	BalanceRevision     string `json:"balance_revision"`
	CanRegisterPayments bool   `json:"can_register_payments"`
}

type SettlementGraphNode struct {
	FinancialReportMember
	NetAmountToman int64 `json:"net_amount_toman"`
}

type SettlementGraphReport struct {
	SettlementReportContext
	Nodes []SettlementGraphNode `json:"nodes"`
	Edges []finance.Transfer    `json:"edges"`
}

type SettlementReportItem struct {
	finance.Transfer
	Counterparty FinancialReportMember `json:"counterparty"`
}

type MySettlementReport struct {
	SettlementReportContext
	Outgoing            []SettlementReportItem `json:"outgoing"`
	Incoming            []SettlementReportItem `json:"incoming"`
	NetAmountToman      int64                  `json:"net_amount_toman"`
	TotalToPayToman     int64                  `json:"total_to_pay_toman"`
	TotalToReceiveToman int64                  `json:"total_to_receive_toman"`
}
