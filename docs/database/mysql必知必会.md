检索：

```
select column from table;
select * from table;
select distinct column from table;
select column from table limit 5;
select column from table 5,5;   #  从行5开始的5行
select column from table order by column;
select column from table order by column, column2;
select column from table order by column desc;
select column from table order by column desc, column2;
select column from table order by column desc limit 1;
select name, price from products where price=2;
select name, price from products where price is null;
select name, price from products where price > 3 and price < 5;
select name, price from products where price in (3,5);
select name, price from products where price not in (3,5);
select Concat(vend_name, '(', vend_country, ')') from vendors order by vend_name;    # 计算字段
```

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
