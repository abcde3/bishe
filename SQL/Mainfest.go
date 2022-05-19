package mySql
import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"time"
	// "math/big"
	// "github.com/gogf/gf/v2/container/gset"
)


type ManifestFileds struct{
	ManifestNum string //??
	FileList string
	ThisUpdate time.Time
	NextUpdate time.Time
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

type ManifestInDataBase struct{
	Fileds ManifestFileds
	ID int
	PreID int
	Adding_time time.Time
	Expired_time time.Time
	// IsValid bool
}





func InsertOneMft(db *sql.DB,MftInfo ManifestFileds,adding_time time.Time)(error){
	defaultTime, _:= time.Parse("2006-01-02 15:04:05", "9999-01-01 00:00:00")
	tx, err := db.Begin()
    if err != nil {
		return err
    }
	MftList := make([]ManifestInDataBase ,0)
	var oneMft ManifestInDataBase
	//根据插入证书的URI搜索表
	rows,err := tx.Query("Select * from Mfts where uri=? order by ID DESC",MftInfo.URI) //按照ID降序，便于还原最大ID的条目
	if err!=nil{
		fmt.Println("INSERT Mft: GET CA ERROR!") //后面改成log
		return err
	}
	for rows.Next(){
		if err=rows.Scan(&oneMft.Fileds.ManifestNum,&oneMft.Fileds.FileList,&oneMft.Fileds.ThisUpdate,&oneMft.Fileds.NextUpdate,&oneMft.Fileds.SerialNumber,&oneMft.Fileds.SKI,&oneMft.Fileds.AKI,&oneMft.Fileds.NotBefore,&oneMft.Fileds.NotAfter,&oneMft.Fileds.SubjectPublicKeyInfo,&oneMft.Fileds.IPResources,&oneMft.Fileds.ASResources,&oneMft.Fileds.AIA,&oneMft.Fileds.SIAMft,&oneMft.Fileds.SIACRL,&oneMft.Fileds.URI,&oneMft.ID,&oneMft.PreID,&oneMft.Adding_time,&oneMft.Expired_time,&oneMft.Fileds.IsValid);err!=nil{
			fmt.Println("INSERT Mft: SCAN ROWS ERROR!") 
			return err
		}
		MftList = append(MftList,oneMft)
	}

	if rows.Err() != nil {
        fmt.Println("INSERT Mft: ERROR EXIT SCANNING ROWS") 
		return err
    }

	//if nil  插入
	if len(MftList)==0{
		rs,err := tx.Exec("INSERT INTO Mfts VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",MftInfo.ManifestNum,MftInfo.FileList,MftInfo.ThisUpdate,MftInfo.NextUpdate,MftInfo.SerialNumber,MftInfo.SKI,MftInfo.AKI,MftInfo.NotBefore,MftInfo.NotAfter,MftInfo.SubjectPublicKeyInfo,MftInfo.IPResources,MftInfo.ASResources,MftInfo.AIA,MftInfo.SIAMft,MftInfo.SIACRL,MftInfo.URI,0,-1,adding_time,defaultTime,MftInfo.IsValid)
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
		var MaxIdItem ManifestInDataBase
		var tempMft ManifestFileds
		for _,Mft := range MftList{
			if MaxIdItem.Fileds.ManifestNum == "" && Mft.Fileds.ManifestNum!="null"{
				MaxIdItem.Fileds.ManifestNum=Mft.Fileds.ManifestNum
			}
			if MaxIdItem.Fileds.FileList == "" && Mft.Fileds.FileList!="null"{
				MaxIdItem.Fileds.FileList=Mft.Fileds.FileList
			}
			if MaxIdItem.Fileds.SerialNumber == "" && Mft.Fileds.SerialNumber!="null"{
				MaxIdItem.Fileds.SerialNumber=Mft.Fileds.SerialNumber
			}
			if MaxIdItem.Fileds.SKI == "" && Mft.Fileds.SKI!="null"{
				MaxIdItem.Fileds.SKI=Mft.Fileds.SKI
			}
			if MaxIdItem.Fileds.AKI == "" && Mft.Fileds.AKI!="null"{
				MaxIdItem.Fileds.AKI=Mft.Fileds.AKI
			}
			if MaxIdItem.Fileds.SubjectPublicKeyInfo == "" && Mft.Fileds.SubjectPublicKeyInfo!="null"{
				MaxIdItem.Fileds.SubjectPublicKeyInfo=Mft.Fileds.SubjectPublicKeyInfo
			}
			if MaxIdItem.Fileds.IPResources == "" && Mft.Fileds.IPResources!="null"{
				MaxIdItem.Fileds.IPResources=Mft.Fileds.IPResources
			}
			if MaxIdItem.Fileds.ASResources == "" && Mft.Fileds.ASResources!="null"{
				MaxIdItem.Fileds.ASResources=Mft.Fileds.ASResources
			}
			if MaxIdItem.Fileds.AIA == "" && Mft.Fileds.AIA!="null"{
				MaxIdItem.Fileds.AIA=Mft.Fileds.AIA
			}
			if MaxIdItem.Fileds.SIAMft == "" && Mft.Fileds.SIAMft!="null"{
				MaxIdItem.Fileds.SIAMft=Mft.Fileds.SIAMft
			}
			if MaxIdItem.Fileds.SIACRL == "" && Mft.Fileds.SIACRL!="null"{
				MaxIdItem.Fileds.SIACRL=Mft.Fileds.SIACRL
			}
			//下面每个字段操作同上(除了ID,preID,adding_time,expired_time,IsValid,这五个字段不进行增量存储)
			MaxIdItem.Fileds.ThisUpdate = MftList[0].Fileds.ThisUpdate
			MaxIdItem.Fileds.NextUpdate = MftList[0].Fileds.NextUpdate
			MaxIdItem.Fileds.URI = MftList[0].Fileds.URI
			MaxIdItem.Fileds.NotBefore = MftList[0].Fileds.NotBefore
			MaxIdItem.Fileds.NotAfter = MftList[0].Fileds.NotAfter
			MaxIdItem.Fileds.IsValid =MftList[0].Fileds.IsValid
			MaxIdItem.ID = MftList[0].ID
			MaxIdItem.PreID = MftList[0].PreID
			MaxIdItem.Adding_time = MftList[0].Adding_time
			MaxIdItem.Expired_time =MftList[0].Expired_time

		}

		//逐条目比较{无差异，无操作；有差异，插入一条，更新上一条的失效时间为adding_time}
		if MftInfo == MaxIdItem.Fileds{
			//do nothing
		}else{
			if MftInfo.ManifestNum != MaxIdItem.Fileds.ManifestNum{ //先这样遍历，后面学习怎么遍历结构体
				tempMft.ManifestNum = MftInfo.ManifestNum
			}else{
				tempMft.ManifestNum="null"
			}
			if MftInfo.FileList != MaxIdItem.Fileds.FileList{ //先这样遍历，后面学习怎么遍历结构体
				tempMft.FileList = MftInfo.FileList
			}else{
				tempMft.FileList="null"
			}
			if MftInfo.SerialNumber != MaxIdItem.Fileds.SerialNumber{ //先这样遍历，后面学习怎么遍历结构体
				tempMft.SerialNumber = MftInfo.SerialNumber
			}else{
				tempMft.SerialNumber="null"
			}
			if MftInfo.SKI != MaxIdItem.Fileds.SKI{ //先这样遍历，后面学习怎么遍历结构体
				tempMft.SKI = MftInfo.SKI
			}else{
				tempMft.SKI="null"
			}
			if MftInfo.AKI != MaxIdItem.Fileds.AKI{ //先这样遍历，后面学习怎么遍历结构体
				tempMft.AKI = MftInfo.AKI
			}else{
				tempMft.AKI="null"
			}
			if MftInfo.SubjectPublicKeyInfo != MaxIdItem.Fileds.SubjectPublicKeyInfo{ //先这样遍历，后面学习怎么遍历结构体
				tempMft.SubjectPublicKeyInfo = MftInfo.SubjectPublicKeyInfo
			}else{
				tempMft.SubjectPublicKeyInfo="null"
			}
			if MftInfo.IPResources != MaxIdItem.Fileds.IPResources{ //先这样遍历，后面学习怎么遍历结构体
				tempMft.IPResources = MftInfo.IPResources
			}else{
				tempMft.IPResources="null"
			}
			if MftInfo.ASResources != MaxIdItem.Fileds.ASResources{ //先这样遍历，后面学习怎么遍历结构体
				tempMft.ASResources = MftInfo.ASResources
			}else{
				tempMft.ASResources="null"
			}
			if MftInfo.AIA != MaxIdItem.Fileds.AIA{ //先这样遍历，后面学习怎么遍历结构体
				tempMft.AIA = MftInfo.AIA
			}else{
				tempMft.AIA="null"
			}
			if MftInfo.SIAMft != MaxIdItem.Fileds.SIAMft{ //先这样遍历，后面学习怎么遍历结构体
				tempMft.SIAMft = MftInfo.SIAMft
			}else{
				tempMft.SIAMft="null"
			}
			if MftInfo.SIACRL != MaxIdItem.Fileds.SIACRL{ //先这样遍历，后面学习怎么遍历结构体
				tempMft.SIACRL = MftInfo.SIACRL
			}else{
				tempMft.SIACRL="null"
			}
			tempMft.ThisUpdate = MftInfo.ThisUpdate
			tempMft.NextUpdate = MftInfo.NextUpdate
			tempMft.URI = MftInfo.URI
			tempMft.NotBefore = MftInfo.NotBefore
			tempMft.NotAfter =MftInfo.NotAfter
			tempMft.IsValid = MftInfo.IsValid
			
			//对剩下每个字段进行上述操作 （  :( , 真长 ）
			//插入   注意 preID此时不为默认值
			rs,err := tx.Exec("INSERT INTO Mfts VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",tempMft.ManifestNum,tempMft.FileList,tempMft.ThisUpdate,tempMft.NextUpdate,tempMft.SerialNumber,tempMft.SKI,tempMft.AKI,tempMft.NotBefore,tempMft.NotAfter,tempMft.SubjectPublicKeyInfo,tempMft.IPResources,tempMft.ASResources,tempMft.AIA,tempMft.SIAMft,tempMft.SIACRL,tempMft.URI,0,MaxIdItem.ID,adding_time,defaultTime,tempMft.IsValid)
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
			rs,err = tx.Exec(`update Mfts set expired_time = ? where URI=? and id =?`,adding_time,MaxIdItem.Fileds.URI,MaxIdItem.ID)
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