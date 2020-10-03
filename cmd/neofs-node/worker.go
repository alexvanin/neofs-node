package main

import (
	"context"
)

type worker interface {
	Run(context.Context)
}

func startWorkers(c *cfg) {
	for _, wrk := range c.workers {
		c.wg.Add(1)

		go func() {
			wrk.Run(c.ctx)
			c.wg.Done()
		}()
	}
}