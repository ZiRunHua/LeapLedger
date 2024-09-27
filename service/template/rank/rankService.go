package tmplRank

import (
	"KeepAccount/global"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strconv"
	"sync"
	"time"
)

var rdb *redis.Client

func init() {
	rdb = global.GvaRdb
}

type Member interface {
	fmt.Stringer
}

func NewRank[memberType Member](uniqueKey string, members []memberType, periodValidity time.Duration) Rank[memberType] {
	rank := Rank[memberType]{
		memberKeyMap:   make(map[string]memberType),
		uniqueKey:      uniqueKey,
		periodValidity: periodValidity,
		useZsetKey:     ":use_zset",
		rankZsetKey:    "rank_" + uniqueKey + ":zset",
		lock:           &sync.Mutex{},
	}
	rank.addMembers(members)

	return rank
}

type Rank[memberType Member] struct {
	members                            []memberType
	memberKeyMap                       map[string]memberType
	periodValidity                     time.Duration
	uniqueKey, useZsetKey, rankZsetKey string
	lock                               *sync.Mutex
}

func (r *Rank[memberType]) Init(ctx context.Context) error {
	for _, member := range r.members {
		_, err := r.OnceIncrWeight(member, 0, time.Now().Unix(), ctx)
		if err != nil {
			return err
		}
	}
	err := r.updateRank(ctx)
	if err != nil {
		return err
	}
	return nil
}
func (r *Rank[memberType]) getMemberKey(id memberType, key string) string {
	return "rank_" + r.uniqueKey + ":member" + id.String() + key
}
func (r *Rank[memberType]) addMembers(members []memberType) {
	r.lock.Lock()
	r.members = append(r.members, members...)
	for _, member := range members {
		r.memberKeyMap[member.String()] = member
	}
	r.lock.Unlock()
}

func (r *Rank[memberType]) OnceIncrWeight(useOb memberType, Key uint, value int64, ctx context.Context) (int64, error) {
	return rdb.ZAdd(ctx, r.getMemberKey(useOb, r.useZsetKey), &redis.Z{float64(value), Key}).Result()
}

func (r *Rank[memberType]) updateRank(ctx context.Context) (err error) {
	var memberWidget int64
	maxValue := "(" + strconv.FormatInt(time.Now().Add(-r.periodValidity).Unix(), 10)
	for _, member := range r.members {
		_, err = rdb.ZRemRangeByScore(ctx, r.getMemberKey(member, r.useZsetKey), "-inf", maxValue).Result()
		if err != nil {
			return err
		}
		memberWidget, err = rdb.ZCard(ctx, r.getMemberKey(member, r.useZsetKey)).Result()
		if err != nil {
			return err
		}
		_, err = rdb.ZAdd(ctx, r.rankZsetKey, &redis.Z{float64(memberWidget), member.String()}).Result()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Rank[memberType]) GetAll(ctx context.Context) ([]memberType, error) {
	members, err := rdb.ZRevRangeByScoreWithScores(ctx, r.rankZsetKey,
		&redis.ZRangeBy{Min: "-inf", Max: "+inf"}).Result()
	if err != nil {
		return nil, err
	}
	result := make([]memberType, len(members), len(members))
	for i := 0; i < len(result); i++ {
		member, ok := r.memberKeyMap[members[i].Member.(string)]
		if !ok {
			continue
		}
		result[i] = member
	}
	return result, nil
}
func (r *Rank[memberType]) Clear(ctx context.Context) {
	rdb.Del(ctx, r.rankZsetKey)
	for _, member := range r.members {
		rdb.Del(ctx, r.getMemberKey(member, r.useZsetKey))
	}
}
