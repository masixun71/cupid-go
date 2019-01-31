package process

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"../jobQueue"
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

	rows, err := srcDb.Query("SELECT max(id) as maxId from " + Config.Src.Table)
	defer rows.Close()
	CheckErr(err)
	var maxId int
	for rows.Next() {
		err = rows.Scan(&maxId)
		fmt.Println("maxId", maxId)
		CheckErr(err)
	}

	return maxId
}

func GetUpdateId() []int {

	end := time.Now()
	d, _ := time.ParseDuration("-"+ strconv.Itoa(Config.Src.UpdateScanSecond) + "s")
	start := end.Add(d)
	updateSql := "SELECT id from " + Config.Src.Table + " where " + Config.Src.UpdateColumn + " between '" + start.Format(Config.Src.UpdateTimeFormate) + "' and '" + end.Format(Config.Src.UpdateTimeFormate) + "'"
	rows, err := srcDb.Query(updateSql)
	defer rows.Close()
	CheckErr(err)
	var updateId int
	ids := make([]int, 0)

	for rows.Next() {
		err = rows.Scan(&updateId)
		fmt.Println("updateId", updateId)
		CheckErr(err)
		ids = append(ids, updateId)
	}

	return ids
}



func CompareColumn(srcId uint) {


	srcArray := query(srcDb, "SELECT * FROM "+Config.Src.Table+" where id= "+strconv.Itoa(int(srcId)))
	if len(srcArray) != 0 {
		lens := len(Config.Des)
		for i := 0; i < lens; i++ {
			desArray := query(desDb[i], "SELECT * FROM "+Config.Des[i].Table+" where "+Config.Des[i].ByColumn+"= "+ srcArray[Config.Src.ByColumn].(string))
			if len(desArray) == 0 {
				fmt.Println(srcId, "缺少数据:", srcArray)
				jobQueue.ProcessJobQueue <- &CallbackJob{types:INSERT, srcArray:srcArray,callbackUrl:Config.Des[i].CallbackNotification.Url}
			} else {
				for keyColumn, Column := range Config.Des[i].Columns {
					_, okSrc := srcArray[keyColumn]
					_, okDes := desArray[Column]
					if !okSrc || !okDes {
						fmt.Println(srcId,"数据对不上:", srcArray[keyColumn], desArray[Column])
						jobQueue.ProcessJobQueue <- &CallbackJob{types:UPDATE, srcArray:srcArray,callbackUrl:Config.Des[i].CallbackNotification.Url}
					}else if srcArray[keyColumn] != desArray[Column] {
						jobQueue.ProcessJobQueue <- &CallbackJob{types:UPDATE, srcArray:srcArray,callbackUrl:Config.Des[i].CallbackNotification.Url}
						fmt.Println(srcId, "数据对不上:", srcArray[keyColumn], desArray[Column])
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
			switch col.(type){
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



func CheckErr(err error) {
	if err != nil {
		fmt.Println("err:", err)
		panic(err)
	}
}
