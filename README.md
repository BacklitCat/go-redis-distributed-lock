# go-redis-distributed-lock
 
基于`Golang`和`Redis`，实现的分布式锁。

提供了两种锁方式：基于SetNX+PubSub和BLPop。

## 性能对比、正确性校验

6000个请求增加计数，4000个请求减少计数，共计10000个请求，分别测试正确性和性能：

```
OP nums: 10000
[LIST] expect: 2000 get: 2000 use time: 12.7059663
[NX] expect: 2000 get: 2000 use time: 16.3814946
```

## NxMutex

利用SetNX和PubSub实现。

**特点**：`阻塞式`，`竞争式`。

**过程**：
1. 所有用户Sub同一个频道；
2. Lock()：用户监听频道，有消息后尝试SetNX；
3. SetNX成功代表取得锁，进行操作；SetNX失败则继续监听频道；
4. Unlock()，先删除key，再Pub消息.

**实现细节**：
```go
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
```

**考虑死锁**：

如果Lock()后程序异常关闭，会死锁。

所以SetNX可以设置一个过期时间，这个过期时间必须大于冲突操作的时间。为了保证操作的原子性，应该通过Set方法同时指定NX和EX。

如果死锁，为了保证数据的正确，还应该回滚数据。

## ListMutex

利用队列的阻塞读BLPOP()避免锁竞争。

**特点**：`阻塞式`，`非竞争式`。不存在竞争，效率比利用SetNX和订阅要高。

**过程**：
1. 初始化一个只有一个元素的List;
2. Lock()：调用BLPop()，阻塞式pop;
3. Unlock(): Push()（几种方法都可以）一个元素进List.

**实现细节**：
```go
func (m *ListMutex) Lock() {
	_, err := m.db.BLPop(m.WaitTime, m.LockPath).Result()
	if err != nil {
		panic(err)
	}
}

func (m *ListMutex) Unlock() {
	_, err := m.db.RPush(m.LockPath, "lock").Result()
	if err != nil {
		panic(err)
	}
}
```

**考虑死锁**：
如果Lock()后程序异常关闭，会死锁。需要额外设计机制归还锁。

## 未来计划

加锁数据只存在单Redis节点上，如果发生主从切换，那么就会出现锁丢失的情况。 未来要学习使用RedLock。


