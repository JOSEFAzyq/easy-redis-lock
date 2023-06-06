package easyredisllock

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

type DistributeLock interface {
	//
	// GetLock
	//  @Description: 获取锁
	//  @param ctx
	//  @param lockKey	锁名称
	//  @param lockTime	上锁时间,到期自动释放防止死锁
	//  @return string	锁内容
	//  @return error
	//
	GetLock(ctx *gin.Context, lockKey string, lockTime int) (string, error)
	//
	// UnLock
	//  @Description:释放锁
	//  @param ctx
	//  @param lockKey	锁名称
	//  @param lockValue	锁内容,需传入上锁得到的锁值,防止释放错误锁
	//  @return error
	//
	UnLock(ctx *gin.Context, lockKey string, lockValue string) error
}

type RedisLock struct {
	Store *redis.Client
}

func (r *RedisLock) GetLock(ctx *gin.Context, lockKey string, lockTime int) (string, error) {
	var lockValue string
	lockValue = strconv.Itoa(int(time.Now().Unix()))
	resp := r.Store.SetNX(ctx, lockKey, lockValue, time.Millisecond*time.Duration(lockTime))
	lockSuccess, err := resp.Result()
	if err != nil {
		return "", err
	}
	if lockSuccess {
		return lockValue, nil
	} else {
		return "", errors.New("锁被他人持有")
	}
}

func (r *RedisLock) UnLock(ctx *gin.Context, lockKey string, lockValue string) error {
	unlockScript := redis.NewScript(`
	local lockKey = KEYS[1]
	local lockValue = KEYS[2]
	local localKeyValue = redis.call("GET",lockKey)
	redis.log(redis.LOG_NOTICE,"aaaa","lockValue",lockValue,"localKeyValue",localKeyValue)
	if localKeyValue == lockValue then
		return redis.call("DEL",lockKey)
	else
		return 0
	end
`)
	keys := []string{lockKey, lockValue}
	rs, err := unlockScript.Run(ctx, r.Store, keys).Int()
	if err != nil {
		return err
	} else {
		if rs == 1 {
			return nil
		} else {
			return errors.New("释放锁失败")
		}
	}

}
