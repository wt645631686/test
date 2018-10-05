package controllers

import (
	"crawl_movie/models"
	"fmt"
	"runtime"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
)

type CrawlMovieController struct {
	beego.Controller
}

func PrintErr() {
	if err := recover(); err != nil {
		fmt.Printf("%v", err)
		for i := 0; i < 10; i++ {
			funcName, file, line, ok := runtime.Caller(i)
			if ok {
				fmt.Printf("frame %v:[func:%v,file:%v,line:%v]\n", i, runtime.FuncForPC(funcName).Name(), file, line)
			}
		}
	}
}
func (c *CrawlMovieController) CrawlMovie() {
	PrintErr()
	var movieInfo models.MovieInfo //先声明电影信息结构

	models.ConnectRedis("127.0.0.1:6379") //连接redis

	//爬虫入口url
	sUrl := "https://movie.douban.com/subject/1900841/?from=subject-page" //我不是药神 电影url详情页 //这里作为入口
	models.PutinQueue(sUrl)

	for {
		length := models.GetQueueLength()
		c.Ctx.WriteString(fmt.Sprintf("---%v---", length))
		if length == 0 {
			break //如果url队列为空，则退出当前循环
		}
		sUrl = models.PopfromQueue()
		//判断url是否已经被访问过
		if models.IsVisit(sUrl) { //访问过则跳过
			continue
		}
		rsp := httplib.Get(sUrl)
		//设置User-agent以及cookie是为了防止  豆瓣网的 403
		rsp.Header("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64; rv:50.0) Gecko/20100101 Firefox/50.0")
		rsp.Header("Cookie", `bid=gFP9qSgGTfA; __utma=30149280.1124851270.1482153600.1483055851.1483064193.8; __utmz=30149280.1482971588.4.2.utmcsr=douban.com|utmccn=(referral)|utmcmd=referral|utmcct=/; ll="118221"; _pk_ref.100001.4cf6=%5B%22%22%2C%22%22%2C1483064193%2C%22https%3A%2F%2Fwww.douban.com%2F%22%5D; _pk_id.100001.4cf6=5afcf5e5496eab22.1482413017.7.1483066280.1483057909.; __utma=223695111.1636117731.1482413017.1483055857.1483064193.7; __utmz=223695111.1483055857.6.5.utmcsr=douban.com|utmccn=(referral)|utmcmd=referral|utmcct=/; _vwo_uuid_v2=BDC2DBEDF8958EC838F9D9394CC5D9A0|2cc6ef7952be8c2d5408cb7c8cce2684; ap=1; viewed="1006073"; gr_user_id=e5c932fc-2af6-4861-8a4f-5d696f34570b; __utmc=30149280; __utmc=223695111; _pk_ses.100001.4cf6=*; __utmb=30149280.0.10.1483064193; __utmb=223695111.0.10.1483064193`)
		sMovieHtml, err := rsp.String()

		if err != nil {
			panic(err)
		}

		//获取电影名称
		movieInfo.Movie_name = models.GetMovieName(sMovieHtml)
		if movieInfo.Movie_name != "" { //如果为空，则说明不是电影，如果不为空，则是电影
			//获取电影导演
			movieInfo.Movie_director = models.GetMovieDirector(sMovieHtml)
			//获取主演
			movieInfo.Movie_main_character = models.GetMovieMainCharacters(sMovieHtml)
			//电影类型
			movieInfo.Movie_type = models.GetMovieGenre(sMovieHtml)
			//上映时间
			movieInfo.Movie_on_time = models.GetMovieOnTime(sMovieHtml)
			//评分
			movieInfo.Movie_grade = models.GetMovieGrade(sMovieHtml)
			//时长
			movieInfo.Movie_span = models.GetMovieRunningTime(sMovieHtml)
			//	c.Ctx.WriteString(fmt.Sprintf("%v", movieInfo))
			//入库
			models.AddMovie(&movieInfo)
			//	id, _ := models.AddMovie(&movieInfo)
			//	c.Ctx.WriteString(fmt.Sprintf("%v", id))
		}

		//提取该页面的所有连接
		urls := models.GetMovieUrls(sMovieHtml)

		//遍历url
		//为了把url写入队列
		//同样需要开启一个协程，这个协程专门负责从队列中取，负责get，set，
		//第一判断这个url是不是一个电影，是的话加入到数据库，
		//	第二是提取这个电影有关的url
		//第三把url放入set(集合)里，表明这个url已经访问过
		for _, url := range urls {
			models.PutinQueue(url)
			c.Ctx.WriteString("<br>" + url + "</br>")
		}
		//sUrl 需要记录到set集合里，表明这个url访问过
		models.AddToSet(sUrl)
		time.Sleep(time.Second) //适当休息
	}
	c.Ctx.WriteString("爬虫执行结束")

}
