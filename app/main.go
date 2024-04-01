package main

import (
	"app/auth_grpc"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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

		log.Println(result)
	})

	router.Run(":3001")
}
