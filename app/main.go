package main

import (
	"app/auth_grpc"
	"context"
	"log"
	"net/http"
	"os"

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

	//TODO デバック用
	router.LoadHTMLGlob("templates/*")

	//セッション初期化
	session_store := gin_cookie.NewStore([]byte(key))
	router.Use(gin_sessions.Sessions("csrf_session", session_store))

	/*
	router.Use(gin_csrf.Middleware(gin_csrf.Options{
		Secret: key,
		ErrorFunc: func(ctx *gin.Context) {
			ctx.String(400, "CSRF token mismatch")
			ctx.Abort()
		},
	}))
	*/

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

		_ = data

		//同意画面にリダイレクト
		ctx.Redirect(301, "/terms")
		//ユーザー取得
		//dbconn.GetUser(data.(auth_grpc.User).ProviderUserId)

		//ctx.HTML(http.StatusOK, "index.html", gin.H{})
	})

	//ログインエンドポイント
	router.GET("/login", func(ctx *gin.Context) {
		ctx.Redirect(301, "/auth/discord?redirect_url=https://127.0.0.1:8443/app/")
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
		user,err := dbconn.GetUser(data.(auth_grpc.User).ProviderUserId)

		//エラー処理
		if err != nil {
			log.Println(err)
			ctx.JSON(500, gin.H{
				"message": "error",
			})
			return
		}

		log.Println(user)
		ctx.JSON(200, gin.H{
			"message": "ok",
			"user": user,
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
		Class string
		Agree bool
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

		//同意済み
		userObj.Class = json.Class
		//常にtrue
		userObj.Is_agreed = true

		//ユーザーオブジェクト更新
		err = dbconn.UpdateUser(userObj)

		//エラー処理
		if err != nil {
			ctx.JSON(500, gin.H{
				"message": "error",
			})
			return
		}

		ctx.JSON(200, gin.H{
			"message": "success",
		})
	})

	router.Run(":3010")
}
