package transaction

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"

	"strings"
)

var (
	Secret = []byte("hAUKO10XC6Ck8aUeu8LsEwFT5k5QatzXaVXTcVhJ7JCXVafVWkQ7ExyAADRCCF8t")
)

// トークン生成
func GenToken() (TokenData, error) {
	//トークンのID生成
	tokenid := genid()

	//現在時刻
	now_time := time.Now()

	//トークン有効期限
	after_time := now_time.Add(time.Minute * 5)

	//トークンデータ
	token_claims := jwt.MapClaims{
		"tokenid": tokenid,
		"exp":     after_time.Unix(),
	}

	//トークン生成
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, token_claims)

	//署名
	signed_str, err := token.SignedString(Secret)

	//エラー処理
	if err != nil {
		return TokenData{}, err
	}

	return TokenData{
		Token:   signed_str,
		Tokenid: tokenid,
	}, nil
}

// トークン検証 (トークンIDを返す)
func ParseToken(tokenString string) (string, error) {
	//トークン検証
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//トークンの署名方法を確認
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		//検証できてたら、秘密鍵を返す
		return Secret, nil
	})

	//エラー処理
	if err != nil {
		return "", err
	}

	//クライム取得
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["tokenid"].(string), nil
	}

	return "", err
}

func genid() string {
	//ID生成
	uid := uuid.New()

	//ハイフンを削除
	returnid := strings.ReplaceAll(uid.String(), "-", "")

	return returnid
}
