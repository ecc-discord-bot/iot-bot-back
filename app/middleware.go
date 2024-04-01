package main

import (
	"app/auth_grpc"

	"github.com/gin-gonic/gin"
)

var (
	isinit = false
)

func Init() error {
	//grpcの初期化
	err := auth_grpc.Init("auth_Server:9000")

	//エラー処理
	if err != nil {
		return err
	}

	//初期化済みにする
	isinit = true

	return nil
}

func Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//情報初期化
		ctx.Set("auth", false)
		ctx.Set("user", auth_grpc.User{})
		ctx.Set("token", "")

		//初期化されているか
		if !isinit {
			//初期化されていない場合
			ctx.Next()
			return
		}

		//トークン取得
		token,err := GetToken(ctx)

		//エラー処理
		if err != nil {
			ctx.Next()
			return
		}

		//トークンが存在するか
		if token == "" {
			//トークンが存在しない場合
			return
		}

		//認証する
		user, err := auth_grpc.Auth(token)

		//エラー処理
		if err != nil {
			return
		}

		//認証成功
		ctx.Set("auth", true)
		ctx.Set("user", user)
		ctx.Set("token", token)

		ctx.Next()
	}
}

func GetToken(ctx *gin.Context) (string,error) {
	//token := ctx.Request.Header.Get("token")
	//return token,nil

	return ctx.Cookie("token")
}