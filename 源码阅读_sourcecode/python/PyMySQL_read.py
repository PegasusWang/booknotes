# -*- coding: utf-8 -*-

"""
代码阅读： 传 python 的 mysql driver，gevent patch 了 socket，需要用纯 python driver
https://github.com/PyMySQL/PyMySQL

目的：了解 python 的 mysql driver 如何工作的

知识点：
- mysql server-client 协议:
    - http://dev.mysql.com/doc/internals/en/client-server-protocol.html
    - https://jin-yang.github.io/post/mysql-protocol.html
    - http://hutaow.com/blog/2013/11/06/mysql-protocol-analysis/

- socket 编程: 看下文档就行，主要是参数有些参数需要参考 unix 网络编程的东西

切入点：从示例 demo 看起，自顶向下看代码。对于有些搞不懂的代码片段，可以通过打断点调试

通信过：
1. 三次握手建立 TCP 连接。

2. 建立 MySQL 连接，也就是认证阶段。
    服务端 -> 客户端：发送握手初始化包 (Handshake Initialization Packet)。
    客户端 -> 服务端：发送验证包 (Client Authentication Packet)。
    服务端 -> 客户端：认证结果消息。

3. 认证通过之后，客户端开始与服务端之间交互，也就是命令执行阶段。
    客户端 -> 服务端：发送命令包 (Command Packet)。
    服务端 -> 客户端：发送回应包 (OK Packet, or Error Packet, or Result Set Packet)。

4. 断开 MySQL 连接。
    客户端 -> 服务器：发送退出命令包。

5. 四次握手断开 TCP 连接。



import pymysql.cursors

# Connect to the database
connection = pymysql.connect(host='localhost',
                             user='user',
                             password='passwd',
                             db='db',
                             charset='utf8mb4',
                             cursorclass=pymysql.cursors.DictCursor)

try:
    with connection.cursor() as cursor:
        # Create a new record
        sql = "INSERT INTO `users` (`email`, `password`) VALUES (%s, %s)"
        cursor.execute(sql, ('webmaster@python.org', 'very-secret'))

    # connection is not autocommit by default. So you must commit to save
    # your changes.
    connection.commit()

    with connection.cursor() as cursor:
        # Read a single record
        sql = "SELECT `id`, `password` FROM `users` WHERE `email`=%s"
        cursor.execute(sql, ('webmaster@python.org',))
        result = cursor.fetchone()
        print(result)
finally:
    connection.close()


主要的是 Connection and Cursor 对象.
Connection:
    Representation of a socket with a mysql server.
    The proper way to get an instance of this class is to call connect().
    Establish a connection to the MySQL database. Accepts several arguments:

Cursor:
    This is the object you use to interact with the database.
    Do not create an instance of a Cursor yourself. Call connections.Connection.cursor().
    See Cursor in the specification.

代码中附录了关于 mysql 的 client-server 协议:
# http://dev.mysql.com/doc/internals/en/client-server-protocol.html
"""


"""
先来看 Connection  对象，仅仅分析关键部分(关键函数和流程)，它的 init 函数涉及到的参数非常多，分析代码的时候可以暂时
忽略掉
"""

class Connection(object):

    """ Connection 对象代表了 和 mysql server 的 socket 连接"""

    _sock = None  # 代表和 mysql server 连接的 socket 对象
    _auth_plugin_name = ''
    _closed = False
    _secure = False

    def __init__(self, host=None, user=None, password="",
                 database=None, port=0, unix_socket=None,
                 charset='', sql_mode=None,
                 read_default_file=None, conv=None, use_unicode=None,
                 client_flag=0, cursorclass=Cursor, init_command=None,
                 connect_timeout=10, ssl=None, read_default_group=None,
                 compress=None, named_pipe=None,
                 autocommit=False, db=None, passwd=None, local_infile=False,
                 max_allowed_packet=16*1024*1024, defer_connect=False,
                 auth_plugin_map=None, read_timeout=None, write_timeout=None,
                 bind_address=None, binary_prefix=False, program_name=None,
                 server_public_key=None):
        if use_unicode is None and sys.version_info[0] > 2:
            use_unicode = True

        if db is not None and database is None:
            database = db
        if passwd is not None and not password:
            password = passwd

        if compress or named_pipe:
            raise NotImplementedError("compress and named_pipe arguments are not supported")

        self._local_infile = bool(local_infile)
        if self._local_infile:
            client_flag |= CLIENT.LOCAL_FILES

        if read_default_group and not read_default_file:
            if sys.platform.startswith("win"):
                read_default_file = "c:\\my.ini"
            else:
                read_default_file = "/etc/my.cnf"

        if read_default_file:
            if not read_default_group:
                read_default_group = "client"

            cfg = Parser()  # 解析配置文件
            cfg.read(os.path.expanduser(read_default_file))

            def _config(key, arg):
                if arg:
                    return arg
                try:
                    return cfg.get(read_default_group, key)
                except Exception:
                    return arg

            user = _config("user", user)
            password = _config("password", password)
            host = _config("host", host)
            database = _config("database", database)
            unix_socket = _config("socket", unix_socket)
            port = int(_config("port", port))
            bind_address = _config("bind-address", bind_address)
            charset = _config("default-character-set", charset)
            if not ssl:
                ssl = {}
            if isinstance(ssl, dict):
                for key in ["ca", "capath", "cert", "key", "cipher"]:
                    value = _config("ssl-" + key, ssl.get(key))
                    if value:
                        ssl[key] = value

        self.ssl = False
        if ssl:
            if not SSL_ENABLED:
                raise NotImplementedError("ssl module not found")
            self.ssl = True
            client_flag |= CLIENT.SSL
            self.ctx = self._create_ssl_ctx(ssl)

        self.host = host or "localhost"
        self.port = port or 3306
        self.user = user or DEFAULT_USER
        self.password = password or b""
        if isinstance(self.password, text_type):
            self.password = self.password.encode('latin1')
        self.db = database
        self.unix_socket = unix_socket
        self.bind_address = bind_address
        if not (0 < connect_timeout <= 31536000):
            raise ValueError("connect_timeout should be >0 and <=31536000")
        self.connect_timeout = connect_timeout or None
        if read_timeout is not None and read_timeout <= 0:
            raise ValueError("read_timeout should be >= 0")
        self._read_timeout = read_timeout
        if write_timeout is not None and write_timeout <= 0:
            raise ValueError("write_timeout should be >= 0")
        self._write_timeout = write_timeout
        if charset:
            self.charset = charset
            self.use_unicode = True
        else:
            self.charset = DEFAULT_CHARSET
            self.use_unicode = False

        if use_unicode is not None:
            self.use_unicode = use_unicode

        self.encoding = charset_by_name(self.charset).encoding

        client_flag |= CLIENT.CAPABILITIES
        if self.db:
            client_flag |= CLIENT.CONNECT_WITH_DB

        self.client_flag = client_flag

        self.cursorclass = cursorclass

        self._result = None
        self._affected_rows = 0
        self.host_info = "Not connected"

        #: specified autocommit mode. None means use server default.
        self.autocommit_mode = autocommit

        if conv is None:
            conv = converters.conversions

        # Need for MySQLdb compatibility.
        self.encoders = dict([(k, v) for (k, v) in conv.items() if type(k) is not int])
        self.decoders = dict([(k, v) for (k, v) in conv.items() if type(k) is int])
        self.sql_mode = sql_mode
        self.init_command = init_command
        self.max_allowed_packet = max_allowed_packet
        self._auth_plugin_map = auth_plugin_map or {}
        self._binary_prefix = binary_prefix
        self.server_public_key = server_public_key

        self._connect_attrs = {
            '_client_name': 'pymysql',
            '_pid': str(os.getpid()),
            '_client_version': VERSION_STRING,
        }
        if program_name:
            self._connect_attrs["program_name"] = program_name
        elif sys.argv:
            self._connect_attrs["program_name"] = sys.argv[0]

        if defer_connect:  # 延迟连接
            self._sock = None
        else:
            self.connect()   # 初始化的时候调用 connect()，下边首先分析 connect 函数

    def connect(self, sock=None): # connect 函数主要是和 mysql server 建立 socket 对象并连接
        self._closed = False
        try:
            if sock is None:
                if self.unix_socket:
                    sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
                    sock.settimeout(self.connect_timeout)
                    sock.connect(self.unix_socket)
                    self.host_info = "Localhost via UNIX socket"
                    self._secure = True
                    if DEBUG: print('connected using unix_socket')
                else:
                    kwargs = {}
                    if self.bind_address is not None:
                        kwargs['source_address'] = (self.bind_address, 0)
                    while True:
                        try:
                            sock = socket.create_connection(
                                (self.host, self.port), self.connect_timeout,
                                **kwargs)
                            break
                        except (OSError, IOError) as e:
                            if e.errno == errno.EINTR:
                                continue
                            raise
                    self.host_info = "socket %s:%d" % (self.host, self.port)
                    if DEBUG: print('connected using socket')
                    sock.setsockopt(socket.IPPROTO_TCP, socket.TCP_NODELAY, 1)
                sock.settimeout(None)
                sock.setsockopt(socket.SOL_SOCKET, socket.SO_KEEPALIVE, 1)
            self._sock = sock
            self._rfile = _makefile(sock, 'rb')
            self._next_seq_id = 0

            self._get_server_information()
            self._request_authentication()    # 发送命令之前的 握手认证 阶段

            if self.sql_mode is not None:
                c = self.cursor()
                c.execute("SET sql_mode=%s", (self.sql_mode,))

            if self.init_command is not None:
                c = self.cursor()    # 注意这里调用了 cursor，下边我们就先看下这个函数
                c.execute(self.init_command)
                c.close()
                self.commit()

            if self.autocommit_mode is not None:
                self.autocommit(self.autocommit_mode)
        except BaseException as e:
            self._rfile = None
            if sock is not None:
                try:
                    sock.close()
                except:  # noqa
                    pass

            if isinstance(e, (OSError, IOError, socket.error)):
                exc = err.OperationalError(
                        2003,
                        "Can't connect to MySQL server on %r (%s)" % (
                            self.host, e))
                # Keep original exception and traceback to investigate error.
                exc.original_exception = e
                exc.traceback = traceback.format_exc()
                if DEBUG: print(exc.traceback)
                raise exc

            # If e is neither DatabaseError or IOError, It's a bug.
            # But raising AssertionError hides original error.
            # So just reraise it.
            raise

    def cursor(self, cursor=None):
        """
        Create a new cursor to execute queries with.

        :param cursor: The type of cursor to create; one of :py:class:`Cursor`,
            :py:class:`SSCursor`, :py:class:`DictCursor`, or :py:class:`SSDictCursor`.
            None means use Cursor.
        """
        if cursor:
            return cursor(self)
        return self.cursorclass(self)

"""
# 了解到 connect 建立 socket 连接之后，我们看下示例 demo 使用方式，用了 with 协议，
# 下边就来分析 __enter__ 和  __exit__ 如何实现它的
# with statement context manager
# The with statement is used to wrap the execution of a block with methods defined by a context manager (see section With Statement Context Managers). This allows common try...except...finally usage patterns to be encapsulated for convenient reuse.

connection = pymysql.connect(host='localhost',
                             user='user',
                             password='passwd',
                             db='db',
                             charset='utf8mb4',
                             cursorclass=pymysql.cursors.DictCursor)

try:
    with connection.cursor() as cursor:
        # Create a new record
        sql = "INSERT INTO `users` (`email`, `password`) VALUES (%s, %s)"
        cursor.execute(sql, ('webmaster@python.org', 'very-secret'))

    # connection is not autocommit by default. So you must commit to save
    # your changes.
    connection.commit()

    with connection.cursor() as cursor:
        # Read a single record
        sql = "SELECT `id`, `password` FROM `users` WHERE `email`=%s"
        cursor.execute(sql, ('webmaster@python.org',))
        result = cursor.fetchone()
        print(result)
finally:
    connection.close()

"""
    def __enter__(self):
        """Context manager that returns a Cursor"""
        return self.cursor()

    def __exit__(self, exc, value, traceback):
        """On successful exit, commit. On exception, rollback"""
        if exc:
            self.rollback()
        else:
            self.commit()  # with 如果没有发生异常，自动执行 commit，下边看下 commit rollback 实现

    # The following methods are INTERNAL USE ONLY (called from Cursor)
    # 这里调用 cursor 的 execute 方法实际上是调用 cursor.connection.query
    def query(self, sql, unbuffered=False):
        # if DEBUG:
        #     print("DEBUG: sending query:", sql)
        if isinstance(sql, text_type) and not (JYTHON or IRONPYTHON):
            if PY2:
                sql = sql.encode(self.encoding)
            else:
                sql = sql.encode(self.encoding, 'surrogateescape')
        self._execute_command(COMMAND.COM_QUERY, sql)
        self._affected_rows = self._read_query_result(unbuffered=unbuffered)
        return self._affected_rows

    def commit(self):
        """
        Commit changes to stable storage.

        See `Connection.commit() <https://www.python.org/dev/peps/pep-0249/#commit>`_
        in the specification.
        """
        self._execute_command(COMMAND.COM_QUERY, "COMMIT")  # 发送 commit 命令
        self._read_ok_packet()

    def _execute_command(self, command, sql):
        """
        :raise InterfaceError: If the connection is closed.
        :raise ValueError: If no username was specified.
        """
        if not self._sock:
            raise err.InterfaceError("(0, '')")

        # If the last query was unbuffered, make sure it finishes before
        # sending new commands
        if self._result is not None:    # 这里的 _result 是个 result = MySQLResult(self) 对象，后边再来看 MySQLResult
            if self._result.unbuffered_active:
                warnings.warn("Previous unbuffered result was left incomplete")
                self._result._finish_unbuffered_query()
            while self._result.has_next:
                self.next_result()
            self._result = None

        if isinstance(sql, text_type):
            sql = sql.encode(self.encoding)

        packet_size = min(MAX_PACKET_LEN, len(sql) + 1)  # +1 is for comman, 单个报文的最大长度为 (2^24-1)Bytes ，也即 (16M-1)Bytes

        # tiny optimization: build first packet manually instead of
        # calling self..write_packet()
        prelude = struct.pack('<iB', packet_size, command)  # 看下 python struct 文档,This module performs conversions between Python values and C structs represented as Python strings.
        packet = prelude + sql[:packet_size-1]
        self._write_bytes(packet)  # 下边我们先分析 _write_bytes 方法，其实就是调用了 socket.sendall
        if DEBUG: dump_packet(packet)    # 这里 dump_packet 在 pymysql.protocl.py 里，实现了低层的 mysql cliet-server 协议
        self._next_seq_id = 1

        if packet_size < MAX_PACKET_LEN:
            return

        sql = sql[packet_size-1:]
        while True:   # 如果超过了最大的 packet 长度，分批发送
            packet_size = min(MAX_PACKET_LEN, len(sql))
            self.write_packet(sql[:packet_size])  # 后边分析 write_packet 方法
            sql = sql[packet_size:]
            if not sql and packet_size < MAX_PACKET_LEN:
                break

    """
    上边代码涉及到了一些 和 mysql server 交互的部分，比如 发送 packet，client-server 协议等，可能有些地方看不太懂
    先阅读一些材料方便理解，这里其实并不难，只不过涉及到协议解析的部分稍微麻烦一些：
    http://hutaow.com/blog/2013/11/06/mysql-protocol-analysis/
    https://jin-yang.github.io/post/mysql-protocol.html
    https://dev.mysql.com/doc/internals/en/string.html
    """
    def _write_bytes(self, data):
        self._sock.settimeout(self._write_timeout)
        try:
            self._sock.sendall(data)
        except IOError as e:
            self._force_close()
            raise err.OperationalError(
                CR.CR_SERVER_GONE_ERROR,
                "MySQL server has gone away (%r)" % (e,))

    def write_packet(self, payload):
        """Writes an entire "mysql packet" in its entirety to the network
        addings its length and sequence number.
        """
        # Internal note: when you build packet manualy and calls _write_bytes()
        # directly, you should set self._next_seq_id properly.
        def pack_int24(n):
            # 看 strut 模块文档 '<' 代表 little-endian(小端法), 'I' 是 unsigned int
            return struct.pack('<I', n)[:3] # MySQL 报文中整型值分别有 1、2、3、4、8 字节长度，使用小字节序传输。
        def int2byte(i):
            return struct.pack("!B", i)
        # 报文分为消息头和消息体两部分，其中消息头占用固定的4个字节，消息体长度由消息头中的长度字段决定，报文结构如下：
        # 消息长度 + 序号 + 报文数据， 参考 https://jin-yang.github.io/post/mysql-protocol.html
        # +-------------------+--------------+---------------------------------------------------+
        # |      3 Bytes      |    1 Byte    |                   N Bytes                         |
        # +-------------------+--------------+---------------------------------------------------+
        # |<= length of msg =>|<= sequence =>|<==================== data =======================>|
        # |<============= header ===========>|<==================== body =======================>|
        data = pack_int24(len(payload)) + int2byte(self._next_seq_id) + payload
        if DEBUG: dump_packet(data)
        self._write_bytes(data)   # 直接 socket sendall
        self._next_seq_id = (self._next_seq_id + 1) % 256

    def _read_ok_packet(self): # self.commit 第二步
        pkt = self._read_packet()
        if not pkt.is_ok_packet():
            raise err.OperationalError(2014, "Command Out of Sync")
        ok = OKPacketWrapper(pkt)
        self.server_status = ok.server_status
        return ok

    def _read_packet(self, packet_type=MysqlPacket):
        """Read an entire "mysql packet" in its entirety from the network
        and return a MysqlPacket type that represents the results.

        :raise OperationalError: If the connection to the MySQL server is lost.
        :raise InternalError: If the packet sequence number is wrong.
        """
        buff = b''
        while True:
            packet_header = self._read_bytes(4)  # 从 socket 读取 4 bytes 数据
            #if DEBUG: dump_packet(packet_header)

            btrl, btrh, packet_number = struct.unpack('<HBB', packet_header)
            bytes_to_read = btrl + (btrh << 16)
            if packet_number != self._next_seq_id:
                self._force_close()
                if packet_number == 0:
                    # MariaDB sends error packet with seqno==0 when shutdown
                    raise err.OperationalError(
                        CR.CR_SERVER_LOST,
                        "Lost connection to MySQL server during query")
                raise err.InternalError(
                    "Packet sequence number wrong - got %d expected %d"
                    % (packet_number, self._next_seq_id))
            self._next_seq_id = (self._next_seq_id + 1) % 256

            recv_data = self._read_bytes(bytes_to_read)
            if DEBUG: dump_packet(recv_data)
            buff += recv_data
            # https://dev.mysql.com/doc/internals/en/sending-more-than-16mbyte.html
            if bytes_to_read == 0xffffff:
                continue
            if bytes_to_read < MAX_PACKET_LEN:
                break

        packet = packet_type(buff, self.encoding)
        packet.check_error()
        return packet

    def _read_bytes(self, num_bytes):
        self._sock.settimeout(self._read_timeout)
        while True:
            try:
                data = self._rfile.read(num_bytes)
                break
            except (IOError, OSError) as e:
                if e.errno == errno.EINTR:
                    continue
                self._force_close()
                raise err.OperationalError(
                    CR.CR_SERVER_LOST,
                    "Lost connection to MySQL server during query (%s)" % (e,))
        if len(data) < num_bytes:
            self._force_close()
            raise err.OperationalError(
                CR.CR_SERVER_LOST, "Lost connection to MySQL server during query")
        return data


"""
接下来是 Cursor 对象，cursor 负责和数据库交互，接受一个 Connection 对象作为参数
关键方法是 execute and fetchone() fetchall()
"""

class Cursor(object):
    """
    This is the object you use to interact with the database.

    Do not create an instance of a Cursor yourself. Call
    connections.Connection.cursor().

    See `Cursor <https://www.python.org/dev/peps/pep-0249/#cursor-objects>`_ in
    the specification.
    """

    #: Max statement size which :meth:`executemany` generates.
    #:
    #: Max size of allowed statement is max_allowed_packet - packet_header_size.
    #: Default value of max_allowed_packet is 1048576.
    max_stmt_length = 1024000

    _defer_warnings = False

    def __init__(self, connection):
        self.connection = connection
        self.description = None
        self.rownumber = 0
        self.rowcount = -1
        self.arraysize = 1
        self._executed = None
        self._result = None
        self._rows = None
        self._warnings_handled = False

    def execute(self, query, args=None): # 先从 execute 方法看起
        """Execute a query

        :param str query: Query to execute.

        :param args: parameters used with query. (optional)
        :type args: tuple, list or dict

        :return: Number of affected rows
        :rtype: int

        If args is a list or tuple, %s can be used as a placeholder in the query.
        If args is a dict, %(name)s can be used as a placeholder in the query.
        """
        while self.nextset():
            pass

        query = self.mogrify(query, args)

        result = self._query(query)
        self._executed = query
        return result

    def _query(self, q):
        conn = self._get_db()   # 获取 connection 对象
        self._last_executed = q
        self._clear_result()  # 相关属性设置为空
        conn.query(q)  # 调用 connection 对象的 query 方法, connection.query 执行了其 _execute_command
        self._do_get_result()  # 获取 connection 的 result 值, MySQLResult 对象
        return self.rowcount

    ########## _query 调用的方法 beg:##########
    def _get_db(self):
        if not self.connection:
            raise err.ProgrammingError("Cursor closed")
        return self.connection

    def _clear_result(self):
        self.rownumber = 0
        self._result = None

        self.rowcount = 0
        self.description = None
        self.lastrowid = None
        self._rows = None

    def _do_get_result(self):
        conn = self._get_db()

        self._result = result = conn._result

        self.rowcount = result.affected_rows
        self.description = result.description
        self.lastrowid = result.insert_id
        self._rows = result.rows
        self._warnings_handled = False

        if not self._defer_warnings:
            self._show_warnings()
    ########## _query 调用的方法 :end ##########

    """执行完了 execute 之后调用的的是 fetchone() or fetchall()"""
    def fetchone(self):
        """Fetch next row"""
        self._check_executed() # 检查 if not self._executed:
        row = self.read_next()
        if row is None:
            self._show_warnings()
            return None
        self.rownumber += 1
        return row

    def read_next(self):
        """Read next row"""
        def _conv_row(self, row):  # 会被子类覆写用来修改返回的类型
            return row
        # MySQLResult 的 _read_rowdata_packet_unbuffered，下边来看看 MySQLResult 的代码
        return self._conv_row(self._result._read_rowdata_packet_unbuffered())


"""
最后是 MySQLResult 类，可以看到 cursor.read_next 就是直接调用的 MySQLResult._read_rowdata_packet_unbuffered
"""
class MySQLResult(object):

    def __init__(self, connection):  # 依然是 Connection 对象
        """
        :type connection: Connection
        """
        self.connection = connection
        self.affected_rows = None
        self.insert_id = None
        self.server_status = None
        self.warning_count = 0
        self.message = None
        self.field_count = 0
        self.description = None
        self.rows = None
        self.has_next = None
        self.unbuffered_active = False

    def __del__(self):
        # 注意py 里的 del魔术方法： del x doesn’t directly call x.__del__() — the former decrements the reference count for x by one, and the latter is only called when x‘s reference count reaches zero<Paste>
        if self.unbuffered_active:
            self._finish_unbuffered_query()

    def read(self):
        """

        3. 认证通过之后，客户端开始与服务端之间交互，也就是命令执行阶段。
        客户端 -> 服务端：发送命令包 (Command Packet)。
        服务端 -> 客户端：发送回应包 (OK Packet, or Error Packet, or Result Set Packet)。
        """
        try:
            first_packet = self.connection._read_packet()
            # 处理 mysql 返回的不同类型回应报文
            if first_packet.is_ok_packet():
                self._read_ok_packet(first_packet)
            elif first_packet.is_load_local_packet():
                self._read_load_local_packet(first_packet)
            else:
                self._read_result_packet(first_packet)
        finally:
            self.connection = None

    def init_unbuffered_query(self):
        """
        :raise OperationalError: If the connection to the MySQL server is lost.
        :raise InternalError:
        """
        self.unbuffered_active = True
        first_packet = self.connection._read_packet()

        if first_packet.is_ok_packet():
            self._read_ok_packet(first_packet)
            self.unbuffered_active = False
            self.connection = None
        elif first_packet.is_load_local_packet():
            self._read_load_local_packet(first_packet)
            self.unbuffered_active = False
            self.connection = None
        else:
            self.field_count = first_packet.read_length_encoded_integer()
            self._get_descriptions()

            # Apparently, MySQLdb picks this number because it's the maximum
            # value of a 64bit unsigned integer. Since we're emulating MySQLdb,
            # we set it to this instead of None, which would be preferred.
            self.affected_rows = 18446744073709551615

    def _read_ok_packet(self, first_packet):
        ok_packet = OKPacketWrapper(first_packet)
        self.affected_rows = ok_packet.affected_rows
        self.insert_id = ok_packet.insert_id
        self.server_status = ok_packet.server_status
        self.warning_count = ok_packet.warning_count
        self.message = ok_packet.message
        self.has_next = ok_packet.has_next

    def _read_load_local_packet(self, first_packet):
        if not self.connection._local_infile:
            raise RuntimeError(
                "**WARN**: Received LOAD_LOCAL packet but local_infile option is false.")
        load_packet = LoadLocalPacketWrapper(first_packet)
        sender = LoadLocalFile(load_packet.filename, self.connection)
        try:
            sender.send_data()
        except:
            self.connection._read_packet()  # skip ok packet
            raise

        ok_packet = self.connection._read_packet()
        if not ok_packet.is_ok_packet(): # pragma: no cover - upstream induced protocol error
            raise err.OperationalError(2014, "Commands Out of Sync")
        self._read_ok_packet(ok_packet)

    def _check_packet_is_eof(self, packet):
        if not packet.is_eof_packet():
            return False
        #TODO: Support CLIENT.DEPRECATE_EOF
        # 1) Add DEPRECATE_EOF to CAPABILITIES
        # 2) Mask CAPABILITIES with server_capabilities
        # 3) if server_capabilities & CLIENT.DEPRECATE_EOF: use OKPacketWrapper instead of EOFPacketWrapper
        wp = EOFPacketWrapper(packet)
        self.warning_count = wp.warning_count
        self.has_next = wp.has_next
        return True

    def _read_result_packet(self, first_packet):
        self.field_count = first_packet.read_length_encoded_integer()
        self._get_descriptions()
        self._read_rowdata_packet()

    def _read_rowdata_packet_unbuffered(self):
        # Check if in an active query
        if not self.unbuffered_active:
            return

        # EOF
        packet = self.connection._read_packet()
        if self._check_packet_is_eof(packet):
            self.unbuffered_active = False
            self.connection = None
            self.rows = None
            return

        row = self._read_row_from_packet(packet)
        self.affected_rows = 1
        self.rows = (row,)  # rows should tuple of row for MySQL-python compatibility.
        return row

    def _finish_unbuffered_query(self):
        # After much reading on the MySQL protocol, it appears that there is,
        # in fact, no way to stop MySQL from sending all the data after
        # executing a query, so we just spin, and wait for an EOF packet.
        while self.unbuffered_active:
            packet = self.connection._read_packet()
            if self._check_packet_is_eof(packet):
                self.unbuffered_active = False
                self.connection = None  # release reference to kill cyclic reference.

    def _read_rowdata_packet(self):
        """Read a rowdata packet for each data row in the result set."""
        rows = []
        while True:
            packet = self.connection._read_packet()
            if self._check_packet_is_eof(packet):
                self.connection = None  # release reference to kill cyclic reference.
                break
            rows.append(self._read_row_from_packet(packet))

        self.affected_rows = len(rows)
        self.rows = tuple(rows)

    def _read_row_from_packet(self, packet):
        row = []
        for encoding, converter in self.converters:
            try:
                data = packet.read_length_coded_string()
            except IndexError:
                # No more columns in this row
                # See https://github.com/PyMySQL/PyMySQL/pull/434
                break
            if data is not None:
                if encoding is not None:
                    data = data.decode(encoding)
                if DEBUG: print("DEBUG: DATA = ", data)
                if converter is not None:
                    data = converter(data)
            row.append(data)
        return tuple(row)

    def _get_descriptions(self):
        """Read a column descriptor packet for each column in the result."""
        self.fields = []
        self.converters = []
        use_unicode = self.connection.use_unicode
        conn_encoding = self.connection.encoding
        description = []

        for i in range_type(self.field_count):
            field = self.connection._read_packet(FieldDescriptorPacket)
            self.fields.append(field)
            description.append(field.description())
            field_type = field.type_code
            if use_unicode:
                if field_type == FIELD_TYPE.JSON:
                    # When SELECT from JSON column: charset = binary
                    # When SELECT CAST(... AS JSON): charset = connection encoding
                    # This behavior is different from TEXT / BLOB.
                    # We should decode result by connection encoding regardless charsetnr.
                    # See https://github.com/PyMySQL/PyMySQL/issues/488
                    encoding = conn_encoding  # SELECT CAST(... AS JSON)
                elif field_type in TEXT_TYPES:
                    if field.charsetnr == 63:  # binary
                        # TEXTs with charset=binary means BINARY types.
                        encoding = None
                    else:
                        encoding = conn_encoding
                else:
                    # Integers, Dates and Times, and other basic data is encoded in ascii
                    encoding = 'ascii'
            else:
                encoding = None
            converter = self.connection.decoders.get(field_type)
            if converter is converters.through:
                converter = None
            if DEBUG: print("DEBUG: field={}, converter={}".format(field, converter))
            self.converters.append((encoding, converter))

        eof_packet = self.connection._read_packet()
        assert eof_packet.is_eof_packet(), 'Protocol error, expecting EOF'
        self.description = tuple(description)

