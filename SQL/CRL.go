package mySql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"time"
	"encoding/json"
	"crypto/x509/pkix"
	"math/big"
	// "github.com/gogf/gf/v2/container/gset"
)

type RevokedCertificates struct{
	RevokedCertificateList []pkix.RevokedCertificate
}

type CRLFileds struct{
	CRLNumber string
	AKI string 
	RevokedCertificateList string
	ThisUpdate time.Time
	NextUpdate time.Time
	URI string
	// ID int
	// PreID int
	// Adding_time time.Time
	// Expired_time time.Time
	IsValid bool
}

type CRLInDataBase struct{
	Fileds CRLFileds
	ID int
	PreID int
	Adding_time time.Time
	Expired_time time.Time
	// IsValid bool
}





func InsertOneCRL(db *sql.DB,CRLInfo CRLFileds,adding_time time.Time)(error){
	defaultTime, _:= time.Parse("2006-01-02 15:04:05", "9999-01-01 00:00:00")
	tx, err := db.Begin()
    if err != nil {
		return err
    }
	CRLList := make([]CRLInDataBase ,0)
	var oneCRL CRLInDataBase
	//根据插入证书的URI搜索表
	rows,err := tx.Query("Select * from CRLs where uri=? order by ID DESC",CRLInfo.URI) //按照ID降序，便于还原最大ID的条目
	if err!=nil{
		fmt.Println("INSERT CRL: GET CA ERROR!") //后面改成log
		return err
	}
	for rows.Next(){
		if err=rows.Scan(&oneCRL.Fileds.CRLNumber,&oneCRL.Fileds.AKI,&oneCRL.Fileds.RevokedCertificateList,&oneCRL.Fileds.ThisUpdate,&oneCRL.Fileds.NextUpdate,&oneCRL.Fileds.URI,&oneCRL.ID,&oneCRL.PreID,&oneCRL.Adding_time,&oneCRL.Expired_time,&oneCRL.Fileds.IsValid);err!=nil{
			fmt.Println("INSERT CRL: SCAN ROWS ERROR!") 
			return err
		}
		CRLList = append(CRLList,oneCRL)
	}

	if rows.Err() != nil {
        fmt.Println("INSERT CRL: ERROR EXIT SCANNING ROWS") 
		return err
    }

	//if nil  插入
	if len(CRLList)==0{
		rs,err := tx.Exec("INSERT INTO CRLs VALUES (?,?,?,?,?,?,?,?,?,?,?)",CRLInfo.CRLNumber,CRLInfo.AKI,CRLInfo.RevokedCertificateList,CRLInfo.ThisUpdate,CRLInfo.NextUpdate,CRLInfo.URI,0,-1,adding_time,defaultTime,CRLInfo.IsValid)
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
		var MaxIdItem CRLInDataBase
		var tempCRL CRLFileds
		MaxIdItem = CRLList[0]
		// for _,CRL := range CRLList{
		// 	if MaxIdItem.Fileds.CRLNumber == 0 && CRL.Fileds.CRLNumber!= -1 {
		// 		MaxIdItem.Fileds.CRLNumber=CRL.Fileds.CRLNumber
		// 	}
		// 	if MaxIdItem.Fileds.AKI == "" && CRL.Fileds.AKI!= "null" {
		// 		MaxIdItem.Fileds.AKI=CRL.Fileds.AKI
		// 	}
		// 	if MaxIdItem.Fileds.RevokedCertificateList == "" && CRL.Fileds.RevokedCertificateList!= "null" {
		// 		MaxIdItem.Fileds.RevokedCertificateList=CRL.Fileds.RevokedCertificateList
		// 	}
	

		// 	//下面每个字段操作同上(除了ID,preID,adding_time,expired_time,IsValid,这五个字段不进行增量存储)

		// }
		// MaxIdItem.Fileds.URI = CRLList[0].Fileds.URI
		// MaxIdItem.Fileds.ThisUpdate = CRLList[0].Fileds.ThisUpdate
		// MaxIdItem.Fileds.NextUpdate = CRLList[0].Fileds.NextUpdate
		// MaxIdItem.Fileds.IsValid =CRLList[0].Fileds.IsValid
		// MaxIdItem.ID = CRLList[0].ID
		// MaxIdItem.PreID = CRLList[0].PreID
		// MaxIdItem.Adding_time = CRLList[0].Adding_time
		// MaxIdItem.Expired_time =CRLList[0].Expired_time
			
		//逐条目比较{无差异，无操作；有差异，插入一条，更新上一条的失效时间为adding_time}
		if CRLInfo == MaxIdItem.Fileds{
			//do nothing
		}else{
			// if CRLInfo.CRLNumber != MaxIdItem.Fileds.CRLNumber{ //先这样遍历，后面学习怎么遍历结构体
			// 	tempCRL.CRLNumber = CRLInfo.CRLNumber
			// }else{
			// 	tempCRL.CRLNumber=-1
			// }
			// if CRLInfo.AKI != MaxIdItem.Fileds.AKI{ 
			// 	tempCRL.AKI = CRLInfo.AKI
			// }else{
			// 	tempCRL.AKI="null"
			// }
			// if CRLInfo.RevokedCertificateList != MaxIdItem.Fileds.RevokedCertificateList{ 
			// 	tempCRL.RevokedCertificateList = CRLInfo.RevokedCertificateList
			// }else{
			// 	tempCRL.RevokedCertificateList="null"
			// }
			tempCRL = CRLInfo
			
			//对剩下每个字段进行上述操作 （  :( , 真长 ）
			//插入   注意 preID此时不为默认值
			rs,err := tx.Exec("INSERT INTO CRLs VALUES (?,?,?,?,?,?,?,?,?,?,?)",tempCRL.CRLNumber,tempCRL.AKI,tempCRL.RevokedCertificateList,tempCRL.ThisUpdate,tempCRL.NextUpdate,tempCRL.URI,0,MaxIdItem.ID,adding_time,defaultTime,tempCRL.IsValid)
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
			rs,err = tx.Exec(`update CRLs set expired_time = ? where URI=? and id =?`,adding_time,MaxIdItem.Fileds.URI,MaxIdItem.ID)
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

			//根据撤销列表更新文件的expired_time
			var lastList RevokedCertificates
			var thisList RevokedCertificates
			err = json.Unmarshal([]byte(MaxIdItem.Fileds.RevokedCertificateList),&lastList)
			if err!=nil{
				fmt.Println("json解析失败")
				return err
			}
			err = json.Unmarshal([]byte(CRLInfo.RevokedCertificateList),&thisList)
			if err!=nil{
				fmt.Println("json解析失败")
				return err
			}

			var lastCRLMap map[*big.Int]time.Time
			// var thisCRLMap map[*big.Int]time.Time

			for _,i:=range lastList.RevokedCertificateList{
				lastCRLMap[i.SerialNumber] = i.RevocationTime
			}
			for _,j:=range thisList.RevokedCertificateList{
				revokeTime,ok := lastCRLMap[j.SerialNumber]
				if ok{
					if !revokeTime.Equal(j.RevocationTime){ //序列号相等但是时间不相等（不知道会不会有这种情况）
						rs,err = tx.Exec(`update CAs set expired_time = ? where AKI=? and serialNumber =? and expird_time>?`,j.RevocationTime,CRLInfo.AKI,j.SerialNumber,j.RevocationTime)
						if err != nil {
							tx_rollback(err,tx)
							fmt.Println(err.Error())
							return err
						}
					
						rs,err = tx.Exec(`update ROAs set expired_time = ? where AKI=? and serialNumber =? and expird_time>?`,j.RevocationTime,CRLInfo.AKI,j.SerialNumber,j.RevocationTime)
						if err != nil {
							tx_rollback(err,tx)
							fmt.Println(err.Error())
							return err
						}

						rs,err = tx.Exec(`update Mfts set expired_time = ? where AKI=? and serialNumber =? and expird_time>?`,j.RevocationTime,CRLInfo.AKI,j.SerialNumber,j.RevocationTime)
						if err != nil {
							tx_rollback(err,tx)
							fmt.Println(err.Error())
							return err
						}
					}
				}else{ //未在上次被处理过
					rs,err = tx.Exec(`update CAs set expired_time = ? where AKI=? and serialNumber =? and expird_time>?`,j.RevocationTime,CRLInfo.AKI,j.SerialNumber,j.RevocationTime)
					if err != nil {
						tx_rollback(err,tx)
						fmt.Println(err.Error())
						return err
					}
					
					rs,err = tx.Exec(`update ROAs set expired_time = ? where AKI=? and serialNumber =? and expird_time>?`,j.RevocationTime,CRLInfo.AKI,j.SerialNumber,j.RevocationTime)
					if err != nil {
						tx_rollback(err,tx)
						fmt.Println(err.Error())
						return err
					}

					rs,err = tx.Exec(`update Mfts set expired_time = ? where AKI=? and serialNumber =? and expird_time>?`,j.RevocationTime,CRLInfo.AKI,j.SerialNumber,j.RevocationTime)
					if err != nil {
						tx_rollback(err,tx)
						fmt.Println(err.Error())
						return err
					}

				}

			}
			

		}

	}


	err = tx.Commit()
	tx_rollback(err,tx)
	return nil
}