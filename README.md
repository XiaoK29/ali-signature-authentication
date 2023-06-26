# ali-signature-authentication
#### 项目背景

由于工作中使用go语言，阿里那边没有相关的SDK，只能[使用摘要签名认证方式调用API](https://help.aliyun.com/document_detail/29475.html?spm=a2c4g.157953.0.0.122417adz3x7cg)方式来调用服务端api，项目基于[resty](https://github.com/go-resty/resty)封装具体文档可看该项目使用方法

#### 使用

```go
import client "github.com/XiaoK29/ali-signature-authentication"
```

```go
var result map[string]any
resp, err := client.NewClient("your_APPKey", "your_APPSecret").SetContext(context.Background()).
		SetQueryParam("alipayUid", "xxxxxxxxxxxxxxxx").
		SetResult(&result).
		GET("https://test.com/xxxx")
fmt.Println(result)
fmt.Println(resp.StatusCode())
fmt.Println(resp.String())
fmt.Println(resp.Header())
```
