package mySql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"time"
	// "github.com/gogf/gf/v2/container/gset"
)


type ROAFileds struct{
	AsID int
	IpAddrBlocks string
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

type ROAInDataBase struct{
	Fileds ROAFileds
	ID int
	PreID int
	Adding_time time.Time
	Expired_time time.Time
	// IsValid bool
}





func InsertOneROA(db *sql.DB,ROAInfo ROAFileds,adding_time time.Time)(error){
	defaultTime, _:= time.Parse("2006-01-02 15:04:05", "9999-01-01 00:00:00")
	tx, err := db.Begin()
    if err != nil {
		return err
    }
	ROAList := make([]ROAInDataBase ,0)
	var oneROA ROAInDataBase
	//根据插入证书的URI搜索表
	rows,err := tx.Query("Select * from ROAs where uri=? order by ID DESC",ROAInfo.URI) //按照ID降序，便于还原最大ID的条目
	if err!=nil{
		fmt.Println("INSERT ROA: GET CA ERROR!") //后面改成log
		return err
	}
	for rows.Next(){
		if err=rows.Scan(&oneROA.Fileds.AsID,&oneROA.Fileds.IpAddrBlocks,&oneROA.Fileds.SerialNumber,&oneROA.Fileds.SKI,&oneROA.Fileds.AKI,&oneROA.Fileds.NotBefore,&oneROA.Fileds.NotAfter,&oneROA.Fileds.SubjectPublicKeyInfo,&oneROA.Fileds.IPResources,&oneROA.Fileds.ASResources,&oneROA.Fileds.AIA,&oneROA.Fileds.SIAMft,&oneROA.Fileds.SIACRL,&oneROA.Fileds.URI,&oneROA.ID,&oneROA.PreID,&oneROA.Adding_time,&oneROA.Expired_time,&oneROA.Fileds.IsValid);err!=nil{
			fmt.Println("INSERT ROA: SCAN ROWS ERROR!") 
			return err
		}
		ROAList = append(ROAList,oneROA)
	}

	if rows.Err() != nil {
        fmt.Println("INSERT ROA: ERROR EXIT SCANNING ROWS") 
		return err
    }

	//if nil  插入
	if len(ROAList)==0{
		rs,err := tx.Exec("INSERT INTO ROAs VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",ROAInfo.AsID,ROAInfo.IpAddrBlocks,ROAInfo.SerialNumber,ROAInfo.SKI,ROAInfo.AKI,ROAInfo.NotBefore,ROAInfo.NotAfter,ROAInfo.SubjectPublicKeyInfo,ROAInfo.IPResources,ROAInfo.ASResources,ROAInfo.AIA,ROAInfo.SIAMft,ROAInfo.SIACRL,ROAInfo.URI,0,-1,adding_time,defaultTime,ROAInfo.IsValid)
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
		var MaxIdItem ROAInDataBase
		var tempROA ROAFileds
		for _,ROA := range ROAList{
			if MaxIdItem.Fileds.AsID == 0 && ROA.Fileds.AsID!=-1{
				MaxIdItem.Fileds.AsID=ROA.Fileds.AsID
			}
			if MaxIdItem.Fileds.IpAddrBlocks == "" && ROA.Fileds.IpAddrBlocks!="null"{
				MaxIdItem.Fileds.IpAddrBlocks=ROA.Fileds.IpAddrBlocks
			}
			if MaxIdItem.Fileds.SerialNumber == "" && ROA.Fileds.SerialNumber!="null"{
				MaxIdItem.Fileds.SerialNumber=ROA.Fileds.SerialNumber
			}
			if MaxIdItem.Fileds.SKI == "" && ROA.Fileds.SKI!="null"{
				MaxIdItem.Fileds.SKI=ROA.Fileds.SKI
			}
			if MaxIdItem.Fileds.AKI == "" && ROA.Fileds.AKI!="null"{
				MaxIdItem.Fileds.AKI=ROA.Fileds.AKI
			}
			if MaxIdItem.Fileds.SubjectPublicKeyInfo == "" && ROA.Fileds.SubjectPublicKeyInfo!="null"{
				MaxIdItem.Fileds.SubjectPublicKeyInfo=ROA.Fileds.SubjectPublicKeyInfo
			}
			if MaxIdItem.Fileds.IPResources == "" && ROA.Fileds.IPResources!="null"{
				MaxIdItem.Fileds.IPResources=ROA.Fileds.IPResources
			}
			if MaxIdItem.Fileds.ASResources == "" && ROA.Fileds.ASResources!="null"{
				MaxIdItem.Fileds.ASResources=ROA.Fileds.ASResources
			}
			if MaxIdItem.Fileds.AIA == "" && ROA.Fileds.AIA!="null"{
				MaxIdItem.Fileds.AIA=ROA.Fileds.AIA
			}
			if MaxIdItem.Fileds.SIAMft == "" && ROA.Fileds.SIAMft!="null"{
				MaxIdItem.Fileds.SIAMft=ROA.Fileds.SIAMft
			}
			if MaxIdItem.Fileds.SIACRL == "" && ROA.Fileds.SIACRL!="null"{
				MaxIdItem.Fileds.SIACRL=ROA.Fileds.SIACRL
			}
			//下面每个字段操作同上(除了ID,preID,adding_time,expired_time,IsValid,这五个字段不进行增量存储)

		}
		MaxIdItem.Fileds.URI = ROAList[0].Fileds.URI
		MaxIdItem.Fileds.NotBefore = ROAList[0].Fileds.NotBefore
		MaxIdItem.Fileds.NotAfter = ROAList[0].Fileds.NotAfter
		MaxIdItem.Fileds.IsValid =ROAList[0].Fileds.IsValid
		MaxIdItem.ID = ROAList[0].ID
		MaxIdItem.PreID = ROAList[0].PreID
		MaxIdItem.Adding_time = ROAList[0].Adding_time
		MaxIdItem.Expired_time =ROAList[0].Expired_time

		//逐条目比较{无差异，无操作；有差异，插入一条，更新上一条的失效时间为adding_time}
		if ROAInfo == MaxIdItem.Fileds{
			//do nothing
		}else{
			if ROAInfo.AsID != MaxIdItem.Fileds.AsID{ //先这样遍历，后面学习怎么遍历结构体
				tempROA.AsID = ROAInfo.AsID
			}else{
				tempROA.AsID=-1
			}
			if ROAInfo.IpAddrBlocks != MaxIdItem.Fileds.IpAddrBlocks{ //先这样遍历，后面学习怎么遍历结构体
				tempROA.IpAddrBlocks = ROAInfo.IpAddrBlocks
			}else{
				tempROA.IpAddrBlocks="null"
			}
			if ROAInfo.SerialNumber != MaxIdItem.Fileds.SerialNumber{ //先这样遍历，后面学习怎么遍历结构体
				tempROA.SerialNumber = ROAInfo.SerialNumber
			}else{
				tempROA.SerialNumber="null"
			}
			if ROAInfo.SKI != MaxIdItem.Fileds.SKI{ //先这样遍历，后面学习怎么遍历结构体
				tempROA.SKI = ROAInfo.SKI
			}else{
				tempROA.SKI="null"
			}
			if ROAInfo.AKI != MaxIdItem.Fileds.AKI{ //先这样遍历，后面学习怎么遍历结构体
				tempROA.AKI = ROAInfo.AKI
			}else{
				tempROA.AKI="null"
			}
			if ROAInfo.SubjectPublicKeyInfo != MaxIdItem.Fileds.SubjectPublicKeyInfo{ //先这样遍历，后面学习怎么遍历结构体
				tempROA.SubjectPublicKeyInfo = ROAInfo.SubjectPublicKeyInfo
			}else{
				tempROA.SubjectPublicKeyInfo="null"
			}
			if ROAInfo.IPResources != MaxIdItem.Fileds.IPResources{ //先这样遍历，后面学习怎么遍历结构体
				tempROA.IPResources = ROAInfo.IPResources
			}else{
				tempROA.IPResources="null"
			}
			if ROAInfo.ASResources != MaxIdItem.Fileds.ASResources{ //先这样遍历，后面学习怎么遍历结构体
				tempROA.ASResources = ROAInfo.ASResources
			}else{
				tempROA.ASResources="null"
			}
			if ROAInfo.AIA != MaxIdItem.Fileds.AIA{ //先这样遍历，后面学习怎么遍历结构体
				tempROA.AIA = ROAInfo.AIA
			}else{
				tempROA.AIA="null"
			}
			if ROAInfo.SIAMft != MaxIdItem.Fileds.SIAMft{ //先这样遍历，后面学习怎么遍历结构体
				tempROA.SIAMft = ROAInfo.SIAMft
			}else{
				tempROA.SIAMft="null"
			}
			if ROAInfo.SIACRL != MaxIdItem.Fileds.SIACRL{ //先这样遍历，后面学习怎么遍历结构体
				tempROA.SIACRL = ROAInfo.SIACRL
			}else{
				tempROA.SIACRL="null"
			}
			tempROA.URI = ROAInfo.URI
			tempROA.NotBefore = ROAInfo.NotBefore
			tempROA.NotAfter =ROAInfo.NotAfter
			tempROA.IsValid = ROAInfo.IsValid
			
			//对剩下每个字段进行上述操作 （  :( , 真长 ）
			//插入   注意 preID此时不为默认值
			rs,err := tx.Exec("INSERT INTO ROAs VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",tempROA.AsID,tempROA.IpAddrBlocks,tempROA.SerialNumber,tempROA.SKI,tempROA.AKI,tempROA.NotBefore,tempROA.NotAfter,tempROA.SubjectPublicKeyInfo,tempROA.IPResources,tempROA.ASResources,tempROA.AIA,tempROA.SIAMft,tempROA.SIACRL,tempROA.URI,0,MaxIdItem.ID,adding_time,defaultTime,tempROA.IsValid)
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
			rs,err = tx.Exec(`update ROAs set expired_time = ? where URI=? and id =? and expired_time>?`,adding_time,MaxIdItem.Fileds.URI,MaxIdItem.ID,adding_time)
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