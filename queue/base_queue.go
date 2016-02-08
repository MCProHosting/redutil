package queue

import (
	"sync"

	"github.com/garyburd/redigo/redis"
)

// BaseQueue provides a basic implementation of the Queue interface. Its basic
// methodology is to preform updates using a Processor interface which in and of
// itself defines how updates can be handled.
type BaseQueue struct {
	pool   *redis.Pool
	source string

	pmu       sync.RWMutex
	processor Processor
}

var _ Queue = new(BaseQueue)

func NewBaseQueue(pool *redis.Pool, source string) *BaseQueue {
	return &BaseQueue{
		pool:   pool,
		source: source,
	}
}

// Source implements the Source method on the Queue interface.
func (q *BaseQueue) Source() string {
	return q.source
}

// Push pushes the given payload (a byte slice) into the specified keyspace by
// delegating into the `Processor`'s `func Push`. It obtains a connection to
// Redis using the pool, which is passed into the Processor, and recycles that
// connection after the function has returned.
//
// If an error occurs during Pushing, it will be returned, and it can be assumed
// that the payload is not in Redis.
func (q *BaseQueue) Push(payload []byte) (err error) {
	cnx := q.pool.Get()
	defer cnx.Close()

	return q.Processor().Push(cnx, q.Source(), payload)
}

// Source implements the Source method on the Queue interface.
func (q *BaseQueue) Pull() (payload []byte, err error) {
	cnx := q.pool.Get()
	defer cnx.Close()

	return q.Processor().Pull(cnx, q.Source())
}

// Source implements the Source method on the Queue interface. It functions by
// requesting a read-level lock from the guarding mutex and returning that value
// once obtained. If no processor is set, the the default FIFO implementation is
// returned.
func (q *BaseQueue) Processor() Processor {
	q.pmu.RLock()
	defer q.pmu.RUnlock()

	if q.processor == nil {
		return FIFO
	}

	return q.processor
}

// SetProcessor implements the SetProcessor method on the Queue interface. It
// functions by requesting write-level access from the guarding mutex and
// preforms the update atomically.
func (q *BaseQueue) SetProcessor(processor Processor) {
	q.pmu.Lock()
	defer q.pmu.Unlock()

	q.processor = processor
}
