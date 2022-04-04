package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

/*
 一个简单的版本的 snow flake 分布式 id 生成器 实现
原始的 snowflake 使用 41bits 作为毫秒数， 10 bit 机器 id (5位数据中心+ 5位机器)， 12 bit 作为毫秒内的流水号
意味着每个节点每毫秒可以生成 4096 个 id，最后一个符号位保留，永远是 0

[0] [41] [10] [12]

*/

const (
	//StartTimeEnvName is the env key for ID generating start time . 开始时间环境变量
	StartTimeEnvName = "FASTID_START_TIME"
	//MachineIDEnvName is the env key for machine id
	MachineIDEnvName           = "FASTID_MACHINE_ID" // 机器 id 环境变量，如果没有设置会自动用 ip
	defaultStartTimeStr        = "2018-06-01T00:00:00.000Z"
	defaultStartTimeNano int64 = 1527811200000000000
)

// id 生成器的 配置
type Config struct {
	timeBits      uint
	seqBits       uint
	machineBits   uint
	timeMask      int64
	seqMask       int64
	machineID     int64
	machineIDMask int64
	lastID        int64
}

func NewConfig(timeBits, seqBits, machineBits uint) *Config {
	return NewConfigWithMachineID(timeBits, seqBits, machineBits, getMachineID())
}

func NewConfigWithMachineID(timeBits, seqBits, machineBits uint, machineID int64) *Config {
	machineIdMask := ^(int64(-1) << machineBits) // TODO 位操作
	return &Config{
		timeBits:      timeBits,
		seqBits:       seqBits,
		machineBits:   machineBits,
		timeMask:      ^(int64(-1) << timeBits),
		seqMask:       ^(int64(-1) << seqBits),
		machineIDMask: machineIdMask,
		machineID:     machineID & machineIdMask,
		lastID:        0,
	}
}

var CommonConfig = NewConfig(40, 15, 8)
var startEpochNano = getStartEpochFromEnv()

func (c *Config) getCurrentTimestamp() int64 {
	return (time.Now().UnixNano() - startEpochNano) >> 20 & c.timeMask
}

func (c *Config) GenInt64ID() int64 {
	for {
		localLastID := atomic.LoadInt64(&c.lastID)
		seq := c.GetSeqFromID(localLastID)
		lastIDTime := c.GetTimeFromID(localLastID)
		now := c.getCurrentTimestamp()
		if now > lastIDTime {
			seq = 0
		} else if seq >= c.seqMask {
			time.Sleep(time.Duration(0xFFFFF - (time.Now().UnixNano() & 0xFFFFF)))
			continue
		} else {
			seq++
		}
		newID := now<<(c.machineBits+c.seqBits) + seq<<c.machineBits + c.machineID
		if atomic.CompareAndSwapInt64(&c.lastID, localLastID, newID) {
			return newID
		}
		time.Sleep(time.Duration(20))
	}
}

func (c *Config) GetSeqFromID(id int64) int64 {
	return (id >> c.machineBits) & c.seqMask
}

func (c *Config) GetTimeFromID(id int64) int64 {
	return id >> (c.machineBits + c.seqBits)
}

// 获取机器 id
func getMachineID() int64 {
	if machineIDStr, ok := os.LookupEnv(MachineIDEnvName); ok {
		if machineID, err := strconv.ParseInt(machineIDStr, 10, 64); err == nil {
			return machineID
		}
	}

	if ip, err := getIP(); err == nil {
		fmt.Println(ip, ip[2], ip[3])
		return (int64(ip[2]) << 8) + int64(ip[3])
	}
	return 0
}

func getIP() (net.IP, error) {
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok {
				if !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
					ip := ipNet.IP.To4()
					// TODO 限制网段？
					if ip[0] == 10 || ip[0] == 172 && (ip[1] >= 16 && ip[1] < 32) || ip[0] == 192 && ip[1] == 168 {
						return ip, nil
					}
				}
			}
		}
	}
	return nil, errors.New("Failed to get ip address")
}

func getStartEpochFromEnv() int64 {
	startTimeStr := getEnv(StartTimeEnvName, defaultStartTimeStr)
	var startEpochTime, err = time.Parse(time.RFC3339, startTimeStr)

	// 原来的作者代码我怀疑这个地方写错了。应该是  =nil
	if err != nil {
		return defaultStartTimeNano
	}

	return startEpochTime.UnixNano()
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func main() {
	for i := 0; i < 10; i++ {
		n := CommonConfig.GenInt64ID()
		fmt.Println(n)
	}
}
