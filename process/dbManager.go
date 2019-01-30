package process

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
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
	srcDb = db

	lens := len(Config.Des)
	desDb = make([]*sql.DB, lens)
	for i := 0; i < lens; i++ {
		db, err = sql.Open("mysql", Config.Des[i].Dsn)
		if err != nil {
			panic(err)
		}
		db.SetMaxIdleConns(maxIdleConns)
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
		fmt.Println(maxId)
		CheckErr(err)
	}

	return maxId
}

func CompareColumn(srcId uint) {


	srcArray := query(srcDb, "SELECT * FROM "+Config.Src.Table+" where id= "+strconv.Itoa(int(srcId)))
	if len(srcArray) != 0 {
		lens := len(Config.Des)
		for i := 0; i < lens; i++ {
			desArray := query(desDb[i], "SELECT * FROM "+Config.Des[i].Table+" where "+Config.Des[i].ByColumn+"= "+ srcArray[Config.Src.ByColumn].(string))
			if len(desArray) == 0 {
				fmt.Println(srcId, "缺少数据:", srcArray)
			} else {
				for keyColumn, Column := range Config.Des[i].Columns {
					_, okSrc := srcArray[keyColumn]
					_, okDes := desArray[Column]
					if !okSrc || !okDes {
						fmt.Println(srcId,"数据对不上:", srcArray[keyColumn], desArray[Column])
					}else if srcArray[keyColumn] != desArray[Column] {
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
		panic(err)
		fmt.Println("err:", err)
	}
}