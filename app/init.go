package main

import (
	"app/db"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/microsoftonline"

	"github.com/gorilla/sessions"

	"github.com/markbates/goth"
)

//グローバル変数
var (
	dbconn db.UserI = nil

	//TODO デバック用
	key = "YZ68VIsXUpS8sBCKs22rSh0HDEixi15zIb8tialT1PfBNAHJZmuckKlu5Ji5a7cU" // Replace with your SESSION_SECRET or similar
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

func Init() {
	//env読み込み
	loadEnv()

	//データーベース初期化
	conn,err := db.Init()

	if err != nil {
		log.Panicln(err)
	}

	//ミドルウェア初期化
	//grpcも初期化される
	err = Middleware_Init()

	if err != nil {
		log.Panicln(err)
	}

	//goth 初期化
	maxAge := 86400 * 30                                                      // 30 days
	isProd := false                                                           // Set to true when serving over https

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true // HttpOnly should always be enabled
	store.Options.Secure = isProd

	gothic.Store = store

	//プロバイダ初期化
	goth.UseProviders(
		microsoftonline.New(os.Getenv("Microsoft_ID"), os.Getenv("Microsoft_Secret"), os.Getenv("Microsoft_Redirect_Url")),
	)

	//グローバル変数に設定
	dbconn = conn

	//スプレットシート初期化
	SpreadsheetInit()
}