# 申通快递 SDK for Go

这是申通快递开放平台的 Go SDK，支持运单轨迹查询等功能。

## 安装

```bash
go get github.com/your-username/sto-sdk-go
```

## 功能特性

- 支持运单轨迹查询
- 支持批量查询多个运单号
- 支持轨迹排序（升序/降序）
- 内置自动重试机制
- 支持调试模式
- 完整的错误处理
- 可配置的 HTTP 客户端

## 使用示例

### 运单轨迹查询

```go
package main

import (
    "fmt"
    "log"
    "github.com/your-username/sto-sdk-go/sto"
)

func main() {
    // 创建客户端实例
    client := sto.NewClient(
        "YOUR_APP_KEY",        // 申通开放平台申请的APP KEY
        "YOUR_APP_SECRET",     // 申通开放平台申请的APP SECRET
        "YOUR_FROM_CODE",      // 商户编码，通常与APP KEY相同
    )

    // 可选：开启调试模式，打印请求和响应信息
    client.EnableDebug()

    // 创建查询请求
    req := &sto.TraceQueryRequest{
        Order:         "asc",                    // 排序方式：asc（升序）或desc（降序）
        WaybillNoList: []string{"运单号"},        // 运单号列表，支持多个运单号
    }

    // 发送查询请求
    resp, err := client.QueryTrace(req)
    if err != nil {
        log.Fatalf("查询失败: %v", err)
    }

    // 处理响应
    if resp.IsSuccess() {
        for waybillNo, traces := range resp.Data {
            fmt.Printf("运单号: %s\n", waybillNo)
            if len(traces) == 0 {
                fmt.Println("暂无轨迹信息")
                continue
            }
            for _, t := range traces {
                fmt.Printf("操作时间: %s\n", t.OpTime)
                fmt.Printf("扫描类型: %s\n", t.ScanType)
                fmt.Printf("操作机构: %s (%s)\n", t.OpOrgName, t.OpOrgCode)
                fmt.Printf("所在城市: %s省%s市\n", t.OpOrgProvinceName, t.OpOrgCityName)
                if t.Weight != "" {
                    fmt.Printf("重量: %skg\n", t.Weight)
                }
                fmt.Println("---")
            }
        }
    } else {
        fmt.Printf("查询失败: %s - %s\n", resp.ErrorCode, resp.ErrorMsg)
        if resp.ExpInfo != "" {
            fmt.Printf("异常信息: %s\n", resp.ExpInfo)
        }
    }
}
```

## 配置选项

创建客户端时可以使用以下可选配置：

```go
// 设置超时时间（默认30秒）
client := sto.NewClient(
    "YOUR_APP_KEY",
    "YOUR_APP_SECRET",
    "YOUR_FROM_CODE",
    sto.WithTimeout(30*time.Second),
)

// 设置最大重试次数（默认3次）
client := sto.NewClient(
    "YOUR_APP_KEY",
    "YOUR_APP_SECRET",
    "YOUR_FROM_CODE",
    sto.WithMaxRetries(3),
)

// 设置自定义HTTP客户端
httpClient := &http.Client{
    Timeout: 30 * time.Second,
    // 可以配置其他 HTTP 客户端选项
}
client := sto.NewClient(
    "YOUR_APP_KEY",
    "YOUR_APP_SECRET",
    "YOUR_FROM_CODE",
    sto.WithHTTPClient(httpClient),
)
```

## 请求和响应说明

### TraceQueryRequest 请求参数

| 字段名 | 类型 | 必填 | 说明 |
|-------|------|------|------|
| Order | string | 否 | 排序方式：asc（升序）或desc（降序），默认asc |
| WaybillNoList | []string | 是 | 运单号列表，支持批量查询多个运单号 |

### TraceQueryResponse 响应结构

| 字段名 | 类型 | 说明 |
|-------|------|------|
| Success | string | 是否成功，"true"或"false" |
| ErrorCode | string | 错误码，请求失败时返回 |
| ErrorMsg | string | 错误信息，请求失败时返回 |
| NeedRetry | string | 是否需要重试，"true"或"false" |
| RequestId | string | 请求ID，用于问题排查 |
| ExpInfo | string | 异常信息，请求异常时返回 |
| Data | map[string][]TraceInfo | 运单号对应的轨迹列表，key为运单号，value为轨迹信息数组 |

### TraceInfo 物流轨迹信息

| 字段名 | 类型 | 说明 |
|-------|------|------|
| WaybillNo | string | 运单号 |
| OpTime | string | 操作时间，格式：yyyy-MM-dd HH:mm:ss |
| OpOrgCode | string | 操作机构代码 |
| OpOrgName | string | 操作机构名称 |
| OpOrgProvinceName | string | 操作机构所在省 |
| OpOrgCityName | string | 操作机构所在市 |
| OpOrgTel | string | 操作机构电话 |
| OpEmpCode | string | 操作员工号 |
| OpEmpName | string | 操作员姓名 |
| ScanType | string | 扫描类型 |
| Weight | string | 重量，单位：kg |
| Memo | string | 备注信息 |
| BizEmpCode | string | 业务员工号 |
| BizEmpName | string | 业务员姓名 |
| BizEmpPhone | string | 业务员手机号码 |
| BizEmpTel | string | 业务员固定电话 |
| NextOrgName | string | 下一站机构名称 |
| NextOrgCode | string | 下一站机构代码 |
| IssueName | string | 问题件名称 |
| SignoffPeople | string | 签收人姓名 |
| ContainerNo | string | 集包号 |
| OrderOrgCode | string | 下单机构代码 |
| OrderOrgName | string | 下单机构名称 |
| TransportTaskNo | string | 运输任务号 |
| CarNo | string | 车牌号 |
| OpOrgTypeCode | string | 操作机构类型代码 |
| PartnerName | string | 品牌方名称 |

### 错误码说明

| 错误码 | 说明 | 处理建议 |
|-------|------|---------|
| 005 | 运单号错误 | 检查运单号是否正确 |
| 006 | 无权限访问 | 检查 APP KEY 和 APP SECRET 是否正确，以及是否有接口调用权限 |
| 007 | 签名错误 | 检查签名生成逻辑是否正确，APP SECRET 是否正确 |
| 008 | 请求参数错误 | 检查请求参数是否完整且正确，参考接口文档 |
| 009 | 系统繁忙 | 请稍后重试，如果持续出现请联系技术支持 |

## 错误处理

SDK 会返回详细的错误信息，包括：

- ErrorCode: 错误码，用于识别具体错误类型
- ErrorMsg: 错误信息，描述错误的具体原因
- ExpInfo: 异常信息，提供更详细的错误描述
- RequestId: 请求ID，用于问题排查和跟踪
- NeedRetry: 是否需要重试，如果为 "true" 表示可以尝试重新发送请求

## 调试模式

可以通过 `EnableDebug()` 和 `DisableDebug()` 方法开启或关闭调试模式：

```go
client.EnableDebug()  // 开启调试模式
client.DisableDebug() // 关闭调试模式
```

开启调试模式后，SDK 会打印以下信息：
- 完整的请求URL
- 请求参数
- 签名信息
- HTTP响应状态码
- 响应内容

## 注意事项

1. 请妥善保管您的 APP SECRET，不要泄露给他人
2. 建议在生产环境中关闭调试模式，避免打印敏感信息
3. 如果遇到请求失败，请查看错误信息和异常信息，根据错误码说明进行处理
4. 如果响应中 NeedRetry 为 "true"，表示需要重试请求，SDK 会自动进行重试
5. 批量查询时建议合理控制运单号数量，避免超时
6. 在并发请求场景下，建议复用 Client 实例
7. 正式环境中建议配置适当的超时时间和重试次数

## License

MIT License 