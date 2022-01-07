import radix
import time
import pymysql
import json

with open('config.json','r') as f:
    cfg = json.load(f)

file_name = '20211130.txt'
#db = pymysql.connect(host=cfg['mysql_host'], port=cfg['mysql_port'], user=cfg['mysql_user'], passwd=cfg['mysql_passwd'], db=cfg['mysql_db2'])
#db.close()

def build_tree(file_name):
    rtree = radix.Radix()
    cnt = 0
    f = open(file_name,'r')
    lines = f.readlines()
    f.close()
    for line in lines:
        # print(line)
        if line.startswith('\"ASN\"'):
            continue
        else:
            data = line.rstrip('\n').replace('\"','').split(',')
            tmp = rtree.search_exact(data[1])
            if tmp == None:
                node = rtree.add(data[1])
                node.data['asID'] = [int(data[0])]
                node.data['maxLength'] = [int(data[2])]
                node.data['root'] = [data[3]]
            else:
                tmp.data['asID'].append(int(data[0]))
                tmp.data['maxLength'].append(int(data[2]))
                tmp.data['root'].append(data[3])
            cnt += 1
    return rtree
# print('hello')

def bgp_db_conn():
    db = pymysql.connect(host=cfg['mysql_host'], port=cfg['mysql_port'], user=cfg['mysql_user'], passwd=cfg['mysql_passwd'], db=cfg['mysql_db2'])
    return db
def result_db_conn():
    db = pymysql.connect(host=cfg['mysql_host'], port=cfg['mysql_port'], user=cfg['mysql_user'], passwd=cfg['mysql_passwd'], db=cfg['mysql_db1'])
    return db

def get_result(rtree,bgp_ip,bgp_as):
    node = rtree.search_best(bgp_ip)
    if node == None:
        return 'not found','not coverd by any node'
    else:
        max_length = int(bgp_ip.split('/')[1])
        if bgp_as.startswith('{') or bgp_as =='':
            return 'wrong data',''
        #count = len(node.data['asID'])
        #flag = False
        if int(bgp_as) in node.data['asID']:
            index = node.data['asID'].index(int(bgp_as))
            if max_length <= node.data['maxLength'][index]:
                return 'valid','Covered by '+node.prefix+' '+str(node.data['asID'][index])+'.Maxlength is '+str(node.data['maxLength'][index])
            else:
                return 'invalid','Covered by '+node.prefix+' '+str(node.data['asID'][index])+'.Maxlength is '+str(node.data['maxLength'][index])+'.But the length of announced prefix is illegal'
        else:
            return 'invalid','The ASN is not the owner of the prefix '+bgp_ip+'.'
        
def prove(rtree):
    # rtree = build_tree(file_name)
    cnt = 0
    insert_sql = "insert into result2 values(%s,%s,%s,%s,%s,%s,%s,%s) on duplicate key update " \
                "result=if(first>values(first),values(result),result)"
    conn1 = bgp_db_conn()
    conn2 = result_db_conn()
    cursor = pymysql.cursors.SSCursor(conn1)
    cursor.execute('select * from origins;')
    result_cursor = conn2.cursor()
    data_l = []
    count = 0
    while True:
        rows = cursor.fetchmany(1000)
        if not rows:
            break
        # print(row)
        for row in rows:
            ip_addr = row[0]
            asID = row[1]
            ip_type = row[2]
            first = row[6]
            last = row[7]
            roa_file = ''
            #print(asID)
            result,reason = get_result(rtree,ip_addr,asID)
            data = (ip_addr,asID,ip_type,result,reason,roa_file,first,last)
            data_l.append(data)
            cnt += 1
            count += 1
            print(count)
        if cnt > 0:
            try:
                result_cursor.executemany(insert_sql, data_l)
                conn2.commit()
                cnt =0 
                data_l = []
            except Exception as err:
                print(err)
    result_cursor.close()
    cursor.close()
    conn1.close()
    conn2.close()

#prove()
begin = time.time()
rtree = build_tree(file_name)
prove(rtree)
end = time.time()
print(end -begin)