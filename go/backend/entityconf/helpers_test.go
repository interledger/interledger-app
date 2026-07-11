package entityconf_test

import "github.com/interledger/interledger-app/go/backend/entityconf"

const (
	testWallet     entityconf.EntityType = "wallet"
	testBattleShip entityconf.EntityType = "battleship"
)

// walletConfsFixture mirrors the shape a real WalletConfs struct would
// take: a mix of bool/int/string confs, an intentionally-skipped field, and
// an unexported field that must never be treated as a conf.
type walletConfsFixture struct {
	SendEnabled    bool   `conf:"wallet.send_enabled" default:"true" display:"Send Payments" desc:"Allows sending"`
	ReceiveEnabled bool   `conf:"wallet.receive_enabled" default:"false" display:"Receive Payments" desc:"Allows receiving"`
	MaxDailyLimit  int    `conf:"wallet.max_daily_limit" default:"1000" display:"Max Daily Limit" desc:"Max amount per day"`
	Nickname       string `conf:"wallet.nickname" default:"" display:"Nickname" desc:"Optional wallet nickname"`
	internalOnly   bool   //nolint:unused // deliberately unexported and untagged: Register must ignore it
	Skipped        bool   `conf:"-"`
}

// shipConfs mirrors plan.md §5.9's BattleShip worked example.
type shipConfs struct {
	Name           string `conf:"battleship.name" default:"USS Unnamed" display:"Name" desc:"The display name of the battleship"`
	Length         int    `conf:"battleship.length" default:"250" display:"Length (m)" desc:"Overall hull length, in meters"`
	HasFrontTurret bool   `conf:"battleship.has_front_turret" default:"true" display:"Front Turret" desc:"Whether the battleship carries a forward gun turret"`
	HasBackTurret  bool   `conf:"battleship.has_back_turret" default:"true" display:"Back Turret" desc:"Whether the battleship carries a rear gun turret"`
}
