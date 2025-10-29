package app

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type WorkerPool struct {
	act.Pool
}

func FactoryWorkerPool() gen.ProcessBehavior {
	return &WorkerPool{}
}

func (workerPool *WorkerPool) Init(args ...any) (act.PoolOptions, error) {
	poolOptions := act.PoolOptions{
		WorkerFactory: FactoryWorkerActor,
		PoolSize:      3,
		WorkerArgs:    []any{10},
	}

	workerPool.Log().Info("Started worker pool %v %v on %v", workerPool.PID(), workerPool.Name(), workerPool.Node().Name())
	return poolOptions, nil
}
