package auth_grpc

import (
	"errors"
	"log"
	"os"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	conn   *grpc.ClientConn = nil
	client AuthServerClient = nil
	isinit bool             = false
)

func Init(ServerUrl string) error {
	// 9000番ポートでクライアントからのリクエストを受け付けるようにする
	creds, err := credentials.NewClientTLSFromFile(os.Getenv("TLS_CRT"), "")
	if err != nil {
		log.Fatal(err)
	}

	//サーバーとの接続
	dial_conn, err := grpc.Dial(ServerUrl, grpc.WithTransportCredentials(creds))

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

// 認証する
func Auth(token string) (User, error) {
	//初期化されているか
	if !isinit {
		return User{}, errors.New("not init")
	}

	//認証する
	result, err := client.Auth(context.Background(), &AuthToken{Token: token})

	//エラー処理
	if err != nil {
		log.Printf("Error when calling SayHello: %s", err)
		return User{}, err
	}

	//成功していない場合
	if !result.Success {
		return User{}, errors.New("auth error")
	}

	return *result.User, nil
}

// トークンを取得する
func GetToken(token string, secret string) (string, error) {
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
		return "", err
	}

	//成功していない場合
	if !response.Success {
		return "", errors.New("get token error")
	}

	return response.Token, nil
}

// ログアウト
func Logout(token string) error {
	//初期化されているか
	if !isinit {
		return nil
	}

	//トークンを無効化する
	result, err := client.Logout(context.Background(), &AuthToken{Token: token})

	//エラー処理
	if err != nil {
		log.Printf("Error when calling SayHello: %s", err)
		return err
	}

	//成功していない場合
	if !result.Success {
		log.Printf("Error when calling SayHello: %s", err)
		return errors.New("logout error")
	}

	return nil
}

// トークンを更新する
func Refresh(token string, UserAgent string) (string, error) {
	//初期化されているか
	if !isinit {
		return "", errors.New("not init")
	}

	//トークン取得
	response, err := client.Refresh(context.Background(), &AuthToken{
		Token:     token,
		UserAgent: UserAgent,
	})

	//エラー処理
	if err != nil {
		log.Printf("Error when calling SayHello: %s", err)
		return "", err
	}

	//成功していない場合
	if !response.Success {
		return "", errors.New("refresh token error")
	}

	return response.Token, nil
}

// トークンを更新する
func Submit_Refresh(token string) error {
	//初期化されているか
	if !isinit {
		return errors.New("not init")
	}

	//トークン取得
	response, err := client.Submit(context.Background(), &AuthToken{
		Token: token,
	})

	//エラー処理
	if err != nil {
		log.Printf("Error when calling SayHello: %s", err)
		return err
	}

	//成功していない場合
	if !response.Success {
		return errors.New("refresh token error")
	}

	return nil
}
