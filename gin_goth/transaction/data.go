package transaction

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redis_conn *redis.Client = nil
	isinit = false
)

//トークンデータ
type TokenData struct {
	Token string
	Tokenid string
}

//初期化
func Init() error {
	//Redis接続
	redis_conn = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
		PoolSize: 1000,
	})

	//初期化完了
	isinit = true

	return nil
}

//一時的にデータを保存する ()
func Save(id string,exp time.Duration,data string) error {
	if !isinit {
		return errors.New("not init")
	}

	//コンテキスト取得
	ctx := context.Background()

	//データ登録
	result := redis_conn.Set(ctx,id,data, exp)

	//エラー処理
	if result.Err() != nil {
		return result.Err()
	}

	return nil
}

//データを取得する
func Get(id string) (string,error) {
	if !isinit {
		return "",errors.New("not init")
	}

	//コンテキスト取得
	ctx := context.Background()

	//データ登録
	result := redis_conn.Get(ctx,id)

	//エラー処理
	if result.Err() != nil {
		return "",result.Err()
	}

	//データ取得
	data,err := result.Result()

	//エラー処理
	if err != nil {
		return "",err
	}

	return data,nil
}

//データを削除する
func Delete(id string) error {
	if !isinit {
		return errors.New("not init")
	}

	//コンテキスト取得
	ctx := context.Background()

	//データ削除
	result := redis_conn.Del(ctx,id)

	//エラー処理
	if result.Err() != nil {
		return result.Err()
	}

	return nil
}

