docker run -it\
    -p 6389:6389 \
    -v $PWD/data:/data:rw \
    -v $PWD/conf/redis.conf:/etc/redis/redis.conf:ro \
    --privileged=true \
    --name redis-m \
    -d redis redis-server /etc/redis/redis.conf
