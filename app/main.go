package main

import (
	"app/auth_grpc"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/markbates/goth/gothic"

	gin_sessions "github.com/gin-contrib/sessions"
	gin_cookie "github.com/gin-contrib/sessions/cookie"
	//gin_csrf "github.com/utrack/gin-csrf"
)

// プロバイダ用の関数
func contextWithProviderName(ctx *gin.Context, provider string) *http.Request {
	return ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), "provider", provider))
}

func main() {
	//初期化
	Init()

	router := gin.Default()

	//セッション初期化
	session_store := gin_cookie.NewStore([]byte(key))
	router.Use(gin_sessions.Sessions("csrf_session", session_store))

	//ここまで

	//ミドルウェア設定
	router.Use(Auth())

	//トークンを取得するエンドポイント
	router.GET("/", func(ctx *gin.Context) {
		//トークンを取得する
		token := ctx.DefaultQuery("token", "")

		//トークンがあるか
		if token == "" {
			//トークンがない場合
			ctx.JSON(500, gin.H{
				"message": "no token",
			})
			return
		}

		//トークンを取得する
		result, err := auth_grpc.GetToken(token, os.Getenv("TokenSecret"))

		//エラー処理
		if err != nil {
			log.Println(err)
			ctx.JSON(500, gin.H{
				"message": "error",
			})
			return
		}

		ctx.SetSameSite(http.SameSiteLaxMode)
		ctx.SetCookie("token", result, 2592000, "/", "", true, true)

		ctx.Redirect(301, "/app/bind")
	})

	//トークン更新用エンドポイント
	router.POST("/refresh", func(ctx *gin.Context) {
		//認証されているか
		if !ctx.GetBool("auth") {
			//認証していない場合
			ctx.JSON(401, gin.H{
				"message": "not auth",
			})
			return
		}

		//ユーザーエージェント取得
		UserAgent := ctx.GetHeader("User-Agent")
		//トークンを取得する
		new_token, err := auth_grpc.Refresh(ctx.GetString("token"), UserAgent)

		//エラー処理
		if err != nil {
			log.Println(err)
			ctx.JSON(500, gin.H{
				"message": "error",
			})
			return
		}

		//cookie設定
		ctx.SetSameSite(http.SameSiteLaxMode)
		ctx.SetCookie("token", new_token, 2592000, "/", "", true, true)

		ctx.JSON(200, gin.H{
			"message": "ok",
		})
	})

	//トークン更新確定エンドポイント
	router.POST("/refreshs", func(ctx *gin.Context) {
		//認証されているか
		if !ctx.GetBool("auth") {
			//認証していない場合
			ctx.JSON(401, gin.H{
				"message": "not auth",
			})
			return
		}

		//トークンを取得する
		err := auth_grpc.Submit_Refresh(ctx.GetString("token"))

		//エラー処理
		if err != nil {
			log.Println(err)
			ctx.JSON(500, gin.H{
				"message": "error",
			})
			return
		}

		ctx.JSON(200, gin.H{
			"message": "ok",
		})
	})

	//メールアドレスが使用できないとき
	router.GET("/fail_mail", func(ctx *gin.Context) {
		ctx.String(500, "fail mail address")
	})

	//Outlook ログイン用エンドポイント
	router.GET("/outlook", Outlook)

	//outlook のログインが終わったとき
	router.GET("/allok", func(ctx *gin.Context) {
		//ログインしているか
		if !ctx.GetBool("auth") {
			//認証していない場合
			ctx.JSON(401, gin.H{
				"message": "no auth",
			})
			return
		}

		//ユーザーを取得する
		data, exits := ctx.Get("user")

		//エラー処理
		if !exits {
			ctx.JSON(500, gin.H{
				"message": "error",
			})
			return
		}

		//すでに登録されているか検索する
		result, err := GetLastRow(data.(auth_grpc.User).ProviderUserId)

		//エラー処理
		if err != nil {
			ctx.JSON(500, gin.H{
				"message": "error",
			})
			return
		}

		//ユーザ取得
		appuser, err := dbconn.GetUser(data.(auth_grpc.User).ProviderUserId)

		//エラー処理
		if err != nil {
			log.Println(err)
			ctx.JSON(500, gin.H{
				"message": "error",
			})
			return
		}

		log.Println(result)
		//登録されていないとき
		if !result.Isfind {
			err := WriteUser(fmt.Sprintf("管理シート!B%s", strconv.Itoa(result.Total+3)), User{
				DiscordID:  data.(auth_grpc.User).ProviderUserId,
				StudentsID: appuser.Students_id,
				Name:       appuser.Name,
				Class:      appuser.Class,
				IsPaid:     appuser.Is_paid,
				IsAgreed:   appuser.Is_agreed,
				Time:       appuser.NowTime,
				Signature:  "",
			})

			//エラー処理
			if err != nil {
				log.Println(err)
				ctx.JSON(500, gin.H{
					"message": "error",
				})
				return
			}
		}

		//同意画面にリダイレクト
		ctx.Redirect(301, "/terms")
		//ユーザー取得
		//dbconn.GetUser(data.(auth_grpc.User).ProviderUserId)

		//ctx.HTML(http.StatusOK, "index.html", gin.H{})
	})

	//ログインエンドポイント
	router.GET("/login", func(ctx *gin.Context) {
		ctx.Redirect(301, "/auth/discord?redirect_url="+os.Getenv("Redirect_URL"))
	})

	//ログアウト エンドポイント
	/*
		router.GET("/logout", func(ctx *gin.Context) {
			//認証成功しているか
			if !ctx.GetBool("auth") {
				//認証していない場合
				ctx.JSON(401, gin.H{
					"message": "no auth",
				})
				return
			}

			//トークンを取得する
			token := ctx.GetString("token")

			//ログアウト
			err := auth_grpc.Logout(token)

			//エラー処理
			if err != nil {
				log.Println(err)
				ctx.JSON(500, gin.H{
					"message": "error",
				})
				return
			}

			//トークン削除
			ctx.SetCookie("token", "", -1, "/", "", true, true)
			ctx.JSON(200, gin.H{
				"message": "success",
			})
		})
	*/

	//Outlook 関連付け
	router.GET("/bind", func(ctx *gin.Context) {
		ctx.Request = contextWithProviderName(ctx, "microsoftonline")
		//認証開始
		gothic.BeginAuthHandler(ctx.Writer, ctx.Request)
	})

	//ユーザ情報取得
	router.GET("/userinfo", func(ctx *gin.Context) {
		//認証されているか
		if !ctx.GetBool("auth") {
			//認証していない場合
			ctx.JSON(401, gin.H{
				"message": "no auth",
			})
			return
		}

		//ユーザーを取得する
		data, exits := ctx.Get("user")

		//エラー処理
		if !exits {
			ctx.JSON(500, gin.H{
				"message": "error",
			})
			return
		}

		//ユーザー取得
		user, err := dbconn.GetUser(data.(auth_grpc.User).ProviderUserId)

		//エラー処理
		if err != nil {
			log.Println(err)
			ctx.JSON(500, gin.H{
				"message": "error",
			})
			return
		}

		ctx.JSON(200, gin.H{
			"message": "ok",
			"user":    user,
		})
	})

	router.POST("/logout", func(ctx *gin.Context) {
		//認証成功しているか
		if !ctx.GetBool("auth") {
			//認証していない場合
			ctx.JSON(401, gin.H{
				"message": "no auth",
			})
			return
		}

		//トークンを取得する
		token := ctx.GetString("token")

		//ログアウト
		err := auth_grpc.Logout(token)

		//エラー処理
		if err != nil {
			log.Println(err)
			ctx.JSON(500, gin.H{
				"message": "error",
			})
			return
		}

		ctx.JSON(200, gin.H{
			"message": "success",
		})
	})

	type AgreeData struct {
		Class     string
		Agree     bool
		Signature string
	}

	//同意するエンドポイント
	router.POST("/agree", func(ctx *gin.Context) {
		//json 取得
		var json AgreeData
		if err := ctx.ShouldBindJSON(&json); err != nil {
			log.Println(err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//認証成功しているか
		if !ctx.GetBool("auth") {
			//認証していない場合
			ctx.JSON(401, gin.H{
				"message": "no auth",
			})
			return
		}

		//データ検証
		if json.Signature == "" {
			ctx.JSON(400, gin.H{
				"message": "error",
			})
			return
		}

		//ユーザ取得
		user := ctx.MustGet("user").(auth_grpc.User)

		//ユーザーオブジェクト取得
		userObj, err := dbconn.GetUser(user.ProviderUserId)

		//エラー処理
		if err != nil {
			ctx.JSON(500, gin.H{
				"message": "error",
			})
			return
		}

		//同意した時刻を取得する
		if !userObj.Is_agreed {
			//同意した時刻を取得
			now_time := time.Now()

			//現在時刻を設定
			userObj.NowTime = now_time.Unix()
		}

		//同意済み
		userObj.Class = json.Class
		//常にtrue
		userObj.Is_agreed = true

		userObj.Signature = json.Signature

		//ユーザーオブジェクト更新
		err = dbconn.UpdateUser(userObj)

		//エラー処理
		if err != nil {
			ctx.JSON(500, gin.H{
				"message": "error",
			})
			return
		}

		//すでに登録されているか検索する
		result, err := GetLastRow(user.ProviderUserId)

		//エラー処理
		if err != nil {
			ctx.JSON(500, gin.H{
				"message": "error",
			})
			return
		}

		//ユーザ取得
		appuser := userObj

		//登録されているとき
		if result.Isfind {
			err := WriteUser(fmt.Sprintf("管理シート!B%s", strconv.Itoa(result.Index+3)), User{
				DiscordID:  user.ProviderUserId,
				StudentsID: appuser.Students_id,
				Name:       appuser.Name,
				Class:      appuser.Class,
				IsPaid:     appuser.Is_paid,
				IsAgreed:   appuser.Is_agreed,
				Time:       appuser.NowTime,
				Signature:  appuser.Signature,
			})

			//エラー処理
			if err != nil {
				ctx.JSON(500, gin.H{
					"message": "error",
				})
				return
			}
		}

		ctx.JSON(200, gin.H{
			"message": "success",
		})
	})

	router.Run(":3010")
}
