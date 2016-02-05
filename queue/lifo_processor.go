package queue

import "github.com/garyburd/redigo/redis"

type lifoProcessor struct{}

// FIFO is a last in, first out implementation of the Processor interface.
var LIFO Processor = &lifoProcessor{}

// Push implements the `func Push` from `Processor`. It pushes the right-side
// of the Redis structure using RPUSH, and returns any errors encountered while
// runnning that command.
func (l *lifoProcessor) Push(cnx redis.Conn, src string, payload []byte) (err error) {
	_, err = cnx.Do("RPUSH", src, payload)
	return
}

// Pull implements the `func Pull` from `Processor`. It pulls from the right-side
// of the Redis structure in a blocking-fashion, using BRPOP.
//
// If an redis.ErrNil is returned, it is silenced, and both fields are returend
// as nil. If the err is not a redis.ErrNil, but is still non-nil itself, then
// it will be returend, along with an empty []byte.
//
// If an item can successfully be removed from the keyspace, it is returned
// without error.
func (l *lifoProcessor) Pull(cnx redis.Conn, src string) ([]byte, error) {
	slices, err := redis.ByteSlices(cnx.Do("BRPOP", src, 0))
	if err == redis.ErrNil {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return slices[1], nil
}

// PullTo implements the `func PullTo` from the `Processor` interface. It pulls
// from the right-side of the Redis source (src) structure, and pushes to the
// left side of the Redis destination (dest) structure.
func (l *lifoProcessor) PullTo(cnx redis.Conn, src, dest string) ([]byte, error) {
	bytes, err := redis.Bytes(cnx.Do("BRPOPLPUSH"))
	if err == redis.ErrNil {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// Removes the first element from the source list and adds it to the end
// of the destination list. ErrNil is returns when the source is empty.
func (l *lifoProcessor) Concat(cnx redis.Conn, src, dest string) (err error) {
	return rlConcat(cnx, src, dest)
}
