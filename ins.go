package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/gocolly/colly/proxy"
	"io"
	"net/url"
	"os"
	"regexp"
)

func main(){
	//urls := "https://www.instagram.com/graphql/query/?query_hash=e769aa130647d2354c40ea6a439bfc08&variables=%7B%22id%22%3A%224783464337%22%2C%22first%22%3A12%2C%22after%22%3A%22QVFEbExRZmFTY0wwSWVhUzJCYy1CcFRxNFpJTlVMNHJadlc3eDR1bWJsODlHSXN1SExuaWt0QjdHQUFfYTN3bjFOVkxjYjhHekRlQktyWTZQMjd5ZTJsMg%3D%3D%22%7D"
	//id := "4783464337"
	var targetName string
	fmt.Println("请输入要抓取的用户名：")
	fmt.Scanln(&targetName) //输入要抓取的用户名

	urls :=  "https://www.instagram.com/" +targetName  //拼接url
	aId, aCur := getIdAndCursor(urls)  //获取初始游标以及抓取用户的id
	fmt.Println(aCur,aId)

	fileNameImg := "./src/awesomeProject/src/download/"+targetName+"_img"
	fileNameVideo := "./src/awesomeProject/src/download/"+targetName+"_video"
	dirmk(fileNameImg)
	dirmk(fileNameVideo)  //创建对应用户存放图片和视频的文件夹

	v := url.Values{}
	v.Add("after", aCur)
	v.Add("first", "50")
	v.Add("id",aId)
	//拼接url
	body := "https://www.instagram.com/graphql/query/?query_hash=e769aa130647d2354c40ea6a439bfc08&"+  v.Encode()
	getCount(body,aId,fileNameImg, fileNameVideo)





}

func getCount(urls  string, id string ,fileNameImg string, fileNameVideo string) {
	//获取图片以及视频的url 并下载
	var path string   //存放路经
	c:=colly.NewCollector(func(collector *colly.Collector) {
		extensions.RandomUserAgent(collector) //随机UA

	})

	imageC :=c.Clone()

	//img_url := list.New()
	if p, err := proxy.RoundRobinProxySwitcher(
		"http://127.0.0.1:9666",  // 设置代理
	); err == nil {
		c.SetProxyFunc(p)
	}

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Cookie","ig_did=CDD46F6A-1B7A-43A8-8660-4C9C843FD8FE; mid=XncUvgALAAEGp32npsS73P_cbRKR; csrftoken=yCsmfUzGw1U6XNHPPOL2jGLOjs8dSth7; ds_user_id=32233184728; sessionid=32233184728%3A5lMCAovAehJdts%3A15; rur=FRC; urlgen='{\"65.49.126.82\": 6939}:1jFx1z:wXy5nuI6z2ntuF9F6pdWH7o-Zqs'")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Accept", "*/*")
	})    //添加请求头 和cookie

	c.OnRequest(func(request *colly.Request) {
		fmt.Printf("fetch --->%s\n",request.URL.String())

	})

	c.OnResponse(func(response *colly.Response) {

		var f interface{}
		json.Unmarshal(response.Body, &f) //反序列化
		fmt.Println(json.Unmarshal(response.Body, &f))
		data := f.(map[string]interface{})["data"]
		user :=data.(map[string]interface{})["user"]
		edge_owner := user.(map[string]interface{})["edge_owner_to_timeline_media"]

		// 找到总帖子数量是多少
		//count :=edge_owner.(map[string]interface{})["count"]

		//找到存放图片url 以及视频url的字段
		edges := edge_owner.(map[string]interface{})["edges"]
		for k,v := range edges.([]interface{}) {
			node := v.(map[string]interface{})["node"]
			imgUrl := node.(map[string]interface{})["display_url"].(string)  // 找到图片的url
			path = fileNameImg +"/"  //切换到存放图片的路径
			fmt.Println(k, imgUrl)
			imageC.Visit(imgUrl)  //扔给imageC下载

			if node.(map[string]interface{})["is_video"] == true{
				path = fileNameVideo +"/"  //切换到存放视频的路径
				videoUrl :=  node.(map[string]interface{})["video_url"].(string)
				fmt.Println(k, videoUrl)
				imageC.Visit(videoUrl) //扔给imageC下载

			}else {
				fmt.Println("此贴不包含视频")
			}

		}
		page_info :=edge_owner.(map[string]interface{})["page_info"]
		has_next_page :=page_info.(map[string]interface{})["has_next_page"]

		if has_next_page == true{
			//如果还有下一页 继续爬取
			end_cursor := page_info.(map[string]interface{})["end_cursor"].(string)
			v := url.Values{}
			v.Add("after", end_cursor)
			v.Add("first", "50")
			v.Add("id",id)
			body := "https://www.instagram.com/graphql/query/?query_hash=e769aa130647d2354c40ea6a439bfc08&"+  v.Encode()
			//拼接url
			c.Visit(body)
			fmt.Println("body>", body)
		}
	})
	c.OnError(func(response *colly.Response, err error) {
		fmt.Println(err)
	})
	fmt.Println(path)
	imageC.OnResponse(func(r *colly.Response) {
		fileName :=""
		reg := regexp.MustCompile("(\\d+_n.*)\\?") //获取图片或视频的文件名
		caption := reg.FindAllStringSubmatch(r.Request.URL.String(),-1)

		if caption == nil { //解释失败，返回nil
			fmt.Println("regexp err")
			return
		}
		fileName = caption[0][1]
		fmt.Println(fileName)
		fmt.Printf("下载 -->%s \n",fileName)
		f, err := os.Create(path+fileName)
		if err != nil {
			panic(err)
		}
		//io.Copy(f, bytes.NewReader(r.Body))
		io.Copy(f, bytes.NewReader(r.Body))  //下载
	})
	c.Limit(&colly.LimitRule{
		Parallelism: 2,
		//Delay:      5 * time.Second,
	})

	c.Visit(urls)
	c.Wait()
	imageC.Wait()
}

func getIdAndCursor(urls string) (string, string) {
	//获取初始的游标以及id
	var aId string
	var aCur string

	c:=colly.NewCollector(func(collector *colly.Collector) {
		extensions.RandomUserAgent(collector)

	})

	if p, err := proxy.RoundRobinProxySwitcher(
		"http://127.0.0.1:9666",  // 设置代理  需要根据自己的代理 进行更改
	); err == nil {
		c.SetProxyFunc(p)
	}

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Cookie","ig_did=CDD46F6A-1B7A-43A8-8660-4C9C843FD8FE; mid=XncUvgALAAEGp32npsS73P_cbRKR; csrftoken=yCsmfUzGw1U6XNHPPOL2jGLOjs8dSth7; ds_user_id=32233184728; sessionid=32233184728%3A5lMCAovAehJdts%3A15; rur=FRC; urlgen='{\"65.49.126.82\": 6939}:1jFx1z:wXy5nuI6z2ntuF9F6pdWH7o-Zqs'")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Accept", "*/*")
	})    //添加请求头 和cookie

	c.OnRequest(func(request *colly.Request) {
		fmt.Printf("fetch --->%s\n",request.URL.String())

	})


	c.OnHTML("body", func(e *colly.HTMLElement) {
		//fmt.Println("script:", e.Text)
		reg_id := regexp.MustCompile("profilePage_(\\d+)")  //正则匹配id
		reg_cur := regexp.MustCompile("\"end_cursor\":\"(.*?)\"") //正则匹配初始游标
		idres := reg_id.FindAllStringSubmatch(e.Text,-1)
		curres := reg_cur.FindAllStringSubmatch(e.Text,-1)
		//id := idres[0][1]
		//fmt.Println(id, curres[0][1])
		aId = idres[0][1]
		aCur = curres[0][1]
		//fmt.Println(aId, aCur)
	})
	//fmt.Println("asdasdasdasd",aId, aCur)
	c.Visit(urls)
	c.Wait()

	return aId ,aCur

}


func pathExists(path string) (bool, error) {
	//检查文件夹是否存在
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func dirmk(path string) {
	//如果文件夹不存在则创建
	_dir := path
	exist, err := pathExists(_dir)
	if err != nil {
		fmt.Printf("get dir error![%v]\n", err)
		return
	}
	if exist {
		fmt.Printf("has dir![%v]\n", _dir)
	} else {
		fmt.Printf("no dir![%v]\n", _dir)
		// 创建文件夹
		err := os.Mkdir(_dir, os.ModePerm)
		if err != nil {
			fmt.Printf("mkdir failed![%v]\n", err)
		} else {
			fmt.Printf("mkdir success!\n")
		}
	}
}




