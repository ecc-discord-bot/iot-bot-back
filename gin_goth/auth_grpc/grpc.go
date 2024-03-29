package auth_grpc

import (
	"errors"
	"log"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
)

var (
	conn *grpc.ClientConn = nil
	client AuthServerClient = nil
	isinit bool = false
)

func Init(ServerUrl string) error {
	// 9000番ポートでクライアントからのリクエストを受け付けるようにする
	dial_conn, err := grpc.Dial(ServerUrl, grpc.WithInsecure())
	if err != nil {
		log.Printf("did not connect: %s \n", err)
		return err
	}

	//グローバル変数にする
	conn = dial_conn

	//クライアントとの接続
	client = NewAuthServerClient(conn)

	isinit = true

	return nil
}

func GetToken(token string,secret string) (string,error) {
	//初期化されているか
	if !isinit {
		return "", errors.New("not init")
	}

	//トークン取得
	response, err := client.GetToken(context.Background(), &Secret{
		Secret: secret,
		Token:  token,
	})

	//エラー処理
	if err != nil {
		log.Printf("Error when calling SayHello: %s", err)
		return "",err
	}

	//成功していない場合
	if !response.Success {
		return "", errors.New("get token error")
	}

	return response.Token, nil
}