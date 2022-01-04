import radix
import time
import pymysql
import json

with open('config.json','r') as f:
    cfg = json.load(f)

file_name = '20211130.txt' # 先手动输入
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
            node = rtree.add(data[1])
            node.data['asID'] = int(data[0])
            node.data['maxLength'] = int(data[2])
            node.data['root'] = data[3]
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
        return 'not found'
    else:
        max_length = int(bgp_ip.split('/')[1])
        if bgp_as.startswith('{') or bgp_as =='':
            return 'not known'
        if node.data['asID'] == int(bgp_as) and max_length<= node.data['maxLength']:
            return 'valid'
        else:
            return 'invalid'


def prove(rtree):
    # rtree = build_tree(file_name)
    cnt = 0
    insert_sql = "insert into result values(%s,%s,%s,%s,%s,%s) on duplicate key update " \
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
            print(asID)
            # print(ip_addr)
            # print(asID)
            # print(ip_type)
            result = get_result(rtree,ip_addr,asID)
            data = (ip_addr,asID,ip_type,result,first,last)
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
    # if cnt > 0:
    #         try:
    #             result_cursor.executemany(insert_sql, data_l)
    #             conn2.commit()
    #         except Exception as err:
    #             print(err)
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
# print(rtree.search_best('82.151.32.0/10'))