package mySql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"time"
	// "github.com/gogf/gf/v2/container/gset"
	// "encoding/json"
	// "crypto/x509/pkix"
)
var db *sql.DB



type testCA struct{
	SerialNumber string
	AKI string
	URI string
	// ID int
	// Adding_time time.Time
	IsValid bool
}

type testDB struct{
	CAInfos testCA
	// URI string
	ID int
	// PreID int
	Adding_time time.Time
	Expired_time time.Time

}

// func main(){
	
// 	/////测试1
// 	// var a = testCA{
// 	// 	SerialNumber : "2dfsddfds2",
// 	// 	AKI : "5112",
// 	// 	URI : "a.txt",
// 	// 	IsValid: true,
// 	// }
// 	// var t = time.Now()
	
// 	// // var b = testDB{
// 	// // 	CAInfos: a
// 	// // }
// 	// db, err := sql.Open("mysql", "root:123456@tcp(localhost:3306)/ca?parseTime=true") 
// 	// defer db.Close()
// 	// if err!=nil{
// 	// 	fmt.Println("DB建立失败!")
// 	// 	return
// 	// }
// 	// db.SetMaxOpenConns(1) //设置最大并发连接数
//     // db.SetMaxIdleConns(1)  //设置闲置连接数
// 	// //golang里未实现批量处理

// 	// //检查数据库连接是否可用
// 	// err=db.Ping() 
// 	// if err!=nil{
// 	// 	fmt.Println("DB不可使用")
// 	// 	return
// 	// }
// 	// t := time.Now()
// 	// _,err = db.Exec(`update testca set expired_time = ? where id=? and expired_time>?`,t,8,t)
// 	// if err != nil {
// 	// 	// tx_rollback(err,tx)
// 	// 	fmt.Println(err.Error())
// 	// 	// return err
// 	// }

	
// 	// err = InsertTest(db,a,t)
// 	// if err!=nil{
// 	// 	fmt.Println("插入测试失败")
// 	// 	fmt.Println(err)
// 	// }
	
// 	////////set
// 	// s:=gset.New(true)
// 	// s.Add("a.txt")
// 	// s.Add("b.txt")
// 	// file := "c"
// 	// fmt.Println(s.Contains(file))
// 	// fmt.Println(s.Contains("a.txt"))

// 	/////////////////////json
// 	// s:=`{"RevokedCertificateList":[{"SerialNumber":2,"RevocationTime":"2021-07-30T14:35:08Z","Extensions":null},{"SerialNumber":327,"RevocationTime":"2022-05-11T01:28:47Z","Extensions":null},{"SerialNumber":328,"RevocationTime":"2022-05-12T01:27:01Z","Extensions":null},{"SerialNumber":329,"RevocationTime":"2022-05-13T01:26:11Z","Extensions":null}]}`
// 	// ss := []byte(s)
// 	// var a RevokedCertificates
// 	// json.Unmarshal(ss,&a)
// 	// for _,i:=range a.RevokedCertificateList{
// 	// 	fmt.Println(i.SerialNumber)
// 	// 	fmt.Println(i.RevocationTime)
// 	// }

// 	////////
// 	// test()
// }



// func InsertOneROA()(){

// }

// func InsertOneManifest()(){

// }

// func InsertOneCRL()(){
	
// }

func tx_rollback(err error, tx *sql.Tx) {
    if err != nil {
        err = tx.Rollback()
        if err != nil {
            fmt.Println("tx.Rollback() Error:" + err.Error())
            return
        }
    }
}





func test()(error){
	var n int
	// var oneCA testDB
	//创建DB
	db, err := sql.Open("mysql", "root:123456@tcp(localhost:3306)/ca?parseTime=true") 
	if err!=nil{
		fmt.Println("DB建立失败!")
		return err
	}
	db.SetMaxOpenConns(1) //设置最大并发连接数
    db.SetMaxIdleConns(1)  //设置闲置连接数
	//golang里未实现批量处理

	//检查数据库连接是否可用
	err=db.Ping() 
	if err!=nil{
		fmt.Println("DB不可使用")
		return err
	}
	//使用Query进行查询，返回查询结果和err
	rows,_:=db.Query("select * from testCA")

	for rows.Next(){ //Next()按行进行遍历
		if err = rows.Scan(&n); err != nil { //Scan取出结果，当一行数据较为复杂时可通过构造struct来取出结果
            fmt.Println(err)    
        }
		fmt.Println(n)
	}
	//检查是否正常遍历结束
	if rows.Err() != nil {
        fmt.Println(err)
    }

	rows.Close() //在Close()之前使用的连接不会放回连接池，不Close()会占用连接资源
	
	db.Close() //关闭DB。DB是用来长时间连接的，不要频繁打开关闭
	return nil
}



func InsertTest(db *sql.DB,CAInfo testCA,adding_time time.Time)(error){
	defaultTime, _:= time.Parse("2006-01-02 15:04:05", "9999-01-01 00:00:00")
	tx, err := db.Begin()
    if err != nil {
		return err
    }
	CAList := make([]testDB ,0)
	var oneCA testDB
	//根据插入证书的URI搜索表
	rows,err := tx.Query("Select * from testCA where uri=? order by ID DESC",CAInfo.URI) //按照ID降序，便于还原最大ID的条目
	if err!=nil{
		fmt.Println("INSERT CA: GET CA ERROR!") //后面改成log
		return err
	}
	for rows.Next(){
		if err=rows.Scan(&oneCA.CAInfos.SerialNumber,&oneCA.CAInfos.AKI,&oneCA.CAInfos.URI,&oneCA.ID,&oneCA.Adding_time,&oneCA.Expired_time,&oneCA.CAInfos.IsValid);err!=nil{
			fmt.Println("INSERT CA: SCAN ROWS ERROR!") 
			return err
		}
		CAList = append(CAList,oneCA)
	}

	if rows.Err() != nil {
        fmt.Println("INSERT CA: ERROR EXIT SCANNING ROWS") 
		return err
    }

	//if nil  插入
	if len(CAList)==0{
		rs,err := tx.Exec(`INSERT INTO testCA(SerialNumber,AKI,URI,adding_time,expired_time,isValid) VALUES (?,?,?,?,?,?)`,CAInfo.SerialNumber,CAInfo.AKI,CAInfo.URI,adding_time,defaultTime,CAInfo.IsValid)
		if err != nil {
			tx_rollback(err,tx)
			fmt.Println(err.Error())
			return err
		}
		rowAffected, err := rs.RowsAffected()
   		if err != nil {
			tx_rollback(err,tx)
        	fmt.Println(err.Error())
			return err
    	}
		fmt.Println(rowAffected)
	}else{ //if not nil 还原最大ID条目 逐条目比较
		//还原最大ID条目
		var MaxIdItem testDB
		for _,CA := range CAList{
			if MaxIdItem.CAInfos.SerialNumber == "" && CA.CAInfos.SerialNumber!="null"{
				MaxIdItem.CAInfos.SerialNumber=CA.CAInfos.SerialNumber
			}
			if MaxIdItem.CAInfos.AKI == "" && CA.CAInfos.AKI!="null"{
				MaxIdItem.CAInfos.AKI=CA.CAInfos.AKI
			}
			if MaxIdItem.CAInfos.URI == "" && CA.CAInfos.URI!="null"{
				MaxIdItem.CAInfos.URI=CA.CAInfos.URI
			}
			if MaxIdItem.ID == 0 && CA.ID!= -1{
				MaxIdItem.ID=CA.ID
			}
			//下面每个字段操作同上(除了ID等等)
		}
		MaxIdItem.Adding_time =CAList[0].Adding_time //adding_time
		MaxIdItem.Expired_time =CAList[0].Expired_time //expired_time
		MaxIdItem.CAInfos.IsValid = CAList[0].CAInfos.IsValid // isValid  这三个字段不增量存储

		//逐条目比较{无差异，无操作；有差异，插入一条，更新上一条的失效时间为adding_time}
		if CAInfo == MaxIdItem.CAInfos{
			//do nothing
		}else{
			var tempCA testDB
			if CAInfo.SerialNumber != MaxIdItem.CAInfos.SerialNumber{ //先这样遍历，后面学习怎么遍历结构体
				tempCA.CAInfos.SerialNumber = CAInfo.SerialNumber
			}else{
				tempCA.CAInfos.SerialNumber="null"
			}
			if CAInfo.AKI != MaxIdItem.CAInfos.AKI{ 
				tempCA.CAInfos.AKI = CAInfo.AKI
			}else{
				tempCA.CAInfos.AKI="null"
			}
			tempCA.CAInfos.URI = CAInfo.URI
			tempCA.CAInfos.IsValid = CAInfo.IsValid
			
			rs,err := tx.Exec(`INSERT INTO testCA(SerialNumber,AKI,URI,adding_time,expired_time,isValid) VALUES (?,?,?,?,?,?)`,tempCA.CAInfos.SerialNumber,tempCA.CAInfos.AKI,tempCA.CAInfos.URI,adding_time,defaultTime,tempCA.CAInfos.IsValid)
			if err != nil {
				tx_rollback(err,tx)
				fmt.Println(err.Error())
				return err
			}
			rowAffected, err := rs.RowsAffected()
   			if err != nil {
				tx_rollback(err,tx)
        		fmt.Println(err.Error())
				return err
    		}
			fmt.Println(rowAffected)

			// tempCA.PreID = MaxIdItem.ID
			
			//对剩下每个字段进行上述操作 （  :( , 真长 ）
			//插入   考虑直接使用db.Exec

			//更新preID的expired_time的值 条件更新
			rs,err = tx.Exec(`update testCA set expired_time = ? where URI=? and id =?`,adding_time,MaxIdItem.CAInfos.URI,MaxIdItem.ID)
			if err != nil {
				tx_rollback(err,tx)
				fmt.Println(err.Error())
				return err
			}
			rowAffected, err = rs.RowsAffected()
   			if err != nil {
				tx_rollback(err,tx)
        		fmt.Println(err.Error())
				return err
    		}
			fmt.Println(rowAffected)
		}

	}
	err = tx.Commit()
	tx_rollback(err,tx)
	return nil
}

