# 服务器信息

- 服务器端口：http://8.139.5.99:31335

# AI摄影分析-历史记录

## curl请求

curl --location --request GET 'http://8.139.5.99:31335/api/history/analysis' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6ImRlbW8iLCJyb2xlIjoidXNlciIsImV4cCI6MTc4MjIwMDI0NywibmJmIjoxNzgyMTEzODQ3LCJpYXQiOjE3ODIxMTM4NDd9.2R4e4W1oA70RcWfzxdH2T2neht63g2XCtDUdttuVTvw'

## 响应

```json
{
  "code": 200,
  "err_code": 200,
  "data": {
    "limit": 20,
    "list": [
      {
        "id": 2,
        "user_id": 1,
        "analysis_type": "score",
        "input_prompt": "",
        "focus_dimension": "",
        "result_json": "{\"dimension_scores\":{\"composition\":70,\"color\":73,\"exposure\":62,\"content\":60},\"dimension_notes\":{\"composition\":\"斜向排列形成纵深引导线，前景虚化增加层次，但画面稍显拥挤、主体焦点不够突出。\",\"color\":\"暖橙灯光与蓝绿车身形成冷暖对比，夜景氛围浓郁，但暗部色彩信息有所损失。\",\"exposure\":\"明暗反差较大，高光部分保留尚可，但暗部区域欠曝明显、细节不足。\",\"content\":\"共享单车停放场景记录城市日常，主体类型明确，但叙事性与情感张力较弱。\"},\"overall_analysis\":\"这张照片以低角度近距离拍摄了夜晚停放整齐的共享单车，利用车辆本身的排列构建了斜向延伸的引导线，配合前景虚化营造出一定的空间纵深感。色彩上，暖色调的环境灯光与单车本身的冷色（蓝、绿）形成对比，赋予了画面夜晚城市的氛围感。然而，曝光控制存在明显不足：暗部区域大面积死黑，细节缺失较多，而受光面则有轻微过曝倾向，导致画面宽容度表现一般。内容层面，这属于典型的城市街景随手记录，虽然主体（共享单车）明确，相关品牌标识清晰可辨，但缺乏独特的视角或情感切入点，主题表达较为平淡，未能从日常场景中提炼出更强的视觉或叙事张力。整体而言，这是一张有一定形式感、氛围尚可的夜景记录，但在技术控制与内容深度上均有提升空间。\",\"improvement_tips\":[\"尝试寻找更简洁的背景或调整角度，减少画面边缘的杂乱元素，让主体更突出。\",\"使用HDR或后期提亮暗部并压暗高光，恢复阴影细节，提升整体宽容度与通透感。\",\"若手机拍摄可开启夜景模式，若相机拍摄可降低ISO并借助三脚架稳定设备，减少暗部噪点。\",\"等待或制造一个视觉焦点（如骑行者、独特光影、地面反射），增强内容的叙事性与趣味性。\",\"尝试竖构图或更极致的局部特写，强化线条节奏感，避免同类元素过度拥挤。\"],\"focused_dimension\":\"\",\"focused_deep_analysis\":\"\"}",
        "created_at": "2026-06-23T04:16:35.351Z"
      },
      {
        "id": 1,
        "user_id": 1,
        "analysis_type": "score",
        "input_prompt": "",
        "focus_dimension": "",
        "result_json": "{\"dimension_scores\":{\"composition\":65,\"color\":70,\"exposure\":62,\"content\":70},\"dimension_notes\":{\"composition\":\"对角线塔吊带来延伸感，但右侧黑边与前景遮挡削弱了画面完整性。\",\"color\":\"冷暖黄蓝对比突出主体，整体色调符合夜施工氛围但层次略显单一。\",\"exposure\":\"夜景基调沉稳，但塔吊顶端光源过曝且暗部细节仍有提升空间。\",\"content\":\"夜建主题明确、氛围感强，前景玻璃痕迹虽添情绪却干扰主体清晰度。\"},\"overall_analysis\":\"这张照片捕捉了雨夜（或雾夜）中城市建筑工地的一隅，塔吊在冷色调的环境中发出暖黄的光，形成鲜明的冷暖对比，营造出孤独而静谧的工业氛围。构图上，塔吊臂呈对角线延伸，为画面带来一定的动感与指向性，但右侧及底部的深色边缘略显突兀，干扰了画面的完整性。色彩方面，蓝黄对比有效地突出了主体，然而整体饱和度偏低，夜空层次稍显单调。曝光控制基本合理，保留了夜晚应有的深沉感，但塔吊顶部的强光源出现了轻微过曝与眩光，暗部细节也有待提升。拍摄内容具有明确的城市建设意象，透过带有水珠或污渍的玻璃拍摄，虽增添了现场氛围和层次，但也牺牲了一部分主体的清晰度，削弱了视觉冲击力。\",\"improvement_tips\":[\"尝试调整拍摄角度，减少右侧深色边框的干扰，或干脆利用窗框形成更完整的框式构图。\",\"使用点测光或降低曝光补偿，适当压暗塔吊顶部强光源，避免眩光并保留灯光细节。\",\"若条件允许，清洁玻璃或贴近玻璃拍摄以减少前景污迹对主体的遮挡，增强画面通透感。\",\"后期可适当提亮阴影并增加局部对比度，让脚手架纹理与夜空云层层次更加丰富。\"],\"focused_dimension\":\"\",\"focused_deep_analysis\":\"\"}",
        "created_at": "2026-06-23T03:47:58.56Z"
      }
    ],
    "page": 1,
    "total": 2
  },
  "message": "ok"
}
```

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
