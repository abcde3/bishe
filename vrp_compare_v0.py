# encoding: utf-8
import os
import time
import datetime
import json
import pymysql

with open('config.json','r') as f:
    cfg = json.load(f)
last_vrp = '20211130.txt'
this_vrp = '20220107.txt'

def db_conn():
    db = pymysql.connect(host=cfg['mysql_host'], port=cfg['mysql_port'], user=cfg['mysql_user'], passwd=cfg['mysql_passwd'], db=cfg['mysql_db1'])
    return db

def make_compare_dict(last_vrp,this_vrp):
    f = open(last_vrp,'r')
    lines = f.readlines()
    f.close()
    vrp_dict = {}
    cnt = 0
    for line in lines:
        if line.startswith('\"ASN\"'):
            continue
        else:
            data = line.rstrip('\n').replace('\"','').split(',')
            vrp_dict[data[0]+' '+data[1]+' '+data[2]] = 1
            cnt += 1
            print(cnt,data[0]+' '+data[1]+' '+data[2])
    f = open(this_vrp,'r')
    lines = f.readlines()
    f.close()
    cnt = 0
    for line in lines:
        if line.startswith('\"ASN\"'):
            continue
        else:
            data = line.rstrip('\n').replace('\"','').split(',')
            tmp_key = data[0]+' '+data[1]+' '+data[2] # 数据格式更改在这里改变即可
            cnt += 1
            print(cnt,tmp_key)
            if tmp_key in vrp_dict:
                tmp_value = vrp_dict[tmp_key]
                vrp_dict[tmp_key] = tmp_value + 2
            else:
                vrp_dict[tmp_key] = 2
    return vrp_dict

def analysis_diff(vrp_dict):
    db = db_conn()
    cursor = db.cursor()
    insert_sql = "insert into vrp_change values(%s,%s,%s,%s,%s)"
    cnt = 0
    data_l = []
    today = datetime.datetime.now().strftime('%Y-%m-%d')
    # f = open('vrp_change.txt','w')
    for key,value in vrp_dict.items():
        prefix = key.split(' ')[1]
        origin = key.split(' ')[0]
        max_length = int(key.split(' ')[2])
        if value == 1:
            status = 'delete'
        elif value == 2:
            status = 'add'
        else:
            status = 'not change'
        # f.write(key+' '+status+' '+today+'\n')
        data = (prefix,origin,max_length,status,today)
        data_l.append(data)
        cnt += 1
        if cnt > 999:
            try:
                cursor.executemany(insert_sql, data_l)
                db.commit()
                cnt = 0 
                data_l = []
            except Exception as err:
                print(err)
    if cnt > 0:
            try:
                cursor.executemany(insert_sql, data_l)
                db.commit()
                cnt = 0 
                data_l = []
            except Exception as err:
                print(err)
    cursor.close()
    db.close()
begin = time.time()
tmp_dict = make_compare_dict(last_vrp,this_vrp)
analysis_diff(tmp_dict)
end = time.time()
print(end - begin)