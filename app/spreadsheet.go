package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"google.golang.org/api/option"

	"google.golang.org/api/sheets/v4"
)

var spreadsheetID = ""
var service *sheets.Service = nil
var findstr = "管理シート!B3:B"

func SpreadsheetInit() {
	credential := option.WithCredentialsFile("./google_cred.json")

	srv, err := sheets.NewService(context.TODO(), credential)
	if err != nil {
		log.Fatal(err)
	}

	service = srv

	spreadsheetID = os.Getenv("SPREADSHEET_ID")
}

type User struct {
	DiscordID  string
	StudentsID string
	Name       string
	Class      string
	IsPaid     bool
	IsAgreed   bool
	Time       int64
	Signature  string
}

func WriteUser(range_str string, data User) error {
	vr := &sheets.ValueRange{
		Values: [][]interface{}{
			{
				//DIscord ID
				data.DiscordID,
				//学籍番号
				data.StudentsID,
				//本名
				data.Name,
				//クラス
				data.Class,
				//同意済みか
				data.IsAgreed,
				//支払い済みか
				data.IsPaid,
				//同意時間
				fmt.Sprintf("=EPOCHTODATE(%s)",strconv.FormatInt(data.Time,10)),
				//署名
				data.Signature,
			},
		},
	}

	//書き込み
	resp, err := service.Spreadsheets.Values.Update(spreadsheetID, range_str, vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Println(err)
		return err
	}

	//ステータスコード
	log.Println(resp.HTTPStatusCode)
	return nil
}

type Result struct {
	Isfind bool
	Index  int
	Total  int
}

func GetLastRow(find_value string) (Result, error) {
	//値取得
	resp, err := service.Spreadsheets.Values.Get(spreadsheetID, findstr).Do()

	//エラー処理
	if err != nil {
		return Result{Isfind: false}, err
	}

	isfind := false
	findindex := -1

	//１行ずつ回す
	for index, row := range resp.Values {
		//一致したとき抜ける
		if row[0] == find_value {
			isfind = true
			findindex = index
			break
		}
	}

	//行数を返す
	total_index := len(resp.Values)

	return Result{Isfind: isfind, Index: findindex, Total: total_index}, nil
}
