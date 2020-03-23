# ins_for_go
go-colly爬取instagram

所需要的库
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

在下载开始前  请根据自己的代理 修改第58行 以及第162行

download文件夹 存放下载的视频以及图片

程序实现的功能 输入需要抓取的用户名 比如邓紫棋在instagram上的用户名为gem0816  会自动创建文件夹来存放邓紫棋的照片和视频
