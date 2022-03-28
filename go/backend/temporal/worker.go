package temporal

import "go.temporal.io/sdk/worker"

func NewTemporalWorker() (worker.Worker, error) {
	c, err := NewTemporalClient()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	w := worker.New(c, "backend", worker.Options{})

	return w, nil
}