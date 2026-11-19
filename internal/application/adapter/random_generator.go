package adapter

type RandomGenerator interface {
	Int64n(max int64) (int64, error)
}
