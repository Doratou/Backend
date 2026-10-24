package bootstrap

import "time"

type Constant struct{ Database DBConst }

type DBConst struct {
	MaxOpenDbConn int32
	MaxIdleDbConn time.Duration
	MaxDbLifeTime time.Duration
}

func NewConstant() *Constant {
	return &Constant{Database: DBConst{
		MaxOpenDbConn: 10,
		MaxIdleDbConn: 5 * time.Minute,
		MaxDbLifeTime: 5 * time.Minute,
	}}
}
