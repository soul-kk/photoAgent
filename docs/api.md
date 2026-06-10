# 服务器信息

- 服务器端口：http://8.139.5.99:31335

# 注册

## curl请求

```
curl --location --request POST 'http://8.139.5.99:31335/api/auth/register' \
--header 'Content-Type: application/json' \
--data-raw '{
    "username": "demo",
    "email": "demo@example.com",
    "password": "12345678"
}'
```

## 响应

```
{
    "code": 200,
    "err_code": 200,
    "data": {
        "email": "demo@example.com",
        "id": 1,
        "role": "user",
        "username": "demo"
    },
    "message": "注册成功"
}
```

# 登录

## curl请求

```
curl --location --request POST 'http://8.139.5.99:31335/api/auth/login' \
--header 'Content-Type: application/json' \
--data-raw '{
    "account": "demo@example.com",
    "password": "12345678"
}'
```

## 响应

```
{
    "code": 200,
    "err_code": 200,
    "data": {
        "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6ImRlbW8iLCJyb2xlIjoidXNlciIsImV4cCI6MTc4MTEzODg2NywibmJmIjoxNzgxMDUyNDY3LCJpYXQiOjE3ODEwNTI0Njd9.uX_ThCOSaDia9yzyB79PjPUR3bynAFAhBYxps_MolIw",
        "expires_in": 86400,
        "token_type": "Bearer",
        "user": {
            "email": "demo@example.com",
            "id": 1,
            "role": "user",
            "username": "demo"
        }
    },
    "message": "登录成功"
}
```
