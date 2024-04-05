package main

import (
	"app/auth_grpc"
	"app/db"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
	"gorm.io/gorm"
)

func Outlook(ctx *gin.Context) {
	//すでにdiscord でログインしているか
	isauth := ctx.GetBool("auth")

	//ログインしているか
	if !isauth {
		//ログインしていない場合
		ctx.JSON(401, gin.H{
			"message": "not auth",
		})
		return
	}

	//プロバイダ設定
	ctx.Request = contextWithProviderName(ctx, "microsoftonline")

	//認証
	user, err := gothic.CompleteUserAuth(ctx.Writer, ctx.Request)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	//メールアドレス判定
	if !strings.HasSuffix(user.Email, "@ecc.ac.jp") {
		//ecc のメールアドレスではないとき
		ctx.Redirect(301, "/fail_mail")
		return
	}

	//ユーザーを作成する
	auth_user := ctx.MustGet("user").(auth_grpc.User)

	log.Println(auth_user.ProviderUserId)
	log.Println(user.Name)

	//ユーザーを取得
	usr,err := dbconn.GetUser(auth_user.ProviderUserId)

	//エラー処理
	if err == gorm.ErrRecordNotFound {
		//登録されていないとき作成する
		err = dbconn.CreateUser(db.User{
			Discord_id: auth_user.ProviderUserId,
			Students_id: strings.TrimRight(user.Email, "@ecc.ac.jp"),
			Name: user.Name,
			Class: "",
			Is_agreed: false,
			Is_paid: false,
		})

		//エラー処理
		if err != nil {
			log.Println(err)
			ctx.JSON(500, gin.H{
				"message": "error",
			})
			return
		}
	} else if err != nil {
		//エラーのとき
		log.Println(err)
		ctx.JSON(500, gin.H{
			"message": "error",
		})
	} else {
		//ユーザーを更新する
		usr.Name = user.Name
		usr.Students_id = strings.TrimSuffix(user.Email, "@ecc.ac.jp")
		
		err = dbconn.UpdateUser(usr)

		//エラー処理
		if err != nil {
			log.Println(err)
			ctx.JSON(500, gin.H{
				"message": "error",
			})
			return
		}
	}

	ctx.Redirect(301, "/app/allok")
}