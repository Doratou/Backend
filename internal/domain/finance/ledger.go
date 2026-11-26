package finance

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
	"slices"
	"time"
)

var ErrAmountOutOfRange = errors.New("financial balance exceeds the supported range")

type Transfer struct {
	FromMemberID int64 `json:"from_member_id"`
	ToMemberID   int64 `json:"to_member_id"`
	AmountToman  int64 `json:"amount_toman"`
}

type Position struct {
	MemberID int64
	NetToman int64
}

type RecordedCharge struct {
	Transfer
	RecordedAt time.Time `json:"recorded_at"`
}

type RecordedPayment struct {
	Transfer
	ConfirmedAt time.Time
}

type OpeningMember struct {
	ID       int64      `json:"id"`
	JoinedAt time.Time  `json:"joined_at"`
	LeftAt   *time.Time `json:"left_at"`
	Active   bool       `json:"active"`
}

type OpeningState struct {
	Members []OpeningMember  `json:"members"`
	Charges []RecordedCharge `json:"charges"`
}

type BalanceData struct {
	ActiveMembers []int64
	OpenPairs     []Transfer
	Opening       OpeningState
	Legacy        []RecordedPayment
}

type memberPair struct{ low, high int64 }

// Ledger keeps the signed net liability between each pair of members.
type Ledger struct {
	active map[int64]bool
	pairs  map[memberPair]int64
}

func NewLedger(data BalanceData) (*Ledger, error) {
	ledger := &Ledger{active: make(map[int64]bool, len(data.ActiveMembers)), pairs: make(map[memberPair]int64)}
	for _, id := range data.ActiveMembers {
		ledger.active[id] = true
	}
	for _, debt := range data.OpenPairs {
		if err := ledger.add(debt); err != nil {
			return nil, err
		}
	}
	legacy, err := legacyPaymentLegs(data.Opening, data.Legacy)
	if err != nil {
		return nil, err
	}
	for _, leg := range legacy {
		if !ledger.active[leg.FromMemberID] || !ledger.active[leg.ToMemberID] {
			continue
		}
		if err := ledger.add(reverse(leg)); err != nil {
			return nil, err
		}
	}
	return ledger, nil
}

func legacyPaymentLegs(opening OpeningState, payments []RecordedPayment) ([]Transfer, error) {
	ledger := &Ledger{pairs: make(map[memberPair]int64)}
	legs := make([]Transfer, 0, len(payments))
	nextCharge := 0
	for _, payment := range payments {
		for nextCharge < len(opening.Charges) && !opening.Charges[nextCharge].RecordedAt.After(payment.ConfirmedAt) {
			if err := ledger.add(opening.Charges[nextCharge].Transfer); err != nil {
				return nil, err
			}
			nextCharge++
		}
		ledger.active = make(map[int64]bool, len(opening.Members))
		for _, member := range opening.Members {
			if !member.JoinedAt.After(payment.ConfirmedAt) &&
				(member.Active || (member.LeftAt != nil && member.LeftAt.After(payment.ConfirmedAt))) {
				ledger.active[member.ID] = true
			}
		}
		applied, err := ledger.ApplyPayment(payment.Transfer)
		if err != nil {
			return nil, err
		}
		legs = append(legs, applied...)
	}
	return legs, nil
}

func (l *Ledger) add(t Transfer) error {
	if t.FromMemberID == t.ToMemberID || t.AmountToman == 0 {
		return nil
	}
	key, value := memberPair{t.FromMemberID, t.ToMemberID}, t.AmountToman
	if key.low > key.high {
		key.low, key.high = key.high, key.low
		value = -value
	}
	sum, err := addAmount(l.pairs[key], value)
	if err != nil {
		return err
	}
	if sum == 0 {
		delete(l.pairs, key)
	} else {
		l.pairs[key] = sum
	}
	return nil
}

func reverse(t Transfer) Transfer {
	return Transfer{FromMemberID: t.ToMemberID, ToMemberID: t.FromMemberID, AmountToman: t.AmountToman}
}

func addAmount(a, b int64) (int64, error) {
	if (b > 0 && a > math.MaxInt64-b) || (b < 0 && a < -math.MaxInt64-b) {
		return 0, ErrAmountOutOfRange
	}
	return a + b, nil
}

func (l *Ledger) obligations() []Transfer {
	debts := make([]Transfer, 0, len(l.pairs))
	for pair, amount := range l.pairs {
		if !l.active[pair.low] || !l.active[pair.high] {
			continue
		}
		edge := Transfer{FromMemberID: pair.low, ToMemberID: pair.high, AmountToman: amount}
		if amount < 0 {
			edge = reverse(edge)
			edge.AmountToman = -amount
		}
		debts = append(debts, edge)
	}
	slices.SortFunc(debts, func(a, b Transfer) int {
		if a.FromMemberID != b.FromMemberID {
			return compareID(a.FromMemberID, b.FromMemberID)
		}
		return compareID(a.ToMemberID, b.ToMemberID)
	})
	return debts
}

func compareID(a, b int64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func (l *Ledger) Positions() ([]Position, error) {
	net := make(map[int64]*big.Int, len(l.active))
	for id := range l.active {
		net[id] = new(big.Int)
	}
	for _, edge := range l.obligations() {
		amount := big.NewInt(edge.AmountToman)
		net[edge.FromMemberID].Sub(net[edge.FromMemberID], amount)
		net[edge.ToMemberID].Add(net[edge.ToMemberID], amount)
	}
	positions := make([]Position, 0, len(l.active))
	for id, amount := range net {
		if !amount.IsInt64() || amount.Int64() == math.MinInt64 {
			return nil, ErrAmountOutOfRange
		}
		positions = append(positions, Position{MemberID: id, NetToman: amount.Int64()})
	}
	slices.SortFunc(positions, func(a, b Position) int { return compareID(a.MemberID, b.MemberID) })
	return positions, nil
}

func (l *Ledger) Revision() string {
	hash := sha256.New()
	ids := make([]int64, 0, len(l.active))
	for id := range l.active {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	for _, id := range ids {
		fmt.Fprintf(hash, "m:%d;", id)
	}
	for _, edge := range l.obligations() {
		fmt.Fprintf(hash, "e:%d:%d:%d;", edge.FromMemberID, edge.ToMemberID, edge.AmountToman)
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// ApplyPayment returns postings for a real transfer and updates the ledger in place.
func (l *Ledger) ApplyPayment(payment Transfer) ([]Transfer, error) {
	legs := make([]Transfer, 0, 3)
	affected := make(map[int64]bool)
	remaining := payment.AmountToman
	post := func(leg Transfer) error {
		legs = append(legs, leg)
		affected[leg.FromMemberID], affected[leg.ToMemberID] = true, true
		return l.add(reverse(leg))
	}
	for remaining > 0 && l.active[payment.FromMemberID] && l.active[payment.ToMemberID] {
		debts := l.obligations()
		path := debtPath(debts, payment.FromMemberID, payment.ToMemberID)
		if len(path) == 0 {
			break
		}
		amount := remaining
		for _, debt := range path {
			amount = min(amount, debt.AmountToman)
		}
		for _, debt := range path {
			debt.AmountToman = amount
			if err := post(debt); err != nil {
				return nil, err
			}
		}
		remaining -= amount
	}
	for remaining > 0 && l.active[payment.FromMemberID] && l.active[payment.ToMemberID] {
		// A minimum plan can match two members from different debt chains.
		// Reduce their old edges and carry the remaining liability between the other endpoints.
		var outgoing, incoming Transfer
		for _, edge := range l.obligations() {
			if outgoing.AmountToman == 0 && edge.FromMemberID == payment.FromMemberID {
				outgoing = edge
			}
			if incoming.AmountToman == 0 && edge.ToMemberID == payment.ToMemberID {
				incoming = edge
			}
		}
		if outgoing.AmountToman == 0 || incoming.AmountToman == 0 {
			break
		}
		amount := min(remaining, outgoing.AmountToman, incoming.AmountToman)
		outgoing.AmountToman, incoming.AmountToman = amount, amount
		for _, leg := range []Transfer{outgoing, incoming, {FromMemberID: outgoing.ToMemberID, ToMemberID: incoming.FromMemberID, AmountToman: amount}} {
			if err := post(leg); err != nil {
				return nil, err
			}
		}
		remaining -= amount
	}
	if remaining > 0 {
		payment.AmountToman = remaining
		if err := post(payment); err != nil {
			return nil, err
		}
	}
	if l.active[payment.FromMemberID] && l.active[payment.ToMemberID] {
		for {
			cycle := affectedDebtCycle(l.obligations(), affected)
			if len(cycle) == 0 {
				break
			}
			// Record offsets in the cleared cycle so departure cannot reopen them.
			amount := cycle[0].AmountToman
			for _, edge := range cycle[1:] {
				amount = min(amount, edge.AmountToman)
			}
			for _, edge := range cycle {
				edge.AmountToman = amount
				if err := post(edge); err != nil {
					return nil, err
				}
			}
		}
	}
	return legs, nil
}

func affectedDebtCycle(debts []Transfer, affected map[int64]bool) []Transfer {
	for _, edge := range debts {
		if affected[edge.FromMemberID] {
			path := debtPath(debts, edge.ToMemberID, edge.FromMemberID)
			if len(path) > 0 {
				return append(path, edge)
			}
		}
	}
	return nil
}

func debtPath(debts []Transfer, from, to int64) []Transfer {
	adjacent := make(map[int64][]Transfer)
	for _, debt := range debts {
		adjacent[debt.FromMemberID] = append(adjacent[debt.FromMemberID], debt)
	}
	parent := map[int64]Transfer{from: {}}
	queue := []int64{from}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, edge := range adjacent[current] {
			if _, seen := parent[edge.ToMemberID]; seen {
				continue
			}
			parent[edge.ToMemberID] = edge
			if edge.ToMemberID == to {
				path := []Transfer{edge}
				for edge.FromMemberID != from {
					edge = parent[edge.FromMemberID]
					path = append(path, edge)
				}
				return path
			}
			queue = append(queue, edge.ToMemberID)
		}
	}
	return nil
}
