docker rm mysql
 #   -e MYSQL_ROOT_PASSWORD=123456 \

docker run \
    -p 3309:3306 \
    --name mysql \
    -e MYSQL_ROOT_PASSWORD=123456 \
    -v $PWD/conf/my.cnf:/etc/my.cnf:ro \
    -v $PWD/data/mysql:/data:rw \
    -d mysql

#sleep 10
#mysql -uroot -P3309 -h127.0.0.1 -p123456 < privilege.sql
