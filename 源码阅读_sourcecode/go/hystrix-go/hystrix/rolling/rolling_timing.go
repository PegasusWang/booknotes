package rolling

import (
	"math"
	"sort"
	"sync"
	"time"
)

type timingBucket struct {
	Durations []time.Duration
}

type Timing struct {
	Buckets map[int64]*timingBucket
	Mutex   *sync.RWMutex

	CachedSortedDurations []time.Duration
	LastCachedTime        int64
}

func NewTiming() *Timing {
	r := &Timing{
		Buckets: make(map[int64]*timingBucket),
		Mutex:   &sync.RWMutex{},
	}
	return r
}

type byDuration []time.Duration

func (c byDuration) Len() int           { return len(c) }
func (c byDuration) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c byDuration) Less(i, j int) bool { return c[i] < c[j] }

func (r *Timing) SortedDuration() []time.Duration {
	r.Mutex.RLock()
	t := r.LastCachedTime
	r.Mutex.RUnlock()

	if t+time.Duration(1*time.Second).Nanoseconds() > time.Now().UnixNano() {
		return r.CachedSortedDurations
	}
	var durations byDuration
	now := time.Now()
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	for timestamp, b := range r.Buckets {
		if timestamp >= now.Unix()-60 {
			for _, d := range b.Durations {
				durations = append(durations, d)
			}
		}
	}
	sort.Sort(durations)
	r.CachedSortedDurations = durations
	r.LastCachedTime = time.Now().UnixNano()
	return r.CachedSortedDurations
}

func (r *Timing) getCurrentBucket() *timingBucket {
	r.Mutex.RLock()
	now := time.Now()
	bucket, exists := r.Buckets[now.Unix()]
	r.Mutex.RUnlock()

	if !exists {
		r.Mutex.Lock()
		defer r.Mutex.Unlock()

		r.Buckets[now.Unix()] = &timingBucket{}
		bucket = r.Buckets[now.Unix()]
	}
	return bucket
}

func (r *Timing) removeOldBuckets() {
	now := time.Now()
	for timestamp := range r.Buckets {
		if timestamp <= now.Unix()-60 {
			delete(r.Buckets, timestamp)
		}
	}
}

func (r *Timing) Add(duration time.Duration) {
	b := r.getCurrentBucket()
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	b.Durations = append(b.Durations, duration)
	r.removeOldBuckets()
}

// 百分位数，统计学术语，如果将一组数据从小到大排序，并计算相应的累计百分位，则某一百分位所对应数据的值就称为这一百分位的百分位数，以Pk表示第k百分位数。
func (r *Timing) Percentile(p float64) uint32 {
	sd := r.SortedDuration()
	length := len(sd)
	if length <= 0 {
		return 0
	}
	pos := r.ordinal(len(sd), p) - 1
	return uint32(sd[pos].Nanoseconds() / 1000000)
}

func (r *Timing) ordinal(length int, percentile float64) int64 {
	if percentile == 0 && length > 0 {
		return 1
	}

	return int64(math.Ceil((percentile / float64(100)) * float64(length)))
}

// Mean computes the average timing in the last 60 seconds.
func (r *Timing) Mean() uint32 {
	sd := r.SortedDuration()
	var sum time.Duration

	for _, d := range sd {
		sum += d
	}

	length := int64(len(sd))
	if length == 0 {
		return 0
	}
	return uint32(sum.Nanoseconds()/length) / 1000000
}
