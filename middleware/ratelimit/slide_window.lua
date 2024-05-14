-- 1、确定对象
local key = KEYS[1]
-- 2、确定窗口大小
local window = tonumber(ARGV[1])
-- 3、确定窗口阈值
local threshold = tonumber(ARGV[2])
-- 4、确定当前时间
local now = tonumber(ARGV[3])
-- 5、根据窗口大小和当前时间，确定窗口的其实时间
local min = now - window

-- 6、清除当前时间窗口之前的记录
-- 目的：确保当前时间窗口的记录保留，同时避免导致积累大量的过期记录，从而影响计数的准确性和性能
redis.call("ZREMRANGEBYSCORE",key,"-inf",min)

-- 7、计量当前时间窗口的记录数量
local count = redis.call("ZCOUNT",key,"-inf","+inf")

-- 8、比较当前时间窗口的记录数据和窗口阈值
if count >= threshold then
    -- 限流
    return "true"
else
-- 9、往集合中插入一条记录（用当前时间作为记录，确保独一无二）
    redis.call("ZADD",key,now,now)
-- 10、设置key的过期时间
    -- 确保数据能够及时清理，减少内存占用
    redis.call("PEXPIRE",key,window)
    -- 不限流
    return "false"
end

