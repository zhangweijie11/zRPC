package consumer

type FailMode int

const (
	Failover  FailMode = iota // 故障转移，换个服务端重试
	Failfast                  // 接受失败，不再重试
	Failretry                 // 临时失败，直接重试
)
