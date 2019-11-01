# 编译
项目目录下执行`make`

# 使用说明
使用lkdex需要启用linkchain节点以及钱包节点，具体见[linkchain](https://github.com/lianxiangcloud/linkchain)项目

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

## wallet相关接口（钱包解锁账户）
### 错误说明
接口调用出错可能
1.合约执行失败
2.钱包没有解锁对应账户
3.参数错误
4.钱包没有连接到节点

### wlt_signOrder
#### 参数
- `tokenGet` 需要交换的token地址
- `amountGet`  交换token数量
- `tokenGive`  交换的token地址
- `amountGive`  交换的token数量
- `nonce`  交易nonce值
- `expire`  过期时间
- `user`  订单发起人地址（签名地址）

#### 返回
- `order` 订单信息
- `r` 签名
- `s` 签名 
- `v` 签名

#### 示例
```shell
curl -s -X POST http://127.0.0.1:18804 -d '{"jsonrpc":"2.0","method":"wlt_signOrder","params":[{"tokenGet":"0x95ccc08ab44ac6d071a0c5911df64ad2394a4c54", "amountGet":"0x1", "tokenGive":"0x95ccc08ab44ac6d071a0c5911df64ad2394a4c54","amountGive":"0x1", "expires":"0x1", "nonce":"0x1", "user":"0xa73810e519e1075010678d706533486d8ecc8000"}],"id":67}' -H 'Content-Type:application/json'
```
```
{"jsonrpc":"2.0","id":67,"result":{"order":{"tokenGet":"0x95ccc08ab44ac6d071a0c5911df64ad2394a4c54","amountGet":"0x1","tokenGive":"0x95ccc08ab44ac6d071a0c5911df64ad2394a4c54","amountGive":"0x1","expires":"0x5dafc158","nonce":"0x1","user":"0xa73810e519e1075010678d706533486d8ecc8000"},"v":"0x1b","s":"0x71b0df728b6639f405446be1f98f5ef1bdb1932259456d6b9b4bc3e1bdc19d57","r":"0x809cc318a5eb5a14c9073d0858a98efe268fcf5852d6bc76778147f064f00d5a"}}
```

### wlt_postOrder
#### 参数
- `tokenGet` 需要交换的token地址
- `amountGet`  交换token数量
- `tokenGive`  交换的token地址
- `amountGive`  交换的token数量
- `nonce`  交易nonce值
- `expire`  过期时间
- `user`  订单发起人地址（签名地址）
  
#### 返回
- `[]hash` `[`区块链上交易hash, 订单hash`]`

#### 示例
```shell
curl -s -X POST http://127.0.0.1:18804 -d '{"jsonrpc":"2.0","method":"wlt_postOrder","params":[{"tokenGet":"0x95ccc08ab44ac6d071a0c5911df64ad2394a4c54", "amountGet":"0x1", "tokenGive":"0x95ccc08ab44ac6d071a0c5911df64ad2394a4c54","amountGive":"0x1", "expires":"0x1", "nonce":"0x1", "user":"0xa73810e519e1075010678d706533486d8ecc8000"}],"id":67}' -H 'Content-Type:application/json'
```
```
{"jsonrpc":"2.0","id":67,"result":["0x60c449e87f6cff794d5e30e512f30c4189ded4dbcb9996971b586ee3c50eb328","0x1634c76170aea067b75cb54ec770b040542024889ce82e67c4ef2d726328cb43"]}
```

### wlt_trade
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
curl -s -X POST http://127.0.0.1:18804 -d '{"jsonrpc":"2.0","method":"wlt_trade","params":["0xa73810e519e1075010678d706533486d8ecc8000",{"order":{"tokenGet":"0x95ccc08ab44ac6d071a0c5911df64ad2394a4c54","amountGet":"0x1","tokenGive":"0x95ccc08ab44ac6d071a0c5911df64ad2394a4c54","amountGive":"0x1","expires":"0x5dafc158","nonce":"0x1","user":"0xa73810e519e1075010678d706533486d8ecc8000"},"v":"0x1b","s":"0x71b0df728b6639f405446be1f98f5ef1bdb1932259456d6b9b4bc3e1bdc19d57","r":"0x809cc318a5eb5a14c9073d0858a98efe268fcf5852d6bc76778147f064f00d5a"},"0x1"],"id":67}' -H 'Content-Type:application/json'
```
```
{"jsonrpc":"2.0","id":67,"result":["0x60c449e87f6cff794d5e30e512f30c4189ded4dbcb9996971b586ee3c50eb328","0x1634c76170aea067b75cb54ec770b040542024889ce82e67c4ef2d726328cb43"]}
```

### wlt_takerOrderByHash
#### 参数
#### 示例
#### 返回

### wlt_withdrawToken
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
