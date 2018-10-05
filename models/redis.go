package models

import (
	"github.com/astaxie/goredis"
)

var (
	client goredis.Client
)

//定义常量
const (
	URL_QUEUE     = "url_queue"     //作为队列标识
	URL_VISIT_SET = "url_visit_set" //记录曾经访问过的url
)

func ConnectRedis(addr string) {
	client.Addr = addr
}

//把提取的url放入队列
func PutinQueue(url string) {
	client.Lpush(URL_QUEUE, []byte(url))
}

//从队列中取
func PopfromQueue() string {
	res, err := client.Rpop(URL_QUEUE)
	if err != nil {
		panic(err)
	}
	return string(res)
}

// 把曾经访问过的加入一个集合
func AddToSet(url string) {
	client.Sadd(URL_VISIT_SET, []byte(url))
}

//获取队列长度
func GetQueueLength() int {
	length, err := client.Llen(URL_QUEUE)
	if err != nil {
		return 0
	}

	return length
}

//判断某个URL是否访问过
func IsVisit(url string) bool {
	bIsVisit, err := client.Sismember(URL_VISIT_SET, []byte(url))
	if err != nil {
		return false
	}
	return bIsVisit

}
