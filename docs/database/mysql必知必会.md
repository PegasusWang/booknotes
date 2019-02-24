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

聚集函数:
avg, count, max, min, sum

```
select avg(price) as avg_price from products;
```

COUNT()函数有两种使用方式。
□　使用COUNT(*)对表中行的数目进行计数，不管表列中包含的是空值（NULL）还是非空值。
□　使用COUNT(column)对特定列中具有值的行进行计数，忽略NULL值。

max 函数会忽略 NULL 行

分组数据:

select vend_id, count(*) as num_prods from products group by vend_id;    # 对每个组而不是结果聚集
GROUP BY子句必须出现在WHERE子句之后，ORDER BY子句之前。
过滤分组 having
必须基于完整的分组而不是个别的行进行过滤，where 指定的是 行 而不是分组
唯一的差别是WHERE过滤行，而HAVING过滤分组，并且having 支持所有的where操作符

select cust_id, count(*) as orders from orders group by cust_id having count(2) >= 2;

这里有另一种理解方法，WHERE在数据分组前进行过滤，HAVING在数据分组后进行过滤
不要忘记ORDER BY　一般在使用GROUP BY子句时，应该也给出ORDER BY子句。这是保证数据正确排序的唯一方法。千万不要仅依赖GROUP BY排序数据。

子查询：
在SELECT语句中，子查询总是从内向外处理。

Join:
创建连接:(内连接或者等值连接)
select vend_name, prod_name, prod_price
from vendors, products
where vendors.vend_id = products.vend_id
order vend_name, prod_name;

目前为止所用的联结称为等值联结（equijoin），它基于两个表之间的相等测试。这种联结也称为内部联结。其实，对于这种联结可以使用稍微不同的语法来明确指定联结的类型
select vend_name, prod_name, prod_price
from vendors inner join products
on vendors.vend_id = products.vend_id;

join 多个表：
select prod_name, vend_name, prod_price, quantity
from orderitems, products, vendor
where products.vend_id = vend_id
and orderitems.prod_id = products.prod_id
and order_num = 20005;

创建高级 join：
自连接：用 as 语句别名

外部联结：联结包含了那些在相关表中没有关联行的行。这种类型的联结称为外部联结。
与内部联结关联两个表中的行不同的是，外部联结还包括没有关联行的行。在使用OUTER JOIN语法时，必须使用RIGHT或LEFT关键字指定包括其所有行的表（RIGHT指出的是OUTER JOIN右边的表，而LEFT指出的是OUTER JOIN左边的表）。上面的例子使用LEFT OUTER JOIN从FROM子句的左边表（customers表）中选择所有行。为

复合查询：
多数SQL查询都只包含从一个或多个表中返回数据的单条SELECT语句。MySQL也允许执行多个查询（多条SELECT语句），并将结果作为单个查询结果集返回。这些组合查询通常称为并（union）或复合查询（compound query）。
也可以用 or 条件实现相同功能。简化复杂 where
