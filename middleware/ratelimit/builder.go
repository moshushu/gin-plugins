package ratelimit

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Builder struct {
	// redis prefix
	prefix string
	// redis client
	cmd redis.Cmdable
	// 阈值
	threshold int64
	// 窗口（时间间隔）
	window time.Duration
	// log 记录
	logFunc func(logContent string)
}

//go:embed slide_window.lua
var luaScript string

func NewBuilder(cmd redis.Cmdable, threshold int64, window time.Duration) *Builder {
	return &Builder{
		prefix:    "ip-ratelimit",
		cmd:       cmd,
		threshold: threshold,
		window:    window,
		logFunc: func(logContent string) {
			log.Println(logContent)
		},
	}
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ok, err := b.limit(ctx)
		if err != nil {
			b.logFunc(err.Error())
			// 保守方案
			ctx.AbortWithStatus(http.StatusInternalServerError)
			// 激进方案
			// ctx.Next()
			return
		}
		if ok {
			// 限流
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	}
}

func (b *Builder) SetPrefix(prefix string) *Builder {
	b.prefix = prefix
	return b
}

func (b *Builder) SetLogFunc(logFunc func(logContent string)) *Builder {
	b.logFunc = logFunc
	return b
}

func (b *Builder) limit(ctx *gin.Context) (bool, error) {
	key := fmt.Sprintf("%s:%s", b.prefix, ctx.ClientIP())
	// agrs：窗口大小，阈值，当前时间（毫秒）
	return b.cmd.Eval(ctx, luaScript, []string{key},
		b.window.Milliseconds(), b.threshold, time.Now().UnixMilli()).Bool()
}
