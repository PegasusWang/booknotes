package myhystrix

//用了一个简单的令牌算法，能得到令牌的就可以执行后继的工作，执行完后要返还令牌。得不到令牌就拒绝，拒绝后调用用户设置的callback方法，如果没有设置就不执行。
type executorPool struct {
	Name    string
	Metrics *poolMetrics // TODO
	Max     int
	Tickets chan *struct{}
}

func newExecutorPool(name string) *executorPool {
	p := &executorPool{}
	p.Name = name
	p.Metrics = newPoolMetrics(name)
	p.Max = getSettings(name).MaxConcurrentRequests

	p.Tickets = make(chan *struct{}, p.Max)
	for i := 0; i < p.Max; i++ {
		p.Tickets <- &struct{}{}
	}
	return p
}

// 令牌用完是要返还的
func (p *executorPool) Return(ticket *struct{}) {
	if ticket == nil {
		return
	}
	p.Metrics.Updates <- poolMetricsUpdate{
		activeCount: p.ActiveCount(),
	}
	p.Tickets <- ticket
}

func (p executorPool) ActiveCount() int {
	return p.Max - len(p.Tickets)
}
