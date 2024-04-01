package main

import (
	"app/auth_grpc"
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/microsoftonline"

	"github.com/gorilla/sessions"

	gin_sessions "github.com/gin-contrib/sessions"
	gin_cookie "github.com/gin-contrib/sessions/cookie"
	csrf "github.com/utrack/gin-csrf"
	gin_csrf "github.com/utrack/gin-csrf"
)

func loadEnv() {
	// ここで.envファイル全体を読み込みます。
	// この読み込み処理がないと、個々の環境変数が取得出来ません。
	// 読み込めなかったら err にエラーが入ります。
	err := godotenv.Load(".env")
	
	// もし err がnilではないなら、"読み込み出来ませんでした"が出力されます。
	if err != nil {
		log.Panicf("読み込み出来ませんでした: %v", err)
	} 
}

//プロバイダ用の関数
func contextWithProviderName(ctx *gin.Context, provider string) *http.Request {
	return ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), "provider", provider))
}

func main() {
	loadEnv()

	router := gin.Default()

	//初期化処理

	//ミドルウェア初期化
	err := Init()

	//エラー処理
	if err != nil {
		panic(err)
	}

	//grpc初期化
	err = auth_grpc.Init("auth_Server:9000")

	//エラー処理
	if err != nil {
		panic(err)
	}

	//goth 初期化
	//TODO デバック用 
	key := "YZ68VIsXUpS8sBCKs22rSh0HDEixi15zIb8tialT1PfBNAHJZmuckKlu5Ji5a7cU"             // Replace with your SESSION_SECRET or similar
	maxAge := 86400 * 30  // 30 days
	isProd := false       // Set to true when serving over https

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true   // HttpOnly should always be enabled
	store.Options.Secure = isProd

	gothic.Store = store

	//プロバイダ初期化
	goth.UseProviders(
		microsoftonline.New(os.Getenv("Microsoft_ID"), os.Getenv("Microsoft_Secret"),os.Getenv("Microsoft_Redirect_Url")),
	)

	//セッション初期化
	session_store := gin_cookie.NewStore([]byte(key))
	router.Use(gin_sessions.Sessions("csrf_session", session_store))
	router.Use(gin_csrf.Middleware(gin_csrf.Options{
		Secret: key,
		ErrorFunc: func(ctx *gin.Context) {
			ctx.String(400, "CSRF token mismatch")
			ctx.Abort()
		},
	}))

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

		//csrf設定
		csrf.GetToken(ctx)

		ctx.SetSameSite(http.SameSiteLaxMode)
		ctx.SetCookie("token", result, 2592000, "/", "", true, true)
		
		ctx.Redirect(301,"/app/bind")
	})

	//メールアドレスが使用できないとき
	router.GET("/fail_mail", func(ctx *gin.Context) {
		ctx.String(500, "fail mail address")
	})

	//Outlook ログイン用エンドポイント
	router.GET("/outlook", func(ctx *gin.Context) {
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
		if !strings.HasSuffix(user.Email,"@ecc.ac.jp") {
			//ecc のメールアドレスではないとき
			ctx.Redirect(301, "/fail_mail")
			return
		}

		log.Println(user.Name)

		ctx.Redirect(301, "/app/allok")
	})

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
		data,exits := ctx.Get("user")

		//エラー処理
		if !exits {
			log.Println(err)
			ctx.JSON(500, gin.H{
				"message": "error",
			})
			return
		}

		//ユーザー取得
		user := data.(auth_grpc.User)
		log.Println(user.Name)

		ctx.JSON(200, gin.H{
			"message": "success",
		})
	})

	//ログアウト エンドポイント
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

	//Outlook 関連付け
	router.GET("/bind", func(ctx *gin.Context) {
		ctx.Request = contextWithProviderName(ctx, "microsoftonline")
		//認証開始
		gothic.BeginAuthHandler(ctx.Writer, ctx.Request)
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

	router.Run(":3010")
}
