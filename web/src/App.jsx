import { useState, useEffect } from 'react'
import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
  useNavigate
} from 'react-router-dom'
import './App.css'

// 请求封装
const API_BASE = import.meta.env.VITE_API_BASE || '/api'

const request = async (url, options = {}) => {
  const token = localStorage.getItem('token')
  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
  }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const res = await fetch(`${API_BASE}${url}`, {
    ...options,
    headers,
  })

  const data = await res.json()
  if (data.code === 401) {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    window.location.href = '/login'
  }
  return data
}

// 登录页面
function Login() {
  const [form, setForm] = useState({ username: '', password: '' })
  const [isRegister, setIsRegister] = useState(false)
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (token) {
      navigate('/')
    }
  }, [navigate])

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    try {
      const url = isRegister ? '/auth/register' : '/auth/login'
      const data = await request(url, {
        method: 'POST',
        body: JSON.stringify(form),
      })
      if (data.code === 200) {
        if (!isRegister) {
          localStorage.setItem('token', data.data.token)
          localStorage.setItem('user', JSON.stringify(data.data))
          navigate('/')
        } else {
          alert('注册成功，请登录')
          setIsRegister(false)
        }
      } else {
        alert(data.msg)
      }
    } catch (err) {
      alert('网络错误，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="auth-page">
      <div className="auth-card">
        <h1>🏠 家庭账号管理</h1>
        <h2>{isRegister ? '注册账号' : '用户登录'}</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>用户名 *</label>
            <input
              type="text"
              required
              value={form.username}
              onChange={(e) => setForm({ ...form, username: e.target.value })}
              placeholder="请输入用户名"
            />
          </div>
          <div className="form-group">
            <label>密码 *</label>
            <input
              type="password"
              required
              value={form.password}
              onChange={(e) => setForm({ ...form, password: e.target.value })}
              placeholder="请输入密码"
            />
          </div>
          {isRegister && (
            <div className="form-group">
              <label>邀请码 *</label>
              <input
                type="text"
                required
                value={form.invite_code || ''}
                onChange={(e) => setForm({ ...form, invite_code: e.target.value })}
                placeholder="请输入邀请码"
              />
            </div>
          )}
          <div className="form-actions">
            <button type="submit" className="btn-primary" disabled={loading}>
              {loading ? '处理中...' : (isRegister ? '注册' : '登录')}
            </button>
          </div>
        </form>
        <div className="auth-switch">
          <p>
            {isRegister ? '已有账号？' : '还没有账号？'}
            <button type="button" onClick={() => setIsRegister(!isRegister)}>
              {isRegister ? '去登录' : '去注册'}
            </button>
          </p>
        </div>
      </div>
    </div>
  )
}

// 主应用页面
function MainApp() {
  const [accounts, setAccounts] = useState([])
  const [form, setForm] = useState({
    name: '',
    type: '其他',
    username: '',
    password: '',
    note: '',
    is_shared: false
  })
  const [showPassword, setShowPassword] = useState(false)
  const [editingId, setEditingId] = useState(null)
  const [user, setUser] = useState(null)
  const [showAdmin, setShowAdmin] = useState(false)
  const [inviteCodes, setInviteCodes] = useState([])
  const navigate = useNavigate()

  // 检查登录状态
  useEffect(() => {
    const userStr = localStorage.getItem('user')
    if (!userStr) {
      navigate('/login')
      return
    }
    setUser(JSON.parse(userStr))
    fetchAccounts()
  }, [navigate])

  // 获取账号列表
  const fetchAccounts = async () => {
    try {
      const data = await request('/accounts')
      if (data.code === 200) {
        setAccounts(data.data || [])
      }
    } catch (err) {
      console.error('获取账号列表失败:', err)
    }
  }

  // 获取邀请码列表
  const fetchInviteCodes = async () => {
    try {
      const data = await request('/admin/invite-codes')
      if (data.code === 200) {
        setInviteCodes(data.data || [])
      }
    } catch (err) {
      console.error('获取邀请码列表失败:', err)
    }
  }

  // 生成邀请码
  const generateInviteCode = async () => {
    try {
      const data = await request('/admin/invite-codes', { method: 'POST' })
      if (data.code === 200) {
        alert(`邀请码生成成功：${data.data.code}`)
        fetchInviteCodes()
      } else {
        alert(data.msg)
      }
    } catch (err) {
      alert('生成邀请码失败')
    }
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    try {
      if (editingId) {
        // 编辑
        const data = await request(`/accounts/${editingId}`, {
          method: 'PUT',
          body: JSON.stringify(form)
        })
        if (data.code === 200) {
          setAccounts(accounts.map(acc => 
            acc.id === editingId ? data.data : acc
          ))
          setEditingId(null)
        }
      } else {
        // 新增
        const data = await request('/accounts', {
          method: 'POST',
          body: JSON.stringify(form)
        })
        if (data.code === 200) {
          setAccounts([...accounts, data.data])
        }
      }
      setForm({
        name: '',
        type: '其他',
        username: '',
        password: '',
        note: '',
        is_shared: false
      })
    } catch (err) {
      console.error('保存失败:', err)
      alert('保存失败，请稍后重试')
    }
  }

  const handleEdit = (account) => {
    setForm(account)
    setEditingId(account.id)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  const handleDelete = async (id) => {
    if (confirm('确定要删除这个账号吗？')) {
      try {
        const data = await request(`/accounts/${id}`, { method: 'DELETE' })
        if (data.code === 200) {
          setAccounts(accounts.filter(acc => acc.id !== id))
        } else {
          alert(data.msg)
        }
      } catch (err) {
        console.error('删除失败:', err)
        alert('删除失败，请稍后重试')
      }
    }
  }

  const handleLogout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    navigate('/login')
  }

  const toggleAdmin = () => {
    setShowAdmin(!showAdmin)
    if (!showAdmin) {
      fetchInviteCodes()
    }
  }

  const accountTypes = [
    '宽带网络',
    '有线电视',
    '水费',
    '电费',
    '燃气费',
    '物业费',
    '停车费',
    '银行卡',
    '信用卡',
    '手机卡',
    '支付宝',
    '微信支付',
    '公积金',
    '社保',
    '医保',
    '社交账号',
    '邮箱账号',
    '视频会员',
    '音乐会员',
    '云盘服务',
    '购物平台',
    '外卖平台',
    '出行服务',
    '智能家居',
    '其他'
  ]

  if (!user) return null

  return (
    <div className="app">
      <header>
        <div className="header-content">
          <div>
            <h1>🏠 家庭账号管理</h1>
            <p>安全管理家庭各类账号信息</p>
          </div>
          <div className="header-actions">
            <span>欢迎，{user.username}</span>
            {user.is_admin && (
              <button 
                className={`btn-secondary ${showAdmin ? 'active' : ''}`}
                onClick={toggleAdmin}
              >
                {showAdmin ? '返回账号管理' : '管理后台'}
              </button>
            )}
            <button className="btn-secondary" onClick={handleLogout}>
              退出登录
            </button>
          </div>
        </div>
      </header>

      <div className="container">
        {showAdmin ? (
          // 管理员后台
          <section className="admin-section">
            <div className="admin-header">
              <h2>邀请码管理</h2>
              <button className="btn-primary" onClick={generateInviteCode}>
                生成邀请码
              </button>
            </div>
            <div className="invite-code-list">
              {inviteCodes.length === 0 ? (
                <div className="empty-state">
                  <p>还没有生成任何邀请码</p>
                </div>
              ) : (
                <table className="invite-code-table">
                  <thead>
                    <tr>
                      <th>邀请码</th>
                      <th>生成人</th>
                      <th>状态</th>
                      <th>使用人</th>
                      <th>生成时间</th>
                      <th>使用时间</th>
                    </tr>
                  </thead>
                  <tbody>
                    {inviteCodes.map(ic => (
                      <tr key={ic.id}>
                        <td><code>{ic.code}</code></td>
                        <td>{ic.created_by_name}</td>
                        <td>
                          <span className={`tag ${ic.is_used ? 'tag-used' : 'tag-unused'}`}>
                            {ic.is_used ? '已使用' : '未使用'}
                          </span>
                        </td>
                        <td>{ic.used_by_name || '-'}</td>
                        <td>{new Date(ic.created_at).toLocaleString()}</td>
                        <td>{ic.used_at ? new Date(ic.used_at).toLocaleString() : '-'}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </section>
        ) : (
          // 账号管理
          <>
            {/* 表单区域 */}
            <section className="form-section">
              <h2>{editingId ? '编辑账号' : '新增账号'}</h2>
              <form onSubmit={handleSubmit}>
                <div className="form-group">
                  <label>账号名称 *</label>
                  <input
                    type="text"
                    required
                    value={form.name}
                    onChange={(e) => setForm({ ...form, name: e.target.value })}
                    placeholder="例如：中国电信宽带"
                  />
                </div>

                <div className="form-group">
                  <label>账号类型</label>
                  <select
                    value={form.type}
                    onChange={(e) => setForm({ ...form, type: e.target.value })}
                  >
                    {accountTypes.map(type => (
                      <option key={type} value={type}>{type}</option>
                    ))}
                  </select>
                </div>

                <div className="form-group">
                  <label>用户名/账号 *</label>
                  <input
                    type="text"
                    required
                    value={form.username}
                    onChange={(e) => setForm({ ...form, username: e.target.value })}
                    placeholder="账号/手机号/邮箱"
                  />
                </div>

                <div className="form-group">
                  <label>密码 *</label>
                  <div className="password-input-wrapper">
                    <input
                      type={showPassword ? "text" : "password"}
                      required
                      value={form.password}
                      onChange={(e) => setForm({ ...form, password: e.target.value })}
                      placeholder="密码"
                    />
                    <button
                      type="button"
                      className="toggle-password"
                      onClick={() => setShowPassword(!showPassword)}
                    >
                      {showPassword ? "🙈 隐藏" : "👁️ 显示"}
                    </button>
                  </div>
                </div>

                <div className="form-group">
                  <label>备注</label>
                  <textarea
                    value={form.note}
                    onChange={(e) => setForm({ ...form, note: e.target.value })}
                    placeholder="备注信息，比如到期时间、绑定手机号等"
                    rows="3"
                  />
                </div>

                <div className="form-group checkbox-group">
                  <label>
                    <input
                      type="checkbox"
                      checked={form.is_shared}
                      onChange={(e) => setForm({ ...form, is_shared: e.target.checked })}
                    />
                    设为共享账号（所有用户可见）
                  </label>
                </div>

                <div className="form-actions">
                  {editingId && (
                    <button
                      type="button"
                      className="btn-secondary"
                      onClick={() => {
                        setEditingId(null)
                        setForm({
                          name: '',
                          type: '其他',
                          username: '',
                          password: '',
                          note: '',
                          is_shared: false
                        })
                      }}
                    >
                      取消
                    </button>
                  )}
                  <button type="submit" className="btn-primary">
                    {editingId ? '保存修改' : '添加账号'}
                  </button>
                </div>
              </form>
            </section>

            {/* 账号列表 */}
            <section className="list-section">
              <h2>账号列表 ({(accounts || []).length})</h2>
              {(accounts || []).length === 0 ? (
                <div className="empty-state">
                  <p>还没有添加任何账号，点击左上角新增第一个吧</p>
                </div>
              ) : (
                <div className="account-grid">
                  {(accounts || []).map(account => (
                    <div key={account.id} className={`account-card ${account.is_shared ? 'shared' : ''}`}>
                      <div className="card-header">
                        <h3>{account.name}</h3>
                        <div>
                          {account.is_shared && <span className="tag tag-shared">共享</span>}
                          <span className="tag">{account.type}</span>
                        </div>
                      </div>
                      <div className="card-content">
                        <div className="field">
                          <label>账号：</label>
                          <span>{account.username}</span>
                        </div>
                        <div className="field">
                          <label>密码：</label>
                          <span>{account.password}</span>
                        </div>
                        {account.note && (
                          <div className="field note">
                            <label>备注：</label>
                            <span>{account.note}</span>
                          </div>
                        )}
                      </div>
                      <div className="card-actions">
                        {account.user_id === user.id || user.is_admin ? (
                          <>
                            <button
                              className="btn-edit"
                              onClick={() => handleEdit(account)}
                            >
                              编辑
                            </button>
                            <button
                              className="btn-delete"
                              onClick={() => handleDelete(account.id)}
                            >
                              删除
                            </button>
                          </>
                        ) : (
                          <span className="read-only">只读（共享账号）</span>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </section>
          </>
        )}
      </div>
    </div>
  )
}

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/" element={<MainApp />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Router>
  )
}

export default App
