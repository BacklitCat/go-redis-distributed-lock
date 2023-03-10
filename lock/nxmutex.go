package lock

import (
	"github.com/go-redis/redis"
	"time"
)

type NxMutex struct {
	db          *redis.Client
	LockPath    string
	ChannelPath string
	ch          <-chan *redis.Message
	LockTime    time.Duration
}

func NewNxMutex(db *redis.Client, lockName string, lockTime time.Duration) (*NxMutex, error) {
	// 检查连接
	_, err := db.Ping().Result()
	if err != nil {
		return nil, err
	}
	if lockTime < 0 {
		lockTime = time.Duration(0) // NO TTL LIMIT
	}
	channelPath := "NxMutex:Channel:" + lockName
	ps := db.Subscribe(channelPath)
	return &NxMutex{
		db:          db,
		LockPath:    "NxMutex:EXIST:" + lockName,
		ChannelPath: channelPath,
		ch:          ps.Channel(),
		LockTime:    lockTime,
	}, err
}

func (m *NxMutex) Lock() {
	for {
		created, err := m.db.SetNX(m.LockPath, "lock", m.LockTime).Result()
		if err != nil {
			panic(err)
		}
		if created {
			break
		}
		<-m.ch // wait pub
	}
}

func (m *NxMutex) Unlock() {
	m.db.Del(m.LockPath)
	m.db.Publish(m.ChannelPath, "unlock")
}
