# Decentralized exchange running on linkchain.
## 编译说明
`go` version >=1.12

linkchain节点与钱包配置，见[linkchain](https://github.com/lianxiangcloud/linkchain)项目

## 编译&安装
```
cd lkdex
make
```
编译完成后，生成文件存于`bin/lkdex`目录

## 启动脚本程序
启动/停止客户端脚本`sbin/dex.sh`

`cd sbin`
`./dex.sh start`

## 客户端命令行启动
客户端RPC服务默认端口为18804,具体配置见config/config.go
```
./bin/lkdex node --home <数据地址> --contract_addr <合约地址> --daemon.peer_rpc <链节点地址> --daemon.peer_ws <链节点ws地址> --wallet_daemon.peer_rpc <钱包地址>
例子
./bin/lkdex node --home ./lkdata --contract_addr 0x28bc0a05d787ff27213322087a8911e1b2c5eacf --daemon.peer_rpc http://10.9.194.103:46000 --daemon.peer_ws ws://10.9.194.103:44000 --wallet_daemon.peer_rpc http://10.9.194.103:18082 
```
## 说明
- contract/dex.cpp 交易所合约源码
- doc/Dexcentralized exchange.pdf 交易所设计方案

## 客户端数据库
客户端将订单与交易数据存储在`dex<合约地址>.db`文件下。数据格式为`sqlite3`
### 订单数据库表名
`order_models`

### 数据库字段格式(gorm)
````
//OrderModel Order DateBase
type OrderModel struct {
	HashID       string          `gorm:"primary_key;type:char(66)"`     //Order HashID
	TokenGet     string          `gorm:"type:char(42);not null"`        //Get Token Address
	AmountGet    string          `gorm:"not null"`                      //Get Token Amount
	TokenGive    string          `gorm:"type:char(42);not null"`        //Give Token Address
	AmountGive   string          `gorm:"not null"`                      //Give Token Amount
	Nonce        sql.NullInt64   `gorm:"not null"`                      //Nonce
	Expires      sql.NullInt64   `gorm:"not null"`                      //Expire time (Unix Timestamp)
	Maker        string          `gorm:"type:char(42);not null"`        //Maker Address
	R            string          `gorm:"type:char(34);not null"`        //Sign R
	S            string          `gorm:"type:char(34);not null"`        //Sign S
	V            string          `gorm:"type:char(4);not null"`         //Sign V
	State        sql.NullInt64   `gorm:"not null"`                      //0:Sending(not save in block)  1:Trading  2:Finish(Cancel)
	Price        sql.NullFloat64 `gorm:"type:numeric(225,20);not null"` //Order Price Calculated from AmountGive and AmountGet
	FilledAmount string          `gorm:"not null"`                      //Order FilledAmount default:0
````
SQL表名

`hash_id|token_get|amount_get|token_give|amount_give|nonce|expires|maker|r|s|v|state|price|filled_amount`

### 历史交易数据库表名
`trade_models`
### 数据库字段格式(gorm)
````
//TradeModel Trade history DateBase
type TradeModel struct {
	gorm.Model
	HashID       string        `gorm:"type:char(66);FOREIGNKEY"` //Order Hash
	DealAmount   string        `gorm:"not null"`                 //Deal amount
	FilledAmount string        `gorm:"not null"`                 //Trade Amount
	BlockNum     sql.NullInt64 `gorm:"not null"`                 //Deal BlockNum
	TxHash       string        `gorm:"type:char(66);not null"`   //Deal Tx hash
	Taker        string        `gorm:"type:char(42);not null"`   //Taker Address
}
````

SQL表名

`id|created_at|updated_at|deleted_at|hash_id|deal_amount|filled_amount|block_num|tx_hash|taker`

#### 相关查询SQL例子

`sqlite3 -line dex0x....db 'select * from order_models;'`
##### 查询所有的订单
`select * from order_models;`

##### 查询支持交易对
`select distinct token_get,token_give from order_models;`

##### 查询指定交易，按价格排序
`select* from order_models where token_get='0xd8b9c3ea884bccdd67c1d9dd115b75cf9f969879' and token_give='0xcbf2a8db3ca6499db97d447f21a0a57198387f61' order by price;`

##### 查询指定交易对的所有交易信息
`select * from trade_models t JOIN order_models o ON t.hash_id = o.hash_id where o.token_get='0xd8b9c3ea884bccdd67c1d9dd115b75cf9f969879' and o.token_give='0xcbf2a8db3ca6499db97d447f21a0a57198387f61';`

##### 查询指定交易对,最新的成交价格(当前市价)
`select price from trade_models t JOIN order_models o ON t.hash_id = o.hash_id where o.token_get='0xd8b9c3ea884bccdd67c1d9dd115b75cf9f969879' and o.token_give='0xcbf2a8db3ca6499db97d447f21a0a57198387f61' order by  block_num desc limit 1 ;`

##### 查询所有的交易
`select * from trade_models;`

##### 查询指定账户的交易记录,按块号排序
`select * from trade_models where taker='0x7eaaae9a69a66559553d41d34405a3377a7fe000';`

## dex相关接口
### dex_getDepositAmount
获取抵押的资金额度
#### 参数
- `address` 账户地址
- `address` token地址
#### 返回
- 抵押金额

#### 示例
```shell
curl -s -X POST http://127.0.0.1:18804 -d '{"jsonrpc":"2.0","method":"dex_getDepositAmount","params":["0xa73810e519e1075010678d706533486d8ecc8000","0x95ccc08ab44ac6d071a0c5911df64ad2394a4123"],"id":67}' -H 'Content-Type:application/json'
```
```
{"jsonrpc":"2.0","id":67,"result":"0x1"}
```
### dex_getOrderByHash
#### 参数
- `hash` 订单hash
#### 返回值
- 订单信息
  - `tokenGet` 需要交换的token地址
  - `amountGet`  交换token数量
  - `tokenGive`  交换的token地址
  - `amountGive`  交换的token数量
  - `nonce`  交易nonce值
  - `expire`  过期时间
  - `maker`  订单发起人地址（签名地址）
  - `r`  订单签名
  - `s`  订单签名
  - `v`  订单签名

### dex_getOrderHash
获取交易hash
#### 参数
- `tokenGet` 需要交换的token地址
- `amountGet`  交换token数量
- `tokenGive`  交换的token地址
- `amountGive`  交换的token数量
- `nonce`  交易nonce值
- `expire`  过期时间
- `maker`  订单发起人地址（签名地址）

#### 返回值
- `hash` 订单hash

### dex_getSignOrderHash
获取含签名值的交易hash
#### 参数
- `tokenGet` 需要交换的token地址
- `amountGet`  交换token数量
- `tokenGive`  交换的token地址
- `amountGive`  交换的token数量
- `nonce`  交易nonce值
- `expire`  过期时间
- `maker`  订单发起人地址（签名地址）
- `r`  订单签名
- `s`  订单签名
- `v`  订单签名
#### 返回值
- `hash` 订单hash


## wallet相关接口（钱包解锁账户）
### 错误说明
接口调用出错可能
1.合约执行失败
2.钱包没有解锁对应账户
3.参数错误
4.钱包没有连接到节点

### wlt_signOrder
订单签名
#### 参数
- `tokenGet` 需要交换的token地址
- `amountGet`  交换token数量
- `tokenGive`  交换的token地址
- `amountGive`  交换的token数量
- `nonce`  交易nonce值
- `expire`  过期时间
- `maker`  订单发起人地址（签名地址）

#### 返回
- `order` 订单信息
- `r` 签名
- `s` 签名 
- `v` 签名

#### 示例
```shell
curl -s -X POST http://127.0.0.1:18804 -d '{"jsonrpc":"2.0","method":"wlt_signOrder","params":[{"tokenGet":"0x95ccc08ab44ac6d071a0c5911df64ad2394a4c54", "amountGet":"0x1", "tokenGive":"0x95ccc08ab44ac6d071a0c5911df64ad2394a4c54","amountGive":"0x1", "expires":"0x1", "nonce":"0x1", "maker":"0xa73810e519e1075010678d706533486d8ecc8000"}],"id":67}' -H 'Content-Type:application/json'
```
```
{"jsonrpc":"2.0","id":67,"result":{"order":{"tokenGet":"0x95ccc08ab44ac6d071a0c5911df64ad2394a4c54","amountGet":"0x1","tokenGive":"0x95ccc08ab44ac6d071a0c5911df64ad2394a4c54","amountGive":"0x1","expires":"0x5dafc158","nonce":"0x1","maker":"0xa73810e519e1075010678d706533486d8ecc8000"},"v":"0x1b","s":"0x71b0df728b6639f405446be1f98f5ef1bdb1932259456d6b9b4bc3e1bdc19d57","r":"0x809cc318a5eb5a14c9073d0858a98efe268fcf5852d6bc76778147f064f00d5a"}}
```

### wlt_postOrder
提交订单
#### 参数
- `tokenGet` 需要交换的token地址
- `amountGet`  交换token数量
- `tokenGive`  交换的token地址
- `amountGive`  交换的token数量
- `nonce`  交易nonce值
- `expire`  过期时间
- `maker`  订单发起人地址（签名地址）
  
#### 返回
- `[]hash` `[`区块链上交易hash, 订单hash`]`

#### 示例
```shell
curl -s -X POST http://127.0.0.1:18804 -d '{"jsonrpc":"2.0","method":"wlt_postOrder","params":[{"tokenGet":"0x95ccc08ab44ac6d071a0c5911df64ad2394a4c54", "amountGet":"0x1", "tokenGive":"0x95ccc08ab44ac6d071a0c5911df64ad2394a4c54","amountGive":"0x1", "expires":"0x1", "nonce":"0x1", "maker":"0xa73810e519e1075010678d706533486d8ecc8000"}],"id":67}' -H 'Content-Type:application/json'
```
```
{"jsonrpc":"2.0","id":67,"result":["0x60c449e87f6cff794d5e30e512f30c4189ded4dbcb9996971b586ee3c50eb328","0x1634c76170aea067b75cb54ec770b040542024889ce82e67c4ef2d726328cb43"]}
```

### wlt_trade
成交指定订单
#### 参数
- `address` taker地址
- `order`  订单信息
- `v` 签名
- `s` 签名
- `r` 签名
- `amount` 交易数量
#### 返回
- `[]hash` `[`区块链上交易hash, 订单hash`]`

#### 示例
```shell
curl -s -X POST http://127.0.0.1:18804 -d '{"jsonrpc":"2.0","method":"wlt_trade","params":["0xa73810e519e1075010678d706533486d8ecc8000",{"order":{"tokenGet":"0x95ccc08ab44ac6d071a0c5911df64ad2394a4c54","amountGet":"0x1","tokenGive":"0x95ccc08ab44ac6d071a0c5911df64ad2394a4c54","amountGive":"0x1","expires":"0x5dafc158","nonce":"0x1","maker":"0xa73810e519e1075010678d706533486d8ecc8000"},"v":"0x1b","s":"0x71b0df728b6639f405446be1f98f5ef1bdb1932259456d6b9b4bc3e1bdc19d57","r":"0x809cc318a5eb5a14c9073d0858a98efe268fcf5852d6bc76778147f064f00d5a"},"0x1"],"id":67}' -H 'Content-Type:application/json'
```
```
{"jsonrpc":"2.0","id":67,"result":["0x60c449e87f6cff794d5e30e512f30c4189ded4dbcb9996971b586ee3c50eb328","0x1634c76170aea067b75cb54ec770b040542024889ce82e67c4ef2d726328cb43"]}
```

### wlt_takerOrderByHash
成交指定订单
#### 参数
#### 示例
#### 返回

### wlt_withdrawToken
提取金额至合约
#### 参数
- `address` 账户地址 
- `address` token地址 
- `amount` token数量

#### 返回
- `hash` 交易hash

#### 示例
```shell
curl -s -X POST http://127.0.0.1:18804 -d '{"jsonrpc":"2.0","method":"wlt_withdrawToken","params":["0xa73810e519e1075010678d706533486d8ecc8000","0x95ccc08ab44ac6d071a0c5911df64ad2394a4c54", "0x1"],"id":67}' -H 'Content-Type:application/json'
```
```
{"jsonrpc":"2.0","id":67,"result":"0x58c47fe8c3044dd179db5328853f3b54f572a0177da1fdabbd6dd885d1be7197"}
```

### wlt_depositToken
抵押金额至合约
#### 参数
- `address` 用户地址
- `address` token地址值
- `amount` 抵押数量

#### 返回
- `hash` 交易hash

#### 示例
```shell
curl -s -X POST http://127.0.0.1:18804 -d '{"jsonrpc":"2.0","method":"wlt_depositToken","params":["0xa73810e519e1075010678d706533486d8ecc8000","0x95ccc08ab44ac6d071a0c5911df64ad2394a4c54", "0x1"],"id":67}' -H 'Content-Type:application/json'
```

```
{"jsonrpc":"2.0","id":67,"result":"0x227c50d045ca22ba74ac1a5662812905c4a7d4f194925a131a4130d121d814f4"}
```


### TODO 
- 可视化客户端开发(价格走势、线上未成交订单、实时价格显示)
- 合约增加成交手续费,作为客户端的运营成本收入
- 支持非本地linkchain节点的数据查询,支持节点钱包离线查询
- 移动端客户端开发

