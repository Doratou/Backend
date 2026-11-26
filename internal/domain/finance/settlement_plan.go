package finance

import (
	"context"
	"math/bits"
)

// MinimumTransfers returns an exact minimum plan sorted deterministically by member ID.
func MinimumTransfers(ctx context.Context, positions []Position) ([]Transfer, error) {
	positions = nonzeroPositions(positions)
	var total int64
	for _, position := range positions {
		if position.NetToman > 0 {
			var err error
			total, err = addAmount(total, position.NetToman)
			if err != nil {
				return nil, err
			}
		}
	}
	if len(positions) <= 18 {
		return subsetPlan(ctx, positions)
	}
	return searchPlan(ctx, positions)
}

func nonzeroPositions(positions []Position) []Position {
	result := make([]Position, 0, len(positions))
	for _, position := range positions {
		if position.NetToman != 0 {
			result = append(result, position)
		}
	}
	return result
}

func subsetPlan(ctx context.Context, positions []Position) ([]Transfer, error) {
	size := 1 << len(positions)
	sums := make([]int64, size)
	components := make([]int, size)
	chosen := make([]int, size)
	for mask := 1; mask < size; mask++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		lowest := mask & -mask
		sums[mask] = sums[mask^lowest] + positions[bits.TrailingZeros(uint(lowest))].NetToman
		best := -1
		for candidates := mask; candidates != 0; candidates &= candidates - 1 {
			bit := candidates & -candidates
			if score := components[mask^bit]; score > best {
				best, chosen[mask] = score, bit
			}
		}
		components[mask] = best
		if sums[mask] == 0 {
			components[mask]++
		}
	}

	result := make([]Transfer, 0)
	var group []Position
	for mask := size - 1; mask != 0; {
		bit := chosen[mask]
		group = append(group, positions[bits.TrailingZeros(uint(bit))])
		mask ^= bit
		if sums[mask] == 0 {
			result = append(result, settleComponent(group)...)
			group = group[:0]
		}
	}
	return result, nil
}

func searchPlan(ctx context.Context, positions []Position) ([]Transfer, error) {
	best := settleComponent(positions)
	residual := make([]int64, len(positions))
	for i, position := range positions {
		residual[i] = position.NetToman
	}
	type join struct{ from, to int }
	path := make([]join, 0, len(positions))

	var search func(int)
	search = func(start int) {
		if ctx.Err() != nil {
			return
		}
		for start < len(residual) && residual[start] == 0 {
			start++
		}
		if start == len(residual) {
			parent := make([]int, len(positions))
			for i := range parent {
				parent[i] = i
			}
			for _, link := range path {
				parent[link.from] = link.to
			}
			groups := make([][]Position, len(positions))
			for i, position := range positions {
				root := i
				for parent[root] != root {
					root = parent[root]
				}
				groups[root] = append(groups[root], position)
			}
			best = make([]Transfer, 0, len(path))
			for _, group := range groups {
				best = append(best, settleComponent(group)...)
			}
			return
		}

		var debtors, creditors int
		for _, amount := range residual[start:] {
			if amount < 0 {
				debtors++
			} else if amount > 0 {
				creditors++
			}
		}
		if len(path)+max(debtors, creditors) >= len(best) {
			return
		}
		first := residual[start]
		seen := make(map[int64]bool)
		for i := start + 1; i < len(residual) && ctx.Err() == nil; i++ {
			other := residual[i]
			if other == 0 || (first > 0) == (other > 0) || seen[other] {
				continue
			}
			seen[other] = true
			residual[i] = other + first
			path = append(path, join{start, i})
			search(start + 1)
			path = path[:len(path)-1]
			residual[i] = other
			if other+first == 0 {
				break
			}
		}
	}
	search(0)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return best, nil
}

func settleComponent(positions []Position) []Transfer {
	debtors := make([]Position, 0, len(positions))
	creditors := make([]Position, 0, len(positions))
	for _, position := range positions {
		if position.NetToman < 0 {
			position.NetToman = -position.NetToman
			debtors = append(debtors, position)
		} else if position.NetToman > 0 {
			creditors = append(creditors, position)
		}
	}
	edges := make([]Transfer, 0)
	for i, j := 0, 0; i < len(debtors) && j < len(creditors); {
		amount := min(debtors[i].NetToman, creditors[j].NetToman)
		edges = append(edges, Transfer{FromMemberID: debtors[i].MemberID, ToMemberID: creditors[j].MemberID, AmountToman: amount})
		debtors[i].NetToman -= amount
		creditors[j].NetToman -= amount
		if debtors[i].NetToman == 0 {
			i++
		}
		if creditors[j].NetToman == 0 {
			j++
		}
	}
	return edges
}
