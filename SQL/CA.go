package mySql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"time"
	// "github.com/gogf/gf/v2/container/gset"
	// "encoding/json"
	// "crypto/x509/pkix"
	// "math/big"
)


type CAFileds struct{
	SerialNumber string
	SKI string 
	AKI string 
	NotBefore time.Time
	NotAfter time.Time
	SubjectPublicKeyInfo string
	IPResources string
	ASResources string
	AIA string
	SIAMft string
	SIACRL string
	URI string
	// ID int
	// PreID int
	// Adding_time time.Time
	// Expired_time time.Time
	IsValid bool
}

type CAInDataBase struct{
	Fileds CAFileds
	ID int
	PreID int
	Adding_time time.Time
	Expired_time time.Time
	// IsValid bool
}


func InsertOneCA(db *sql.DB,CAInfo CAFileds,adding_time time.Time)(error){
	defaultTime, _:= time.Parse("2006-01-02 15:04:05", "9999-01-01 00:00:00")
	tx, err := db.Begin()
    if err != nil {
		return err
    }
	CAList := make([]CAInDataBase ,0)
	var oneCA CAInDataBase
	//根据插入证书的URI搜索表
	rows,err := tx.Query("Select * from CAs where uri=? order by ID DESC",CAInfo.URI) //按照ID降序，便于还原最大ID的条目
	if err!=nil{
		fmt.Println("INSERT CA: GET CA ERROR!") //后面改成log
		return err
	}
	for rows.Next(){
		if err=rows.Scan(&oneCA.Fileds.SerialNumber,&oneCA.Fileds.SKI,&oneCA.Fileds.AKI,&oneCA.Fileds.NotBefore,&oneCA.Fileds.NotAfter,&oneCA.Fileds.SubjectPublicKeyInfo,&oneCA.Fileds.IPResources,&oneCA.Fileds.ASResources,&oneCA.Fileds.AIA,&oneCA.Fileds.SIAMft,&oneCA.Fileds.SIACRL,&oneCA.Fileds.URI,&oneCA.ID,&oneCA.PreID,&oneCA.Adding_time,&oneCA.Expired_time,&oneCA.Fileds.IsValid);err!=nil{
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
		rs,err := tx.Exec("INSERT INTO CAs VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",CAInfo.SerialNumber,CAInfo.SKI,CAInfo.AKI,CAInfo.NotBefore,CAInfo.NotAfter,CAInfo.SubjectPublicKeyInfo,CAInfo.IPResources,CAInfo.ASResources,CAInfo.AIA,CAInfo.SIAMft,CAInfo.SIACRL,CAInfo.URI,0,-1,adding_time,defaultTime,CAInfo.IsValid)
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
		var MaxIdItem CAInDataBase
		var tempCA CAFileds
		for _,CA := range CAList{
			if MaxIdItem.Fileds.SerialNumber == "" && CA.Fileds.SerialNumber!="null"{
				MaxIdItem.Fileds.SerialNumber=CA.Fileds.SerialNumber
			}
			if MaxIdItem.Fileds.SKI == "" && CA.Fileds.SKI!="null"{
				MaxIdItem.Fileds.SKI=CA.Fileds.SKI
			}
			if MaxIdItem.Fileds.AKI == "" && CA.Fileds.AKI!="null"{
				MaxIdItem.Fileds.AKI=CA.Fileds.AKI
			}
			if MaxIdItem.Fileds.SubjectPublicKeyInfo == "" && CA.Fileds.SubjectPublicKeyInfo!="null"{
				MaxIdItem.Fileds.SubjectPublicKeyInfo=CA.Fileds.SubjectPublicKeyInfo
			}
			if MaxIdItem.Fileds.IPResources == "" && CA.Fileds.IPResources!="null"{
				MaxIdItem.Fileds.IPResources=CA.Fileds.IPResources
			}
			if MaxIdItem.Fileds.ASResources == "" && CA.Fileds.ASResources!="null"{
				MaxIdItem.Fileds.ASResources=CA.Fileds.ASResources
			}
			if MaxIdItem.Fileds.AIA == "" && CA.Fileds.AIA!="null"{
				MaxIdItem.Fileds.AIA=CA.Fileds.AIA
			}
			if MaxIdItem.Fileds.SIAMft == "" && CA.Fileds.SIAMft!="null"{
				MaxIdItem.Fileds.SIAMft=CA.Fileds.SIAMft
			}
			if MaxIdItem.Fileds.SIACRL == "" && CA.Fileds.SIACRL!="null"{
				MaxIdItem.Fileds.SIACRL=CA.Fileds.SIACRL
			}
			//下面每个字段操作同上(除了URI,ID,preID,adding_time,expired_time,IsValid和时间这几个字段不进行增量存储)
		}
		MaxIdItem.Fileds.URI = CAList[0].Fileds.URI
		MaxIdItem.Fileds.NotBefore = CAList[0].Fileds.NotBefore
		MaxIdItem.Fileds.NotAfter = CAList[0].Fileds.NotAfter
		MaxIdItem.Fileds.IsValid =CAList[0].Fileds.IsValid
		MaxIdItem.ID = CAList[0].ID
		MaxIdItem.PreID = CAList[0].PreID
		MaxIdItem.Adding_time = CAList[0].Adding_time
		MaxIdItem.Expired_time =CAList[0].Expired_time

		//逐条目比较{无差异，无操作；有差异，插入一条，更新上一条的失效时间为adding_time}
		if CAInfo == MaxIdItem.Fileds{
			//do nothing
		}else{
			if CAInfo.SerialNumber != MaxIdItem.Fileds.SerialNumber{ //先这样遍历，后面学习怎么遍历结构体
				tempCA.SerialNumber = CAInfo.SerialNumber
			}else{
				tempCA.SerialNumber="null"
			}
			if CAInfo.SKI != MaxIdItem.Fileds.SKI{ //先这样遍历，后面学习怎么遍历结构体
				tempCA.SKI = CAInfo.SKI
			}else{
				tempCA.SKI="null"
			}
			if CAInfo.AKI != MaxIdItem.Fileds.AKI{ //先这样遍历，后面学习怎么遍历结构体
				tempCA.AKI = CAInfo.AKI
			}else{
				tempCA.AKI="null"
			}
			if CAInfo.SubjectPublicKeyInfo != MaxIdItem.Fileds.SubjectPublicKeyInfo{ //先这样遍历，后面学习怎么遍历结构体
				tempCA.SubjectPublicKeyInfo = CAInfo.SubjectPublicKeyInfo
			}else{
				tempCA.SubjectPublicKeyInfo="null"
			}
			if CAInfo.IPResources != MaxIdItem.Fileds.IPResources{ //先这样遍历，后面学习怎么遍历结构体
				tempCA.IPResources = CAInfo.IPResources
			}else{
				tempCA.IPResources="null"
			}
			if CAInfo.ASResources != MaxIdItem.Fileds.ASResources{ //先这样遍历，后面学习怎么遍历结构体
				tempCA.ASResources = CAInfo.ASResources
			}else{
				tempCA.ASResources="null"
			}
			if CAInfo.AIA != MaxIdItem.Fileds.AIA{ //先这样遍历，后面学习怎么遍历结构体
				tempCA.AIA = CAInfo.AIA
			}else{
				tempCA.AIA="null"
			}
			if CAInfo.SIAMft != MaxIdItem.Fileds.SIAMft{ //先这样遍历，后面学习怎么遍历结构体
				tempCA.SIAMft = CAInfo.SIAMft
			}else{
				tempCA.SIAMft="null"
			}
			if CAInfo.SIACRL != MaxIdItem.Fileds.SIACRL{ //先这样遍历，后面学习怎么遍历结构体
				tempCA.SIACRL = CAInfo.SIACRL
			}else{
				tempCA.SIACRL="null"
			}
			tempCA.URI = CAInfo.URI
			tempCA.NotBefore = CAInfo.NotBefore
			tempCA.NotAfter =CAInfo.NotAfter
			tempCA.IsValid = CAInfo.IsValid
			//对剩下每个字段进行上述操作 （  :( , 真长 ）
			//插入   注意 preID此时不为默认值
			rs,err := tx.Exec("INSERT INTO CAs VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",tempCA.SerialNumber,tempCA.SKI,tempCA.AKI,tempCA.NotBefore,tempCA.NotAfter,tempCA.SubjectPublicKeyInfo,tempCA.IPResources,tempCA.ASResources,tempCA.AIA,tempCA.SIAMft,tempCA.SIACRL,tempCA.URI,0,MaxIdItem.ID,adding_time,defaultTime,tempCA.IsValid)
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

			//更新preID的expired_time的值 条件更新
			rs,err = tx.Exec(`update CAs set expired_time = ? where URI=? and id =? and expired_time>?`,adding_time,MaxIdItem.Fileds.URI,MaxIdItem.ID,adding_time)
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