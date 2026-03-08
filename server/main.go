package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/golang-jwt/jwt/v5"
)

// 全局配置
var (
	db            *sql.DB
	jwtSecret     = []byte(os.Getenv("JWT_SECRET"))
	adminPassword = os.Getenv("ADMIN_PASSWORD")
)

// User 用户模型
type User struct {
	Id        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
}

// InviteCode 邀请码模型
type InviteCode struct {
	Id        int       `json:"id"`
	Code      string    `json:"code"`
	CreatedBy int       `json:"created_by"`
	UsedBy    int       `json:"used_by,omitempty"`
	UsedAt    time.Time `json:"used_at,omitempty"`
	IsUsed    bool      `json:"is_used"`
	CreatedAt time.Time `json:"created_at"`
}

// Account 账号模型
type Account struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Note      string    `json:"note"`
	UserId    int       `json:"user_id"`
	IsShared  bool      `json:"is_shared"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Claims JWT 载荷
type Claims struct {
	UserId   int  `json:"user_id"`
	Username string `json:"username"`
	IsAdmin  bool `json:"is_admin"`
	jwt.RegisteredClaims
}

func main() {
	// 初始化配置
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("home_mgr_default_secret_2024")
	}
	if adminPassword == "" {
		adminPassword = "admin123456"
		g.Log().Warning(nil, "未设置 ADMIN_PASSWORD 环境变量，使用默认密码：admin123456")
	}

	// 连接数据库
	var err error
	db, err = sql.Open("mysql", "root:home_mgr_2024@tcp(db:3306)/home_mgr?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		g.Log().Fatal(nil, "数据库连接失败:", err)
	}
	defer db.Close()

	// 自动创建表
	err = initTables()
	if err != nil {
		g.Log().Error(nil, "初始化表失败:", err)
	}

	// 初始化管理员账号
	err = initAdminUser()
	if err != nil {
		g.Log().Error(nil, "初始化管理员账号失败:", err)
	}

	s := g.Server()
	s.SetPort(8000)

	// 跨域配置
	s.Use(func(r *ghttp.Request) {
		r.Response.CORSDefault()
		r.Middleware.Next()
	})

	// 公开接口（不需要登录）
	s.BindHandler("POST:/api/auth/register", register)
	s.BindHandler("POST:/api/auth/login", login)

	// 需要登录的接口组
	authGroup := s.Group("/api")
	authGroup.Middleware(authMiddleware)
	{
		// 账号相关接口
		authGroup.BindHandler("GET:/accounts", getAccounts)
		authGroup.BindHandler("POST:/accounts", createAccount)
		authGroup.BindHandler("PUT:/accounts/{id}", updateAccount)
		authGroup.BindHandler("DELETE:/accounts/{id}", deleteAccount)

		// 管理员接口
		adminGroup := authGroup.Group("/admin")
		adminGroup.Middleware(adminMiddleware)
		{
			adminGroup.BindHandler("GET:/invite-codes", getInviteCodes)
			adminGroup.BindHandler("POST:/invite-codes", generateInviteCode)
			adminGroup.BindHandler("DELETE:/invite-codes/{id}", deleteInviteCode)
			adminGroup.BindHandler("GET:/users", getUsers)
		}
	}

	s.Run()
}

// 初始化数据库表
func initTables() error {
	// 用户表
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS user (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',
			password VARCHAR(255) NOT NULL COMMENT '密码',
			is_admin TINYINT(1) DEFAULT 0 COMMENT '是否是管理员',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`)
	if err != nil {
		return err
	}

	// 邀请码表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS invite_code (
			id INT AUTO_INCREMENT PRIMARY KEY,
			code VARCHAR(32) NOT NULL UNIQUE COMMENT '邀请码',
			created_by INT NOT NULL COMMENT '创建人ID',
			used_by INT DEFAULT NULL COMMENT '使用人ID',
			used_at DATETIME DEFAULT NULL COMMENT '使用时间',
			is_used TINYINT(1) DEFAULT 0 COMMENT '是否已使用',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (created_by) REFERENCES user(id),
			FOREIGN KEY (used_by) REFERENCES user(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`)
	if err != nil {
		return err
	}

	// 账号表（扩展字段）
	_, err = db.Exec(`
		ALTER TABLE account ADD COLUMN IF NOT EXISTS user_id INT NOT NULL DEFAULT 0 COMMENT '创建用户ID' AFTER note;
	`)
	if err != nil {
		// 如果表不存在，创建新表
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS account (
				id INT AUTO_INCREMENT PRIMARY KEY,
				name VARCHAR(255) NOT NULL COMMENT '账号名称',
				type VARCHAR(100) NOT NULL COMMENT '账号类型',
				username VARCHAR(255) NOT NULL COMMENT '用户名',
				password VARCHAR(255) NOT NULL COMMENT '密码',
				note TEXT COMMENT '备注',
				user_id INT NOT NULL COMMENT '创建用户ID',
				is_shared TINYINT(1) DEFAULT 0 COMMENT '是否共享',
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES user(id)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
		`)
		if err != nil {
			return err
		}
	} else {
		// 表已存在，添加 is_shared 字段
		_, err = db.Exec(`
			ALTER TABLE account ADD COLUMN IF NOT EXISTS is_shared TINYINT(1) DEFAULT 0 COMMENT '是否共享' AFTER user_id;
		`)
		if err != nil {
			return err
		}
	}

	return nil
}

// 初始化管理员账号
func initAdminUser() error {
	// 检查管理员是否存在
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM user WHERE username = 'admin'").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// 加密密码
		hashedPassword := hashPassword(adminPassword)
		_, err := db.Exec(
			"INSERT INTO user (username, password, is_admin) VALUES ('admin', ?, 1)",
			hashedPassword,
		)
		if err != nil {
			return err
		}
		g.Log().Info(nil, "管理员账号初始化成功，用户名：admin")
	} else {
		// 更新管理员密码
		hashedPassword := hashPassword(adminPassword)
		_, err := db.Exec(
			"UPDATE user SET password = ? WHERE username = 'admin'",
			hashedPassword,
		)
		if err != nil {
			return err
		}
		g.Log().Info(nil, "管理员密码已更新")
	}

	return nil
}

// 密码加密
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// 生成 JWT Token
func generateToken(user *User) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour)
	claims := &Claims{
		UserId:   user.Id,
		Username: user.Username,
		IsAdmin:  user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "home_mgr",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// 认证中间件
func authMiddleware(r *ghttp.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		r.Response.WriteJsonExit(g.Map{
			"code": 401,
			"msg":  "未登录",
		})
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		r.Response.WriteJsonExit(g.Map{
			"code": 401,
			"msg":  "无效的认证格式",
		})
	}

	tokenStr := parts[1]
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		r.Response.WriteJsonExit(g.Map{
			"code": 401,
			"msg":  "无效的Token",
		})
	}

	// 将用户信息存入上下文
	r.SetParam("userId", claims.UserId)
	r.SetParam("username", claims.Username)
	r.SetParam("isAdmin", claims.IsAdmin)

	r.Middleware.Next()
}

// 管理员中间件
func adminMiddleware(r *ghttp.Request) {
	isAdmin := r.GetParam("isAdmin").Bool()
	if !isAdmin {
		r.Response.WriteJsonExit(g.Map{
			"code": 403,
			"msg":  "需要管理员权限",
		})
	}
	r.Middleware.Next()
}

// 用户注册
func register(r *ghttp.Request) {
	var req struct {
		Username   string `json:"username"`
		Password   string `json:"password"`
		InviteCode string `json:"invite_code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 400,
			"msg":  "参数错误",
			"err":  err.Error(),
		})
	}

	if req.Username == "" || req.Password == "" || req.InviteCode == "" {
		r.Response.WriteJsonExit(g.Map{
			"code": 400,
			"msg":  "用户名、密码和邀请码不能为空",
		})
	}

	// 检查邀请码是否有效
	var inviteCodeId int
	err := db.QueryRow(`
		SELECT id FROM invite_code 
		WHERE code = ? AND is_used = 0
	`, req.InviteCode).Scan(&inviteCodeId)
	if err != nil {
		if err == sql.ErrNoRows {
			r.Response.WriteJsonExit(g.Map{
				"code": 400,
				"msg":  "无效的邀请码",
			})
		}
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "验证邀请码失败",
			"err":  err.Error(),
		})
	}

	// 检查用户名是否已存在
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM user WHERE username = ?", req.Username).Scan(&count)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "检查用户名失败",
			"err":  err.Error(),
		})
	}
	if count > 0 {
		r.Response.WriteJsonExit(g.Map{
			"code": 400,
			"msg":  "用户名已存在",
		})
	}

	// 加密密码
	hashedPassword := hashPassword(req.Password)

	// 开启事务
	tx, err := db.Begin()
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "系统错误",
			"err":  err.Error(),
		})
	}
	defer tx.Rollback()

	// 创建用户
	result, err := tx.Exec(
		"INSERT INTO user (username, password, is_admin) VALUES (?, ?, 0)",
		req.Username, hashedPassword,
	)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "创建用户失败",
			"err":  err.Error(),
		})
	}

	userId, _ := result.LastInsertId()

	// 标记邀请码为已使用
	_, err = tx.Exec(`
		UPDATE invite_code 
		SET is_used = 1, used_by = ?, used_at = NOW()
		WHERE id = ?
	`, userId, inviteCodeId)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "更新邀请码失败",
			"err":  err.Error(),
		})
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "注册失败",
			"err":  err.Error(),
		})
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 200,
		"msg":  "注册成功",
	})
}

// 获取账号列表
func getAccounts(r *ghttp.Request) {
	userId := r.GetParam("userId").Int()
	
	// 查询自己创建的账号和共享的账号
	rows, err := db.Query(`
		SELECT id, name, type, username, password, note, user_id, is_shared, created_at, updated_at 
		FROM account 
		WHERE user_id = ? OR is_shared = 1
		ORDER BY created_at DESC
	`, userId)
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
		err := rows.Scan(
			&acc.Id, &acc.Name, &acc.Type, &acc.Username, &acc.Password, &acc.Note,
			&acc.UserId, &acc.IsShared, &acc.CreatedAt, &acc.UpdatedAt,
		)
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
	userId := r.GetParam("userId").Int()
	
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
		"INSERT INTO account (name, type, username, password, note, user_id, is_shared) VALUES (?, ?, ?, ?, ?, ?, ?)",
		account.Name, account.Type, account.Username, account.Password, account.Note, userId, account.IsShared,
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
	account.UserId = userId

	r.Response.WriteJsonExit(g.Map{
		"code": 200,
		"msg":  "success",
		"data": account,
	})
}

// 更新账号
func updateAccount(r *ghttp.Request, id int) {
	userId := r.GetParam("userId").Int()
	isAdmin := r.GetParam("isAdmin").Bool()
	
	// 检查账号是否存在且属于当前用户，或者是管理员
	var ownerId int
	err := db.QueryRow("SELECT user_id FROM account WHERE id = ?", id).Scan(&ownerId)
	if err != nil {
		if err == sql.ErrNoRows {
			r.Response.WriteJsonExit(g.Map{
				"code": 404,
				"msg":  "账号不存在",
			})
		}
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "查询账号失败",
			"err":  err.Error(),
		})
	}

	if ownerId != userId && !isAdmin {
		r.Response.WriteJsonExit(g.Map{
			"code": 403,
			"msg":  "无权限修改此账号",
		})
	}

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

	_, err = db.Exec(
		"UPDATE account SET name=?, type=?, username=?, password=?, note=?, is_shared=? WHERE id=?",
		account.Name, account.Type, account.Username, account.Password, account.Note, account.IsShared, id,
	)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "更新账号失败",
			"err":  err.Error(),
		})
	}

	account.Id = id
	account.UserId = ownerId
	r.Response.WriteJsonExit(g.Map{
		"code": 200,
		"msg":  "success",
		"data": account,
	})
}

// 删除账号
func deleteAccount(r *ghttp.Request, id int) {
	userId := r.GetParam("userId").Int()
	isAdmin := r.GetParam("isAdmin").Bool()
	
	// 检查账号是否存在且属于当前用户，或者是管理员
	var ownerId int
	err := db.QueryRow("SELECT user_id FROM account WHERE id = ?", id).Scan(&ownerId)
	if err != nil {
		if err == sql.ErrNoRows {
			r.Response.WriteJsonExit(g.Map{
				"code": 404,
				"msg":  "账号不存在",
			})
		}
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "查询账号失败",
			"err":  err.Error(),
		})
	}

	if ownerId != userId && !isAdmin {
		r.Response.WriteJsonExit(g.Map{
			"code": 403,
			"msg":  "无权限删除此账号",
		})
	}

	_, err = db.Exec("DELETE FROM account WHERE id=?", id)
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

// 管理员：获取邀请码列表
func getInviteCodes(r *ghttp.Request) {
	rows, err := db.Query(`
		SELECT ic.id, ic.code, ic.created_by, ic.used_by, ic.used_at, ic.is_used, ic.created_at, u.username as created_by_name, ub.username as used_by_name
		FROM invite_code ic
		LEFT JOIN user u ON ic.created_by = u.id
		LEFT JOIN user ub ON ic.used_by = ub.id
		ORDER BY ic.created_at DESC
	`)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "获取邀请码列表失败",
			"err":  err.Error(),
		})
	}
	defer rows.Close()

	var inviteCodes []g.Map
	for rows.Next() {
		var (
			id            int
			code          string
			createdBy     int
			usedBy        sql.NullInt64
			usedAt        sql.NullTime
			isUsed        bool
			createdAt     time.Time
			createdByName string
			usedByName    sql.NullString
		)
		err := rows.Scan(&id, &code, &createdBy, &usedBy, &usedAt, &isUsed, &createdAt, &createdByName, &usedByName)
		if err != nil {
			continue
		}

		ic := g.Map{
			"id":               id,
			"code":             code,
			"created_by":       createdBy,
			"created_by_name":  createdByName,
			"is_used":          isUsed,
			"created_at":       createdAt,
		}

		if usedBy.Valid {
			ic["used_by"] = usedBy.Int64
			ic["used_by_name"] = usedByName.String
			ic["used_at"] = usedAt.Time
		}

		inviteCodes = append(inviteCodes, ic)
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 200,
		"msg":  "success",
		"data": inviteCodes,
	})
}

// 管理员：生成邀请码
func generateInviteCode(r *ghttp.Request) {
	userId := r.GetParam("userId").Int()
	
	// 生成随机邀请码
	code := strings.ToUpper(strconv.FormatInt(time.Now().UnixNano(), 36))[:8]
	
	_, err := db.Exec(
		"INSERT INTO invite_code (code, created_by) VALUES (?, ?)",
		code, userId,
	)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "生成邀请码失败",
			"err":  err.Error(),
		})
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 200,
		"msg":  "生成成功",
		"data": g.Map{
			"code": code,
		},
	})
}

// 管理员：删除邀请码
func deleteInviteCode(r *ghttp.Request) {
	id, _ := strconv.Atoi(r.Get("id").String())
	
	// 只能删除未使用的邀请码
	var isUsed bool
	err := db.QueryRow("SELECT is_used FROM invite_code WHERE id = ?", id).Scan(&isUsed)
	if err != nil {
		if err == sql.ErrNoRows {
			r.Response.WriteJsonExit(g.Map{
				"code": 404,
				"msg":  "邀请码不存在",
			})
		}
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "查询邀请码失败",
			"err":  err.Error(),
		})
	}

	if isUsed {
		r.Response.WriteJsonExit(g.Map{
			"code": 400,
			"msg":  "已使用的邀请码不能删除",
		})
	}

	_, err = db.Exec("DELETE FROM invite_code WHERE id = ?", id)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "删除邀请码失败",
			"err":  err.Error(),
		})
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 200,
		"msg":  "删除成功",
	})
}

// 管理员：获取用户列表
func getUsers(r *ghttp.Request) {
	rows, err := db.Query(`
		SELECT id, username, is_admin, created_at FROM user ORDER BY created_at DESC
	`)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{
			"code": 500,
			"msg":  "获取用户列表失败",
			"err":  err.Error(),
		})
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.Id, &user.Username, &user.IsAdmin, &user.CreatedAt)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	r.Response.WriteJsonExit(g.Map{
		"code": 200,
		"msg":  "success",
		"data": users,
	})
}
