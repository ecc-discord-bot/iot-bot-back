package main

//クラス
var classes = []string{}

func SyncClass() (error) {
	//値取得
	resp,err := service.Spreadsheets.Values.Get(spreadsheetID, "設定!E3:E").Do()

	//エラー処理
	if err != nil {
		return err
	}

	//初期化
	classes = []string{}

	//ループする
	for _, row := range resp.Values {
		//0この時
		if len(row) == 0 {
			continue
		}

		//値を追加
		classes = append(classes, row[0].(string))
	}

	return nil
}