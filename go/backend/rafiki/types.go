package rafiki

const (
	Provider          = "rafiki"
	ZARBalanceAccount = "905e2c9b-a8b7-4cf2-9449-197e7029052e"
)

type Grant struct {
	Id                 string
	Client             string
	State              string
	FinalizationReason string
}
