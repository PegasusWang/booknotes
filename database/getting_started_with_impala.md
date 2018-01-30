# 1. Why Impala

flexibliity for your big data flow
high-performance analytics
exploratory business intelligence

extrac-transform-load(ETL)
BI: Business intelligence

# 2. Getting Up and Running with Impala

Cloudera live demo

A view is an alias for a longer query, and takes no time or storage to set up

# 3. Impala for the database developer

OLTP-style(online transaction processing)

impala implements SQL-92 standard features for queries, with som enhancements from later SQL standards
Hadoop Distributed File System(HDFS)

- Impala currently doesn't have OLTP-style data manipulation language (DML) such as DELETE or UPDATE.
- Impala also does not have indexes, constraints or foreign keys.
- No transactions

impala can very effeciently perform full table scans of large tables.


HDFS Storage Model:
CDH: Cloudera Distribution with Hadoop
Parquet File Format: binary file format

# 4. Common Developer Tasks for Impala

ETL(Extract-trnasform-load)

Make sure always close query handles when finished(release memory)
JDBC or ODBC

with Impala, the biggest I/O savings com from using partitioned tables and choosing the most appropriate file format

Impala partitioned tables are just HDFS directories
UDF(user defined functions)
