## 4 检索：

```
select column from table;
select * from table;
select distinct column from table;
select column from table limit 5;
select column from table 5,5;   #  从行5开始的5行
```

## 5 排序检索数据

默认以数据底层表中出现的顺序展示，不应该假定检索出来的顺序有意义。
```
select column from table order by column;   # 使用非检索的列排序也是合法的
select column from table order by column, column2;  # 多列排序，只有当第一列不同时，才会使用第二列
select column from table order by column desc;  # 降序排列
select column from table order by column desc, column2; # desc 直接作用它前边的列，多列降序需要每个列都指定 desc
select column from table order by column desc limit 1; # limit 放最后，这里找到最大的值
```


## 6 过滤数据

where子句操作符： = , != , <, <=, >, >=, BETWEEN

```
select name, price from products where price=2;
select name, price from products where price is null;  # 空值 NULL 检查
select name, price from products where price BETWEEN 3 and 5;  # 找到3和5之间，包括3和5
```


## 7 数据过滤
组合 where 子句。注意 AND 优先级高于 OR，如果必要应该使用括号，尽量都用括号防止歧义。

`select prod_name, prod_price from products where (vend_id=1002 or vend_id=1003) and prod_price>=10`

```
select name, price from products where price > 3 and price < 5;
select name, price from products where price in (3,5);  # in 比使用 or 更快
select name, price from products where price not in (3,5);  # 使用 not 否定条件
select Concat(vend_name, '(', vend_country, ')') from vendors order by vend_name;    # 计算字段
```


## 8 用通配符过滤

LIKE操作符(谓词)。

```
# %表示任何字符出现任意次数，不会匹配 NULL
select prod_id,prod_name from products where prod_name LIKE 'jet%';   # 注意尾空格可能会 干扰通配符
select prod_id,prod_name from products where prod_name LIKE '%anvil%';

# _ 只能匹配单个字符
select prod_id,prod_name from products where prod_name LIKE '% ton anvil';
```

注意：不要过度使用通配符；尽量不要把通配符放在开始位置；


## 9 使用正则表达式搜索

mysql 支持正则表达式的子集 REGEXP, LIKE匹配整串

```
# 搜索prod_name 包含文本 1000 的所有行
select prod_id,prod_name from products where prod_name REGEXP '1000' order by prod_name;
select prod_id,prod_name from products where prod_name REGEXP '1000|2000' order by prod_name; # 搜索两个串之一
```

## 10 创建计算字段

拼接字段： concat()

```
select concat(vend_name, ' (', vend_country, ')') from vendors order by vend_name;
select concat(vend_name, ' (', RTrim(vend_country), ')') from vendors order by vend_name;  # RTrim/LTrim/TRim 去除空格
```

别名：使用 as 支持列 别名

```
select Concat(RTrim(vend_name), ' (', RTrim(vend_country), ')') AS vend_title from vendors order by vend_name;
```

算数计算：对检索出的数据进行算数计算


```
select prod_id,quantity,item_price,quantity*item_price AS expanded_price from orderitems where order_num=20005;
```


## 11 使用数据处理函数

文本、数值计算、日期处理、系统函数等

```
# 文本：Left, Length, Locate, Lower, LTrim, Right, RTrim, Soundex(替换为描述语音表示的字母数字模式), SubString, Upper
select vned_name,Upper(vend_name) as vend_name_upcase from vendors order by vend_name;

# 日期：CurDate, Date, Day, Hour, Minute, Month, Now, Second, Time, Year
select cust_id,order_num from orders where order_date ='2005-09-01';
select cust_id,order_num from orders where Date(order_date) ='2005-09-01';  # use Date，更准确
select cust_id,order_num from orders where Date(order_date) BETWEEN '2005-09-01' and '2005-09-30';
select cust_id,order_num from orders where Year(order_date) =2005 and Month(order_date) =9;

# 数值：Abs, Cos, Exp, Mod, Pi, Rand, Sin, Sqrt, Tan
```


## 12 汇总数据

汇总而不是检索数据，确定行数、获取和、找出最大最小平均值。


五个聚集函数（运行在行组上，计算和返回单个值的函数）: avg, count, max, min, sum

```
# avg
select avg(price) as avg_price from products;  # avg会忽略列值为 NULL 的行

# count
select count(*) as num_cust from customers;   # count(*)对表中的行数计算，不管包含的是 NULL 还是非空
select count(cust_email) as num_cust from customers;   # count(column) 忽略 NULL 的值

# max、min, 忽略 NULL 值
select max(prod_price) as max_price FROM products;

# sum
select sum(quantity) as items_ordered from orderitems where order_num = 20005; # ignore NULL

# distince
select avg(quantity) as items_ordered from orderitems where order_num = 20005; # ignore NULL
```

## 13 分组数据
group by and having，分组允许把数据分为多个逻辑组，以便能够对每个组进行聚集计算。

```
# 分组
select vend_id,count(*) as num_prods from products group by vend_id;
# 使用 having 过滤分组，where 过滤行，having 支持所有的where子句条件
select cust_id, count(*) as orders from orders group by cust_id having count(*)>=2;
# having and where 一起用
select cust_id, count(*) as orders from orders where prod_price>=10 group by cust_id having count(*)>=2;

# order by and group by
select order_num, sum(quantity*item_price) as ordertotal from orderitems group by order_num
having sum(quantity*item_price)>=50
order by ordertotal;
```

## 14 使用子查询


子查询： 在SELECT语句中，子查询总是从内向外处理。

```
# 利用子查询进行过滤。可以把一条 select 语句返回的结果用于另一条 select 语句的 where 子句
select cust_name, cust_contact
from customers
where cust_id in (select cust_id
                  from orders
                  where order_num in (select
                      order_num from orderitems where prod_id='TNT2')); # 参考15章使用join 处理


# 作为计算字段使用子查询，相关子查询需要限定列名
select cust_name, cust_state, (select count(*) from orders where orders.cust_id=customers.cust_id) as orders
from customers order by cust_name;
```


## 15 联结表

```
# 引用的列可能出现二义性时，必须使用完全限定列名
select vend_name, prod_name, prod_price from vendors, products where vendors.vend_id=products.vend_id order by vend_name,prod_name;
# 内部联结（等值联结）
select vend_name, prod_name, prod_price from vendors INNER JOIN products on vendors.vend_id = products.vend_id;
# 连接多个表，sql 对一条 select 中的连接的表数目没有限制。先列出所有表，然后定义表之间的关系
select prod_name,vend_name,prod_price,quantity

# 14章的例子使用 join 处理
select cust_name,cust_contact, from customers,orders,orderitems
where customers.cust_id=orders.cust_id and orderitems.order_num=orders.order_num and prod_id='TNT2';
```


## 16 创建高级联结

何时使用表别名？允许单条 select 中多次引用相同的表

自连接：用 as 语句别名

`select p1.prod_id,p1.prod_name from products as p1, products as p2 where p1.vend_id=p2.vend_id and p2.prod_id='DTNTR';`

外部联结：联结包含了那些在相关表中没有关联行的行。这种类型的联结称为外部联结。
与内部联结关联两个表中的行不同的是，外部联结还包括没有关联行的行。在使用OUTER JOIN语法时，必须使用RIGHT或LEFT关键字指定包括其所有行的表（RIGHT指出的是OUTER JOIN右边的表，而LEFT指出的是OUTER JOIN左边的表）。上面的例子使用LEFT OUTER JOIN从FROM子句的左边表（customers表）中选择所有行。为

复合查询：
多数SQL查询都只包含从一个或多个表中返回数据的单条SELECT语句。MySQL也允许执行多个查询（多条SELECT语句），并将结果作为单个查询结果集返回。这些组合查询通常称为并（union）或复合查询（compound query）。
也可以用 or 条件实现相同功能。简化复杂 where


## 17 组合查询
可以用 union 操作符来组合多个 SQL 查询，把结果合并成单个结果集。使用 union 可以使用多个 where 条件替换。

```
# union 必须是相同的列，并且返回的是不重复的行。可以使用 union all 返回所有的行(这个 where 无法完成)
select vend_id,prod_id,prod_price from products where prod_price<=5 union
select vend_id,prod_id,prod_price from products wehre vend_id in (1002,1002);
```
