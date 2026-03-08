# 🏠 Home Manager - 家庭账号管理器

一个简单实用的家庭账号密码管理工具，帮助您安全地管理和存储各类账号信息。

## ✨ 功能特性

- 📝 **账号管理** - 添加、编辑、删除账号信息
- 🔐 **分类存储** - 按类型分类管理账号（网站、应用、银行卡等）
- 📋 **备注功能** - 为每个账号添加备注信息
- 🐳 **Docker 部署** - 一键部署，开箱即用

## 🛠️ 技术栈

### 后端
- **Go 1.22+** - 编程语言
- **GoFrame v2** - Web 框架
- **MySQL 8.0** - 数据库

### 前端
- **React 19** - UI 框架
- **Vite 7** - 构建工具
- **Nginx** - 静态资源服务

### 部署
- **Docker Compose** - 容器编排

## 🚀 快速开始

### 前置要求

- Docker & Docker Compose
- Git

### 一键部署

```bash
# 克隆仓库
git clone https://github.com/terrymls91617/02_home_mgr.git
cd 02_home_mgr

# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps
```

### 访问应用

- **前端界面**: http://localhost:8080
- **后端 API**: http://localhost:8000
- **MySQL**: localhost:3306

## 📁 项目结构

```
02_home_mgr/
├── server/                 # 后端服务
│   ├── main.go            # 主程序入口
│   ├── Dockerfile         # 后端 Docker 配置
│   └── go.mod             # Go 模块定义
├── web/                    # 前端服务
│   ├── src/               # 源代码
│   │   ├── App.jsx        # 主组件
│   │   ├── App.css        # 样式文件
│   │   └── main.jsx       # 入口文件
│   ├── Dockerfile         # 前端 Docker 配置
│   ├── nginx.conf         # Nginx 配置
│   └── package.json       # 依赖配置
├── docker-compose.yml      # Docker 编排配置
└── README.md              # 项目文档
```

## 🔌 API 接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/accounts` | 获取账号列表 |
| POST | `/api/accounts` | 创建新账号 |
| PUT | `/api/accounts/:id` | 更新账号信息 |
| DELETE | `/api/accounts/:id` | 删除账号 |

### 数据模型

```json
{
  "id": 1,
  "name": "账号名称",
  "type": "账号类型",
  "username": "用户名",
  "password": "密码",
  "note": "备注信息"
}
```

## ⚙️ 本地开发

### 后端开发

```bash
cd server

# 安装依赖
go mod download

# 运行开发服务器（需要本地 MySQL）
go run main.go
```

### 前端开发

```bash
cd web

# 安装依赖
npm install

# 运行开发服务器
npm run dev

# 构建生产版本
npm run build
```

## 🔧 配置说明

### 环境变量

后端服务支持以下环境变量配置：

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| GF_SERVER_PORT | 8000 | 服务端口 |
| GF_DATABASE_HOST | db | 数据库主机 |
| GF_DATABASE_PORT | 3306 | 数据库端口 |
| GF_DATABASE_NAME | home_mgr | 数据库名称 |
| GF_DATABASE_USER | root | 数据库用户 |
| GF_DATABASE_PASS | home_mgr_2024 | 数据库密码 |

### 数据持久化

MySQL 数据通过 Docker Volume 持久化存储：

```bash
# 查看数据卷
docker volume ls | grep mysql_data

# 备份数据
docker exec <mysql_container> mysqldump -u root -phome_mgr_2024 home_mgr > backup.sql
```

## 🛡️ 安全建议

> ⚠️ **重要提示**: 此项目为个人家庭使用场景设计，密码以明文存储在数据库中。

生产环境建议：
- 修改默认数据库密码
- 启用 HTTPS
- 限制网络访问范围
- 定期备份数据

## 📝 更新日志

### v0.0.1 (2024-03-07)
- 初始版本发布
- 实现账号 CRUD 基础功能
- Docker 容器化部署支持

## 📄 许可证

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！
