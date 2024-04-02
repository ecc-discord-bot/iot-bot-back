package main

import (
	"errors"
	"gin_oauth/auth_grpc"
	"gin_oauth/transaction"
	"log"
	"net"
	"os"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"gin_oauth/auth"
)

func start_grpc() {
	log.Println("Start grpc server")

	// 9000番ポートでクライアントからのリクエストを受け付けるようにする
	listen, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	//GRPCサーバ
	grpcServer := grpc.NewServer()

	// Auth構造体のアドレスを渡すことで、クライアントからGetDataリクエストされると
	// GetDataメソッドが呼ばれるようになる
	auth_grpc.RegisterAuthServerServer(grpcServer, &Auth{})

	// 以下でリッスンし続ける
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}

	log.Print("grpc end")
}

// Auth構造体
type Auth struct{}

// トークンで認証
func (auther *Auth) Auth(
	ctx context.Context,
	token *auth_grpc.AuthToken,
) (*auth_grpc.AuthResult, error) {
	//トークン検証
	token_data, err := auth.ParseToken(token.Token)
	if err != nil {
		//失敗した場合エラー返す
		return &auth_grpc.AuthResult{Success: false}, err
	}

	//ユーザ取得
	user, err := auth.GetUserInfo(token_data.BindId)
	if err != nil {
		//失敗した場合エラー返す
		return &auth_grpc.AuthResult{Success: false}, err
	}
	
	//認証結果
	result := auth_grpc.AuthResult{
		Success: true,
		User: &auth_grpc.User{
			UserID:         user.UserID,         //ユーザID
			Name:           user.Name,           //ユーザ名
			Provider:       user.Provider,       //認証プロバイダー
			ProviderUserId: user.ProviderUserId, //認証プロバイダーのユーザID
			NickName:       user.NickName,       //ニックネーム
			LastName:       user.LastName,       //姓
			FirstName:      user.FirstName,      //名
			AvatarURL:      user.AvatarURL,      //アバターURL
			Email:          user.Email,          //メールアドレス
			Description:    user.Description,    //説明文
		},
	}

	//成功した場合、認証結果を返す
	return &result, nil
}

func (auther *Auth) Refresh(
	ctx context.Context,
	token *auth_grpc.AuthToken,
) (*auth_grpc.RefreshResult, error) {
	//トークン検証
	token_data, err := auth.ParseToken(token.Token)

	//エラー処理
	if err != nil {
		log.Println(err)
		//失敗した場合エラー返す
		return &auth_grpc.RefreshResult{Success: false}, err
	}

	//新しいトークン発行
	new_token, err := auth.UpdateToken(token_data.TokenId, "")

	//エラー処理
	if err != nil {
		log.Println(err)
		//失敗した場合エラー返す
		return &auth_grpc.RefreshResult{Success: false}, err
	}

	return &auth_grpc.RefreshResult{Success: true, Token: new_token}, nil
}

//ログアウト関数
func (auther *Auth) Logout(
	ctx context.Context,
	token *auth_grpc.AuthToken,
) (*auth_grpc.Result, error) {
	//トークンを無効化する
	err := auth.DisableToken(token.Token)

	//エラー処理
	if err != nil {
		log.Println(err)
		//失敗した場合エラー返す
		return &auth_grpc.Result{Success: false}, err
	}

	//成功した場合
	return &auth_grpc.Result{Success: true}, nil
}

func (auther *Auth) GetToken(
	ctx context.Context,
	data *auth_grpc.Secret,
) (*auth_grpc.TokenData, error) {
	//シークレット検証
	if data.Secret != os.Getenv("ClientSecret") {
		//失敗した場合エラー返す
		return &auth_grpc.TokenData{
			Token: "",
			Success: false,
		}, errors.New("secret error")
	}

	//トークンをパースする
	tokenid,err := transaction.ParseToken(data.Token)

	//エラー処理
	if err != nil {
		log.Println(err)
		//失敗した場合エラー返す
		return &auth_grpc.TokenData{
			Token: "",
			Success: false,
		}, err
	}

	//トークン取得
	token,err := transaction.Get(tokenid)

	//エラー処理
	if err != nil {
		log.Println(err)

		//失敗した場合エラー返す
		return &auth_grpc.TokenData{
			Token: "",
			Success: false,
		}, err
	}

	//トークン削除
	err = transaction.Delete(tokenid)

	//エラー処理
	if err != nil {
		log.Println(err)
		return &auth_grpc.TokenData{
			Token: "",
			Success: false,
		}, err
	}

	return &auth_grpc.TokenData{
		Token: token,
		Success: true,
	}, nil
}
