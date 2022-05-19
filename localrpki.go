package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cloudflare/cfrpki/sync/lib"
	"github.com/cloudflare/cfrpki/validator/lib"
	"github.com/cloudflare/cfrpki/validator/pki"
	log "github.com/sirupsen/logrus"
	// "net"
	"sort"
	// "io"
	// "os"
	// "crypto/x509"
	"math/big"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"runtime"
	"strconv"
	"strings"
	"time"
	"a/SQL"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)
const (
	TYPE_UNKNOWN = iota
	TYPE_CER
	TYPE_MFT
	TYPE_ROA
	TYPE_CRL
	TYPE_ROACER
	TYPE_MFTCER
	TYPE_CAREPO
	TYPE_TAL
)

var (
	// RootTAL         = flag.String("tal.root", "tals/apnic.tal,tals/ripe.tal,tals/lacnic.tal,tals/arin.tal,tals/afrinic.tal", "List of TAL separated by comma")
	RootTAL         = flag.String("tal.root", "tals/afrinic.tal", "List of TAL separated by comma")
	
	//MapDir          = flag.String("map.dir", "rsync://rpki.apnic.net/=./repos/rpki.apnic.net/,rsync://rpki.ripe.net/=./repos/rpki.ripe.net/,rsync://repository.lacnic.net/=./repos/repository.lacnic.net/,rsync://rpki.afrinic.net/=./repos/rpki.afrinic.net/,rsync://rpki.arin.net/=./repos/rpki.arin.net/,rsync://ca.rg.net/=./repos/ca.rg.net/,rsync://repo-rpki.idnic.net/=./repos/repo-rpki.idnic.net/,rsync://rpki-repo.registro.br/=./repos/rpki-repo.registro.br/,rsync://rpkica.twnic.tw/=./repos/rpkica.twnic.tw/,rsync://rpki-repository.nic.ad.jp/=./repos/rpki-repository.nic.ad.jp/,rsync://rpki.cnnic.cn/=./repos/rpki.cnnic.cn/", "Map of the paths separated by commas")
	MapDir          = flag.String("map.dir", "rsync://rpki.apnic.net/=./repos/rpki.apnic.net/,rsync://rpki.ripe.net/=./repos/rpki.ripe.net/,rsync://repository.lacnic.net/=./repos/repository.lacnic.net/,rsync://rpki.afrinic.net/=./repos/rpki.afrinic.net/,rsync://rpki.arin.net/=./repos/rpki.arin.net/", "Map of the paths separated by commas")
	UseManifest     = flag.Bool("manifest.use", true, "Use manifests file to explore instead of going into the repository")
	StrictManifests = flag.Bool("strict.manifests", true, "Manifests must be complete or invalidate CA")
	StrictHash      = flag.Bool("strict.hash", true, "Check the hash of files")
	StrictCms       = flag.Bool("strict.cms", false, "Decode CMS with strict settings")
	ValidTime       = flag.String("valid.time", "now", "Validation time (now/timestamp/RFC3339)")
	LogLevel        = flag.String("loglevel", "info", "Log level")
	Output          = flag.String("output", "output.json", "Output file")

	SubjectInfoAccess   = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 1, 11}
	AuthorityInfoAccess = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 1, 1}
	SubjectKeyIdentifier   = asn1.ObjectIdentifier{2, 5, 29, 14}
	AuthorityKeyIdentifier = asn1.ObjectIdentifier{2, 5, 29, 35}
	CRLNumber = asn1.ObjectIdentifier{2, 5, 29, 20}
	SIAAccessMethod_mft = asn1.ObjectIdentifier{1,3,6,1,5,5,7,48,10}
)

// type OutputROA struct {
// 	ASN       string `json:"asn"`
// 	Prefix    string `json:"prefix"`
// 	MaxLength int    `json:"maxLength"`
// 	Path      string `json:"path"`
// 	//Time 	  string `json:"signingTime"`
// 	NotBefore string `json:notBefore`
// 	NotAfter  string `json:notAfter`
// }

// type OutputROAs struct {
// 	ROAs []OutputROA `json:"roas"`
// }

//解析AIA的结构体
type AIA struct{
	AccessMethod asn1.ObjectIdentifier
	GeneralName  []byte `asn1:"tag:6"`
}
//解析SIA的结构体
type SIA struct{
	AccessMethod asn1.ObjectIdentifier
	GeneralName  []byte `asn1:"tag:6"`
}

//解析ROA的ipAddrBlock需要的结构体，总共包含以下三个
type ROAIPAddress struct{
	Address string `json:"address"`
	MaxLength int `json:"maxLength"`
}

type ROAIPAddressFamily struct{
	AddressFamily int `json:"addressFamily"`
	Addresses []ROAIPAddress `json:"addresses"`
}
type IPAddrBlocks struct{
	IPAddrBlocks []ROAIPAddressFamily `json:"ipAddrBlocks"`
}

//解析manifest filelist json
// type BitString struct {
// 	Bytes     []byte // 比特打包成字节。
// 	BitLength int    // 比特长度。
// }

type ValueAndLength struct{
	Value []byte `json:"value"`
	Length int `json:"length"`
}

type File struct {
	File string `json:"file"`
	Hash ValueAndLength `json:"hash"`
}

type FileList struct{
	FileList []File `json:"fileList"`
}

type RevokedCertificates struct{
	RevokedCertificateList []pkix.RevokedCertificate
}


 
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()
	lvl, _ := log.ParseLevel(*LogLevel)
	log.SetLevel(lvl)
	log.Infof("Validator started")

	mapDir := syncpki.ParseMapDirectory(*MapDir)

	s := syncpki.LocalFetch{
		MapDirectory: mapDir,
		Log:          log.StandardLogger(),
	}

	var vt time.Time
	if *ValidTime == "now" {
		vt = time.Now().UTC()
	} else if ts, err := strconv.ParseInt(*ValidTime, 10, 64); err == nil {
		vt = time.Unix(int64(ts), 0)
		log.Infof("Setting time to %v (timestamp)", vt)
	} else if vttmp, err := time.Parse(time.RFC3339, *ValidTime); err == nil {
		vt = vttmp
		log.Infof("Setting time to %v (RFC3339)", vt)
	}

	rootTALs := strings.Split(*RootTAL, ",")
	// ors := OutputROAs{
	// 	ROAs: make([]OutputROA, 0),
	// }
	for _, tal := range rootTALs {
		validator := pki.NewValidator()
		validator.Time = vt

		validator.DecoderConfig.ValidateStrict = *StrictCms

		manager := pki.NewSimpleManager()
		manager.Validator = validator
		manager.FileSeeker = &s
		manager.ReportErrors = true
		manager.Log = log.StandardLogger()
		manager.StrictHash = *StrictHash
		manager.StrictManifests = *StrictManifests

		go func(sm *pki.SimpleManager) {
			for err := range sm.Errors {
				log.Error(err)
			}
		}(manager)

		manager.AddInitial([]*pki.PKIFile{
			&pki.PKIFile{
				Path: tal,
				Type: pki.TYPE_TAL,
			},
		})
		// var timeLayoutStr = "2006-01-02 15:04:05"
		manager.Explore(!*UseManifest, false)
		// manager.Explore(false, false)
		// fmt.Println(len(manager.Validator.Objects))
		// fmt.Println(len(manager.Validator.ValidObjects))
		// fmt.Println(len(manager.Validator.CRL))
		// fmt.Println(len(manager.Validator.ROA))
		// fmt.Println(len(manager.Validator.ValidROA))
		// fmt.Println(len(manager.Validator.Manifest))



		db, err := sql.Open("mysql", "root:123456@tcp(localhost:3306)/ca?parseTime=true") 
		defer db.Close()
		if err!=nil{
			fmt.Println("DB建立失败!")
			return
		}
		db.SetMaxOpenConns(10) //设置最大并发连接数
    	db.SetMaxIdleConns(2)  //设置闲置连接数
	//golang里未实现批量处理

	//检查数据库连接是否可用
		err=db.Ping() 
		if err!=nil{
			fmt.Println("DB不可使用")
			return
		}
		
		//Objects里包含RC和EE证书
		//需要的证书部分字段获取
		for cerKey, cer := range manager.Validator.Objects {
			if(cer.Type == TYPE_CER){
				var CAInfo mySql.CAFileds
				_,isValid := manager.Validator.ValidObjects[cerKey]
				// fmt.Println(isValid)
				CAInfo.IsValid = isValid
				d := cer.Resource.(*librpki.RPKICertificate)
				CAInfo.URI = manager.PathOfResource[cer].ComputePath()
				// fmt.Println(d.Certificate.SerialNumber) //serialNumber
				// fmt.Println(d.Certificate.SerialNumber.String())
				// fmt.Println(d.Certificate.NotBefore.Format(timeLayoutStr)) //NotBefore
				// fmt.Println(d.Certificate.NotAfter.Format(timeLayoutStr)) //NotAfter
				// fmt.Println(hex.EncodeToString(d.Certificate.SubjectKeyId)) //SKI
				// fmt.Println(hex.EncodeToString(d.Certificate.AuthorityKeyId)) //AKI
				// fmt.Println(d.Certificate.CRLDistributionPoints)
				CAInfo.SerialNumber = d.Certificate.SerialNumber.String()
				CAInfo.SKI = hex.EncodeToString(d.Certificate.SubjectKeyId)
				CAInfo.AKI = hex.EncodeToString(d.Certificate.AuthorityKeyId)
				CAInfo.NotBefore =d.Certificate.NotBefore
				CAInfo.NotAfter = d.Certificate.NotAfter
				// CAInfo.SubjectPublicKeyInfo =string(d.Certificate.RawSubjectPublicKeyInfo)
				CAInfo.SubjectPublicKeyInfo = "zanshihulue"
				ipaddresses := ""
				for _, i := range d.IPAddresses {
					ipaddresses += fmt.Sprintf("%v, ", i.String())
				}
				ipstr := fmt.Sprintf("[ %v]", ipaddresses)
				CAInfo.IPResources = ipstr

				asns := ""
				for _, i := range d.ASNums {
					asns += fmt.Sprintf("%v, ", i.String())
				}
				asstr := fmt.Sprintf("[ %v]", asns)
				CAInfo.ASResources = asstr
				for _,extension :=range d.Certificate.Extensions{
					if extension.Id.Equal(SubjectInfoAccess){  //SIA
						// fmt.Println("SIA")
						// fmt.Println(extension.Id)
						sias,err := ParseSIA(extension.Value)
						if err!=nil{
							fmt.Println("SIA解析错误！")
						}
						for _,sia :=range sias{
							// fmt.Println(sia.AccessMethod)
							// fmt.Println(string(sia.GeneralName))
							if sia.AccessMethod.Equal(SIAAccessMethod_mft){
								// fmt.Println(string(sia.GeneralName)) //SIA mft 将.mft替换成.crl就是SIA CRL
								CAInfo.SIAMft = string(sia.GeneralName)
								CAInfo.SIACRL =strings.Replace(string(sia.GeneralName), ".mft", ".crl", 1)
							}
						}
					}else if extension.Id.Equal(AuthorityInfoAccess){ //AIA
						// fmt.Println("AIA")
						// fmt.Println(extension.Id)
						aias,err := ParseAIA(extension.Value)
						if err!=nil{
							fmt.Println("AIA解析错误！")
						}
						for _,aia :=range aias{
							// fmt.Println(aia.AccessMethod)
							// fmt.Println(string(aia.GeneralName))
							CAInfo.AIA = string(aia.GeneralName)
						}
					}
				}
				mySql.InsertOneCA(db,CAInfo,validator.Time)
				
				// fmt.Println(d.IPAddresses) //IP Resources
				// fmt.Println(d.ASNums) //ASN Resoureces
				// fmt.Println(d.ASNRDI)
				
				// break
				
			}
			
		}
		//ROA独有字段获取
		for roaKey,roa := range manager.Validator.ROA{
			var ROAInfo mySql.ROAFileds
			d:= roa.Resource.(*librpki.RPKIROA)
			_,isValid := manager.Validator.ValidROA[roaKey]
			ROAInfo.IsValid = isValid
			// fmt.Println(roa.File.Path)
			ROAInfo.AsID = d.ASN
			ROAInfo.URI = roa.File.Path
			// fmt.Println(isValid)
			// fmt.Println(d.ASN) //ASN
			ipaddrblock,err:=EncodeROAEntries(d.Entries) //ipAddrBlock json格式
			if err!=nil{
				fmt.Println("ipAddrBlock还原失败")
			}
			ROAInfo.IpAddrBlocks = string(ipaddrblock)
			// fmt.Println(string(ipaddrblock))
			// for _,entry := range d.Valids{
			// 	fmt.Println(entry.IPNet.String())
			// 	fmt.Println(entry.MaxLength)
			// }
			// break
			//EE证书啊 :>
			EECer := d.Certificate
			ROAInfo.SerialNumber = EECer.Certificate.SerialNumber.String()
			ROAInfo.SKI = hex.EncodeToString(EECer.Certificate.SubjectKeyId)
			ROAInfo.AKI = hex.EncodeToString(EECer.Certificate.AuthorityKeyId)
			ROAInfo.NotBefore =EECer.Certificate.NotBefore
			ROAInfo.NotAfter = EECer.Certificate.NotAfter
			// CAInfo.SubjectPublicKeyInfo =string(d.Certificate.RawSubjectPublicKeyInfo)
			ROAInfo.SubjectPublicKeyInfo = "zanshihulue"
			ipaddresses := ""
			for _, i := range EECer.IPAddresses {
				ipaddresses += fmt.Sprintf("%v, ", i.String())
			}
			ipstr := fmt.Sprintf("[ %v]", ipaddresses)
			ROAInfo.IPResources = ipstr

			asns := ""
			for _, i := range EECer.ASNums {
				asns += fmt.Sprintf("%v, ", i.String())
			}
			asstr := fmt.Sprintf("[ %v]", asns)
			ROAInfo.ASResources = asstr
			for _,extension :=range EECer.Certificate.Extensions{
				if extension.Id.Equal(SubjectInfoAccess){  //SIA
					// fmt.Println("SIA")
					// fmt.Println(extension.Id)
					sias,err := ParseSIA(extension.Value)
					if err!=nil{
						fmt.Println("SIA解析错误！")
					}
					for _,sia :=range sias{
						// fmt.Println(sia.AccessMethod)
						// fmt.Println(string(sia.GeneralName))
						if sia.AccessMethod.Equal(SIAAccessMethod_mft){
							// fmt.Println(string(sia.GeneralName)) //SIA mft 将.mft替换成.crl就是SIA CRL
							ROAInfo.SIAMft = string(sia.GeneralName)
							ROAInfo.SIACRL = strings.Replace(string(sia.GeneralName), ".mft", ".crl", 1)
						}
					}
				}else if extension.Id.Equal(AuthorityInfoAccess){ //AIA
					// fmt.Println("AIA")
					// fmt.Println(extension.Id)
					aias,err := ParseAIA(extension.Value)
					if err!=nil{
						fmt.Println("AIA解析错误！")
					}
					for _,aia :=range aias{
						// fmt.Println(aia.AccessMethod)
						// fmt.Println(string(aia.GeneralName))
						ROAInfo.AIA = string(aia.GeneralName)
					}
				}
			}
			mySql.InsertOneROA(db,ROAInfo,validator.Time)
			// break

		}
	

		//清单独有字段获取
		for mftKey, mft := range manager.Validator.Manifest{
			var MftInfo mySql.ManifestFileds
			d:=mft.Resource.(*librpki.RPKIManifest)
			_,isValid := manager.Validator.ValidManifest[mftKey]
			MftInfo.IsValid = isValid
			MftInfo.ManifestNum = d.Content.ManifestNumber.String()
			MftInfo.ThisUpdate = d.Content.ThisUpdate
			MftInfo.NextUpdate = d.Content.NextUpdate
			MftInfo.URI = mft.File.Path
			// fmt.Println(isValid)
			// fmt.Println(d.Content.ManifestNumber) //ManifestNumber
			// fmt.Println(d.Content.ThisUpdate) //ThisUpdate
			// fmt.Println(d.Content.NextUpdate) // NextUpdate
			jsonFilelist,err:=FileListToJson(&d.Content)
			if err!=nil{
				fmt.Println("filelist转换为json格式失败")
			}
			// fmt.Println(string(jsonFilelist))
			MftInfo.FileList = string(jsonFilelist)
			// EE证书
			EECer := d.Certificate
			MftInfo.SerialNumber = EECer.Certificate.SerialNumber.String()
			MftInfo.SKI = hex.EncodeToString(EECer.Certificate.SubjectKeyId)
			MftInfo.AKI = hex.EncodeToString(EECer.Certificate.AuthorityKeyId)
			MftInfo.NotBefore =EECer.Certificate.NotBefore
			MftInfo.NotAfter = EECer.Certificate.NotAfter
			// CAInfo.SubjectPublicKeyInfo =string(d.Certificate.RawSubjectPublicKeyInfo)
			MftInfo.SubjectPublicKeyInfo = "zanshihulue"
			ipaddresses := ""
			for _, i := range EECer.IPAddresses {
				ipaddresses += fmt.Sprintf("%v, ", i.String())
			}
			ipstr := fmt.Sprintf("[ %v]", ipaddresses)
			MftInfo.IPResources = ipstr

			asns := ""
			for _, i := range EECer.ASNums {
				asns += fmt.Sprintf("%v, ", i.String())
			}
			asstr := fmt.Sprintf("[ %v]", asns)
			MftInfo.ASResources = asstr
			for _,extension :=range EECer.Certificate.Extensions{
				if extension.Id.Equal(SubjectInfoAccess){  //SIA
					// fmt.Println("SIA")
					// fmt.Println(extension.Id)
					sias,err := ParseSIA(extension.Value)
					if err!=nil{
						fmt.Println("SIA解析错误！")
					}
					for _,sia :=range sias{
						// fmt.Println(sia.AccessMethod)
						// fmt.Println(string(sia.GeneralName))
						if sia.AccessMethod.Equal(SIAAccessMethod_mft){
							// fmt.Println(string(sia.GeneralName)) //SIA mft 将.mft替换成.crl就是SIA CRL
							MftInfo.SIAMft = string(sia.GeneralName)
							MftInfo.SIACRL = strings.Replace(string(sia.GeneralName), ".mft", ".crl", 1)
						}
					}
				}else if extension.Id.Equal(AuthorityInfoAccess){ //AIA
					// fmt.Println("AIA")
					// fmt.Println(extension.Id)
					aias,err := ParseAIA(extension.Value)
					if err!=nil{
						fmt.Println("AIA解析错误！")
					}
					for _,aia :=range aias{
						// fmt.Println(aia.AccessMethod)
						// fmt.Println(string(aia.GeneralName))
						MftInfo.AIA = string(aia.GeneralName)
					}
				}
			}
			mySql.InsertOneMft(db,MftInfo,validator.Time)
			// break
		}

		for crlKey, crl := range manager.Validator.CRL {
			var CRLInfo mySql.CRLFileds
			d := crl.Resource.(*pkix.CertificateList)
			_,isValid := manager.Validator.ValidCRL[crlKey]
			CRLInfo.IsValid = isValid
			CRLInfo.URI = crl.File.Path
			CRLInfo.ThisUpdate =d.TBSCertList.ThisUpdate
			CRLInfo.NextUpdate = d.TBSCertList.NextUpdate
			// fmt.Println(isValid)
			// fmt.Println(crl.File.ComputePath())
			// fmt.Println(d.TBSCertList.ThisUpdate)
			// fmt.Println(d.TBSCertList.NextUpdate)
			RevokedCerListJson,err:= TBSCertListToJson(d) //revokeCertificateList 
			if err!=nil{
				fmt.Println("err")
			}
			// fmt.Println(string(RevokedCerListJson)) 
			CRLInfo.RevokedCertificateList = string(RevokedCerListJson)
			for _,extension :=range d.TBSCertList.Extensions{
				if extension.Id.Equal(AuthorityKeyIdentifier){//AKI
					// var key []byte
					// _,err := asn1.Unmarshal(extension.Value,&key)
					// if err!=nil{
					// 	fmt.Println("CRL的AKI解析错误")
					// }
					// fmt.Println(hex.EncodeToString(key))
					// fmt.Println(hex.EncodeToString(extension.Value))
					type KeyAuthority struct {
						Key []byte `asn1:"tag:0"`
					}
					var key KeyAuthority
					_, err := asn1.Unmarshal(extension.Value, &key)
					if err != nil {
						fmt.Println("CRL的AKI解析错误")
					}
					// fmt.Println(hex.EncodeToString(key.Key))
					CRLInfo.AKI = hex.EncodeToString(key.Key)
					
				}else if extension.Id.Equal(CRLNumber){ //CRLNumber
					var CRLNumber *big.Int  ///不知道行不行
					_, err := asn1.Unmarshal(extension.Value, &CRLNumber)
					if err != nil {
						fmt.Println("CRL的CRLnumber解析错误")
					}
					// fmt.Println(CRLNumber.String())
					// fmt.Println(CRLNumber)
					CRLInfo.CRLNumber = CRLNumber.String()
					// CRLInfo.CRLNumber = fmt.Sprintf("%d",CRL)

				}
				
			}
			mySql.InsertOneCRL(db,CRLInfo,validator.Time)
			// break
			
		}







		//manager.Explore(true,false)
		// for _, roa := range manager.Validator.ValidROA {
		// 	d := roa.Resource.(*librpki.RPKIROA)
		// 	for _, entry := range d.Valids {
		// 		oroa := OutputROA{
		// 			ASN:       fmt.Sprintf("AS%v", d.ASN),
		// 			Prefix:    entry.IPNet.String(),
		// 			MaxLength: entry.MaxLength,
		// 			Path:      manager.PathOfResource[roa].ComputePath(),
		// 			//Time:	   d.SigningTime.Format(timeLayoutStr),
		// 			NotAfter:  d.Certificate.Certificate.NotAfter.Format(timeLayoutStr),
		// 			NotBefore: d.Certificate.Certificate.NotBefore.Format(timeLayoutStr),
		// 		}
		// 		ors.ROAs = append(ors.ROAs, oroa)
		// 	}
		// }
	}

	// var buf io.Writer
	// var err error
	// if *Output != "" {
	// 	buf, err = os.Create(*Output)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// } else {
	// 	buf = os.Stdout
	// }

	// enc := json.NewEncoder(buf)
	// enc.SetIndent("", "    ")
	// enc.Encode(ors)
}
func ParseSIA(data []byte)([]SIA,error){
	var sias []SIA
	_,err := asn1.Unmarshal(data,&sias)
	if err!=nil{
		return sias,err
	}
	return sias,err
}
func ParseAIA(data []byte)([]AIA,error){
	var aias []AIA
	_,err := asn1.Unmarshal(data,&aias)
	if err!=nil{
		return aias,err
	}
	return aias,err
}
//分别解析ROAEntry并按照IP族分开
func GroupEntries(entries []*librpki.ROAEntry) map[byte][]*librpki.ROAEntry {
	mapIps := make(map[byte][]*librpki.ROAEntry)
	for _, entry := range entries {
		afi := byte(2)
		if entry.IPNet.IP.To4() != nil {
			afi = 1
		}

		ipsList, ok := mapIps[afi]
		if !ok {
			ipsList = make([]*librpki.ROAEntry, 0)
		}
		ipsList = append(ipsList, entry)

		mapIps[afi] = ipsList
	}
	return mapIps
}
//根据解析好的ROA 的Entry内容还原ipAddrBlocks的Json格式(RFC 6482)
func EncodeROAEntries(entries []*librpki.ROAEntry) ([]byte, error) {
	groups := GroupEntries(entries)

	versionList := make([]int, 0)
	for version, _ := range groups {
		versionList = append(versionList, int(version))
	}
	sort.Ints(versionList)
	roaFam := make([]ROAIPAddressFamily, 0)
	for _, cversion := range versionList {
		version := byte(cversion)

		listAddresses := make([]ROAIPAddress, 0)
		for _, v := range groups[version] {
			ipnetbs := v.IPNet.String()
			listAddresses = append(listAddresses, ROAIPAddress{
				Address:   ipnetbs,
				MaxLength: v.MaxLength,
			})
		}

		roa := ROAIPAddressFamily{
			AddressFamily: int(version),
			Addresses:     listAddresses,
		}
		roaFam = append(roaFam, roa)
	}

	ipAddrBlocks := IPAddrBlocks{
		IPAddrBlocks:	roaFam,
	}
	jsonIPAddrBlock, err := json.Marshal(ipAddrBlocks)
	if err != nil {
		return nil, err
	}
	return jsonIPAddrBlock, nil
}
//转换清单的filelist为Json格式（改了一些JSON的标签名称）
func FileListToJson(Manifest *librpki.ManifestContent) ([]byte, error) {
	Files :=make([]File,0)
	for _,entry :=range Manifest.FileList{
		hash := ValueAndLength{
			Value:	entry.Hash.Bytes,
			Length:	entry.Hash.BitLength,
		}
		file:=File{
			File: entry.Name,
			Hash: hash,
		}
		Files = append(Files,file)
	}
	FileList := FileList{
		FileList:	Files,
	}
	json,err := json.Marshal(FileList)
	if err!=nil{
		return nil,err
	}
	return json,err
}
//转换CRL的撤销列表为Json格式
func TBSCertListToJson(CRL *pkix.CertificateList)([]byte,error){
	RevokedCertificateList := RevokedCertificates{
		RevokedCertificateList: CRL.TBSCertList.RevokedCertificates,
	}
	json,err :=json.Marshal(RevokedCertificateList)
	if err!=nil{
		return nil,err
	}
	return json,err

}