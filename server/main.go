package main

import (
	"database/sql"
	"encoding/json"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// Account 账号模型
type Account struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Username string `json:"username"`
	Password string `json:"password"`
	Note     string `json:"note"`
}

var db *sql.DB

func main() {
	// 连接数据库
	var err error
	db, err = sql.Open("mysql", "root:home_mgr_2024@tcp(db:3306)/home_mgr?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		g.Log().Fatal(nil, "数据库连接失败:", err)
	}
	defer db.Close()

	// 自动创建表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS account (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL COMMENT '账号名称',
			type VARCHAR(100) NOT NULL COMMENT '账号类型',
			username VARCHAR(255) NOT NULL COMMENT '用户名',
			password VARCHAR(255) NOT NULL COMMENT '密码',
			note TEXT COMMENT '备注',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`)
	if err != nil {
		g.Log().Error(nil, "创建表失败:", err)
	}

	s := g.Server()
	s.SetPort(8000)

	// 跨域配置
	s.Use(func(r *ghttp.Request) {
		r.Response.CORSDefault()
		r.Middleware.Next()
	})

	// 统一处理账号相关请求
	s.BindHandler("/api/accounts", func(r *ghttp.Request) {
		switch r.Method {
		case "GET":
			getAccounts(r)
		case "POST":
			createAccount(r)
		default:
			r.Response.Status = 405
		}
	})

	s.BindHandler("/api/accounts/{id}", func(r *ghttp.Request) {
		id, _ := strconv.Atoi(r.Get("id").String())
		switch r.Method {
		case "PUT":
			updateAccount(r, id)
		case "DELETE":
			deleteAccount(r, id)
		default:
			r.Response.Status = 405
		}
	})

	s.Run()
}

// 获取账号列表
func getAccounts(r *ghttp.Request) {
	rows, err := db.Query("SELECT id, name, type, username, password, note FROM account")
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "获取账号列表失败",
			"err":  err.Error(),
		})
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var acc Account
		err := rows.Scan(&acc.Id, &acc.Name, &acc.Type, &acc.Username, &acc.Password, &acc.Note)
		if err != nil {
			continue
		}
		accounts = append(accounts, acc)
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 200,
		"msg":  "success",
		"data": accounts,
	})
}

// 创建账号
func createAccount(r *ghttp.Request) {
	var account Account
	if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 400,
			"msg":  "参数错误",
			"err":  err.Error(),
		})
	}

	if account.Name == "" || account.Type == "" || account.Username == "" || account.Password == "" {
		r.Response.WriteJsonExit(g.Map{
			"code": 400,
			"msg":  "必填字段不能为空",
		})
	}

	result, err := db.Exec(
		"INSERT INTO account (name, type, username, password, note) VALUES (?, ?, ?, ?, ?)",
		account.Name, account.Type, account.Username, account.Password, account.Note,
	)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "新增账号失败",
			"err":  err.Error(),
		})
	}

	id, _ := result.LastInsertId()
	account.Id = int(id)

	r.Response.WriteJsonExit(g.Map{
		"code": 200,
		"msg":  "success",
		"data": account,
	})
}

// 更新账号
func updateAccount(r *ghttp.Request, id int) {
	var account Account
	if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 400,
			"msg":  "参数错误",
			"err":  err.Error(),
		})
	}

	if account.Name == "" || account.Type == "" || account.Username == "" || account.Password == "" {
		r.Response.WriteJsonExit(g.Map{
			"code": 400,
			"msg":  "必填字段不能为空",
		})
	}

	_, err := db.Exec(
		"UPDATE account SET name=?, type=?, username=?, password=?, note=? WHERE id=?",
		account.Name, account.Type, account.Username, account.Password, account.Note, id,
	)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "更新账号失败",
			"err":  err.Error(),
		})
	}

	account.Id = id
	r.Response.WriteJsonExit(g.Map{
		"code": 200,
		"msg":  "success",
		"data": account,
	})
}

// 删除账号
func deleteAccount(r *ghttp.Request, id int) {
	_, err := db.Exec("DELETE FROM account WHERE id=?", id)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "删除账号失败",
			"err":  err.Error(),
		})
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 200,
		"msg":  "删除成功",
	})
}
