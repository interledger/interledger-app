package runner

import (
	"testing"

	"github.com/interledger/interledger-app/go/performance/config"
	pb "github.com/interledger/interledger-app/go/proto/backend/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func balance(linkedAccount, asset string, amount int64) *pb.Balance {
	return &pb.Balance{
		LinkedAccount: linkedAccount,
		Balance:       &pb.Amount{Amount: amount, Asset: asset, AssetScale: 2},
	}
}

func TestSelectBalancePicksTheLargest(t *testing.T) {
	// With no account pinned, the biggest balance gives the longest drain run.
	balances := []*pb.Balance{
		balance("acc-small", "ZAR", 100),
		balance("acc-big", "ZAR", 9000),
		balance("acc-mid", "ZAR", 500),
	}

	got, err := selectBalance(balances, "", "")
	require.NoError(t, err)
	assert.Equal(t, "acc-big", got.GetLinkedAccount())
}

func TestSelectBalanceHonoursPinnedAccount(t *testing.T) {
	balances := []*pb.Balance{
		balance("acc-a", "ZAR", 100),
		balance("acc-b", "ZAR", 9000),
	}

	got, err := selectBalance(balances, "acc-a", "")
	require.NoError(t, err)
	assert.Equal(t, "acc-a", got.GetLinkedAccount(), "an explicit linked_account wins over the largest balance")
}

func TestSelectBalanceFiltersByAsset(t *testing.T) {
	balances := []*pb.Balance{
		balance("acc-usd", "USD", 9000),
		balance("acc-zar", "ZAR", 100),
	}

	got, err := selectBalance(balances, "", "ZAR")
	require.NoError(t, err)
	assert.Equal(t, "acc-zar", got.GetLinkedAccount(), "a pinned asset beats a larger balance in another currency")
}

func TestSelectBalanceErrors(t *testing.T) {
	tests := map[string]struct {
		balances      []*pb.Balance
		linkedAccount string
		asset         string
		wantErr       string
	}{
		"no balances at all": {
			balances: nil,
			wantErr:  "no balances",
		},
		"pinned account missing": {
			balances:      []*pb.Balance{balance("acc-a", "ZAR", 100)},
			linkedAccount: "acc-missing",
			wantErr:       `linked_account "acc-missing" not found`,
		},
		"no balance in the requested asset": {
			balances: []*pb.Balance{balance("acc-usd", "USD", 100)},
			asset:    "ZAR",
			wantErr:  `no balance in asset "ZAR"`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := selectBalance(tt.balances, tt.linkedAccount, tt.asset)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func makeSenders(n int) []*sender {
	senders := make([]*sender, n)
	for i := range n {
		senders[i] = &sender{cfg: config.Sender{Label: string(rune('a' + i))}}
	}
	return senders
}

func makeReceivers(n int) []config.Receiver {
	receivers := make([]config.Receiver, n)
	for i := range n {
		receivers[i] = config.Receiver{Label: string(rune('A' + i)), WalletAddress: "https://ilp.link/" + string(rune('A'+i))}
	}
	return receivers
}

func TestAssignReceiversIndexPairsOneToOne(t *testing.T) {
	// The 100 → 100 shape: every sender gets exactly its counterpart.
	senders := makeSenders(3)
	cfg := &config.Config{Receivers: makeReceivers(3)}
	cfg.Run.Pairing = config.PairingIndex

	assignReceivers(cfg, senders)

	for i, s := range senders {
		require.Len(t, s.receivers, 1)
		assert.Equal(t, cfg.Receivers[i].Label, s.receivers[0].Label)
	}
}

func TestAssignReceiversFanInTargetsTheFirst(t *testing.T) {
	// The 100 → 1 shape.
	senders := makeSenders(4)
	cfg := &config.Config{Receivers: makeReceivers(3)}
	cfg.Run.Pairing = config.PairingFanIn

	assignReceivers(cfg, senders)

	for _, s := range senders {
		require.Len(t, s.receivers, 1)
		assert.Equal(t, "A", s.receivers[0].Label)
	}
}

func TestAssignReceiversRoundRobinStaggersStart(t *testing.T) {
	// Without the stagger every sender would open by hammering receivers[0].
	senders := makeSenders(3)
	cfg := &config.Config{Receivers: makeReceivers(3)}
	cfg.Run.Pairing = config.PairingRoundRobin

	assignReceivers(cfg, senders)

	first := make([]string, len(senders))
	for i, s := range senders {
		require.Len(t, s.receivers, 3)
		first[i] = s.nextReceiver(config.PairingRoundRobin).Label
	}
	assert.Equal(t, []string{"A", "B", "C"}, first)
}

func TestNextReceiverRotates(t *testing.T) {
	s := &sender{receivers: makeReceivers(3)}

	got := make([]string, 0, 4)
	for range 4 {
		got = append(got, s.nextReceiver(config.PairingRoundRobin).Label)
	}
	assert.Equal(t, []string{"A", "B", "C", "A"}, got, "wraps around")
}

func TestNextReceiverRandomStaysInRange(t *testing.T) {
	s := &sender{receivers: makeReceivers(3)}
	valid := map[string]bool{"A": true, "B": true, "C": true}

	for range 50 {
		assert.True(t, valid[s.nextReceiver(config.PairingRandom).Label])
	}
}

func TestMaxPaymentsCapsDrainRuns(t *testing.T) {
	cfg := &config.Config{}
	cfg.Run.Stop = config.StopDrain
	cfg.Run.Amount = 1

	s := &sender{startBalance: 250}
	assert.Equal(t, 250, s.maxPayments(cfg), "balance/amount is the ceiling for a drain run")

	cfg.Run.Amount = 10
	assert.Equal(t, 25, s.maxPayments(cfg))
}

func TestMaxPaymentsIsUncappedOutsideDrain(t *testing.T) {
	cfg := &config.Config{}
	cfg.Run.Stop = config.StopDuration
	cfg.Run.Amount = 1

	s := &sender{startBalance: 250}
	assert.Equal(t, 0, s.maxPayments(cfg), "only a drain run is bounded by balance")
}

func TestIsTerminal(t *testing.T) {
	assert.False(t, isTerminal(stateCreated))
	assert.False(t, isTerminal(stateConfirmed))
	assert.False(t, isTerminal(stateProcessing))
	assert.True(t, isTerminal(stateCompleted))
	assert.True(t, isTerminal(stateFailed))
}
