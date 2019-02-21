package process

import (
	. "../jobQueue"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"strconv"
	"time"
)

var (
	srcDb *sql.DB
	desDb []*sql.DB
)

func InitDb(maxIdleConns int) {
	db, err := sql.Open("mysql", Config.Src.Dsn)
	if err != nil {
		panic(err)
	}
	db.SetMaxIdleConns(maxIdleConns + 1)
	db.SetConnMaxLifetime(time.Minute * 10)
	srcDb = db

	lens := len(Config.Des)
	desDb = make([]*sql.DB, lens)
	for i := 0; i < lens; i++ {
		db, err = sql.Open("mysql", Config.Des[i].Dsn)
		if err != nil {
			panic(err)
		}
		db.SetMaxIdleConns(maxIdleConns)
		db.SetConnMaxLifetime(time.Minute * 10)
		desDb[i] = db
	}
}

func GetMaxId() int {

	rows, err := srcDb.Query("SELECT max(Id) as maxId from " + Config.Src.Table)
	defer rows.Close()
	CheckErr(err)
	var maxId int
	for rows.Next() {
		err = rows.Scan(&maxId)
		Logger.Info("insert定时获取的maxId", zap.String("srcTable", Config.Src.Table), zap.Int("maxId", maxId))
		CheckErr(err)
	}

	return maxId
}

func GetUpdateId() []int {

	end := time.Now()
	d, _ := time.ParseDuration("-" + strconv.Itoa(Config.Src.UpdateScanSecond) + "s")
	start := end.Add(d)
	updateSql := "SELECT Id from " + Config.Src.Table + " where " + Config.Src.UpdateColumn + " between '" + start.Format(Config.Src.UpdateTimeFormate) + "' and '" + end.Format(Config.Src.UpdateTimeFormate) + "'"
	rows, err := srcDb.Query(updateSql)
	defer rows.Close()
	CheckErr(err)
	var updateId int
	ids := make([]int, 0)

	for rows.Next() {
		err = rows.Scan(&updateId)
		Logger.Info("update定时扫描到的updateId", zap.String("srcTable", Config.Src.Table), zap.Int("updateId", updateId))
		CheckErr(err)
		ids = append(ids, updateId)
	}

	return ids
}

func CompareColumn(srcId uint) {

	srcArray := query(srcDb, "SELECT * FROM "+Config.Src.Table+" where Id= "+strconv.Itoa(int(srcId)))
	if len(srcArray) != 0 {
		lens := len(Config.Des)
		for i := 0; i < lens; i++ {
			desArray := query(desDb[i], "SELECT * FROM "+Config.Des[i].Table+" where "+Config.Des[i].ByColumn+"= "+srcArray[Config.Src.ByColumn].(string))
			if len(desArray) == 0 {
				Logger.Warn("数据比对，发现des数据缺少, 发送至insert回调", zap.Int("srcId", int(srcId)),
					zap.String("srcTable", Config.Src.Table),
					zap.String("desTable", Config.Des[i].Table),
					zap.String("desByColumn", Config.Des[i].ByColumn),
					zap.String("desByColumnValue", srcArray[Config.Src.ByColumn].(string)))
				CallbackJobQueue <- &CallbackJob{Types: INSERT, SrcArray: srcArray, CallbackUrl: Config.Des[i].CallbackNotification.Url}
			} else {
				for keyColumn, Column := range Config.Des[i].Columns {
					_, okSrc := srcArray[keyColumn]
					_, okDes := desArray[Column]
					var configSrcByColumnValue = GetStringValue(srcArray[Config.Src.ByColumn])
					var srcKeyColumnValue = GetStringValue(srcArray[keyColumn])
					var desKeyColumnValue = GetStringValue(desArray[Column])

					if !okSrc || !okDes {
						Logger.Warn("数据比对，发现des数据错误, 发送至update回调", zap.Int("srcId", int(srcId)),
							zap.String("srcTable", Config.Src.Table),
							zap.String("desTable", Config.Des[i].Table),
							zap.String("desByColumn", Config.Des[i].ByColumn),
							zap.String("desByColumnValue", configSrcByColumnValue),
							zap.String("srcKeyColumn", keyColumn),
							zap.String("srcKeyColumnValue", srcKeyColumnValue),
							zap.String("desKeyColumn", Column),
							zap.String("desKeyColumnValue", desKeyColumnValue),
						)
						CallbackJobQueue <- &CallbackJob{Types: UPDATE, SrcArray: srcArray, CallbackUrl: Config.Des[i].CallbackNotification.Url}
					} else if srcArray[keyColumn] != desArray[Column] {
						Logger.Warn("数据比对，发现des数据错误, 发送至update回调", zap.Int("srcId", int(srcId)),
							zap.String("srcTable", Config.Src.Table),
							zap.String("desTable", Config.Des[i].Table),
							zap.String("desByColumn", Config.Des[i].ByColumn),
							zap.String("desByColumnValue", configSrcByColumnValue),
							zap.String("srcKeyColumn", keyColumn),
							zap.String("srcKeyColumnValue", srcKeyColumnValue),
							zap.String("desKeyColumn", Column),
							zap.String("desKeyColumnValue", desKeyColumnValue),
						)
						CallbackJobQueue <- &CallbackJob{Types: UPDATE, SrcArray: srcArray, CallbackUrl: Config.Des[i].CallbackNotification.Url}
					}

				}
			}
		}
	}
}

func query(db *sql.DB, sql string) map[string]interface{} {

	rows, err := db.Query(sql)
	defer rows.Close()
	CheckErr(err)
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for j := range values {
		scanArgs[j] = &values[j]
	}

	record := make(map[string]interface{})
	for rows.Next() {
		//将行数据保存到record字典
		err = rows.Scan(scanArgs...)
		for i, col := range values {
			switch col.(type) {
			case int:
				record[columns[i]] = col.(int)
			case []byte:
				record[columns[i]] = string(col.([]byte))
			case nil:
				record[columns[i]] = nil
			}
		}
	}

	return record
}

func GetStringValue(val interface{}) string {

	switch val.(type) {
	case string:
		return val.(string)
	case nil:
		return "null"
	default:
		return "null"
	}
}



func CheckErr(err error) {
	if err != nil {
		Logger.Warn("程序error", zap.String("error", err.Error()))
		panic(err)
	}
}
