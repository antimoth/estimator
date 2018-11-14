## 依据selfconf下的配置模版配置自己的环境，然后使用docker打包获得镜像，直接运行镜像即可。
## 数据库数据格式如下：
```
{
	"_id" : ObjectId("5bea983b97e4d77ec42ec373"),
	"id" : "ethereum_estimated_gas_price",
	"pricelist" : [
		{
			"latency" : NumberLong(600),
			"price" : NumberLong("3142857144")
		},
		{
			"latency" : NumberLong(1800),
			"price" : NumberLong("2478693182")
		},
		{
			"latency" : NumberLong(7200),
			"price" : NumberLong("2465942029")
		},
		{
			"latency" : NumberLong(43200),
			"price" : NumberLong(2000000000)
		},
		{
			"latency" : NumberLong(86400),
			"price" : NumberLong(1600000000)
		}
	]
}
```
