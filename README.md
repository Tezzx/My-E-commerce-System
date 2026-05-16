# order-payment-system
1.对数据的操作进行优化，比如高并发扣库存用lua
2.登录注册功能之后也可以用redis优化
3.创建订单的redis和mq两部分还要改成原子化
----------user_model应该加用户编号