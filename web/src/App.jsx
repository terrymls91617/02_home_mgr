import { useState, useEffect } from 'react'
import './App.css'

function App() {
  const [accounts, setAccounts] = useState([])
  const [form, setForm] = useState({
    name: '',
    type: '其他',
    username: '',
    password: '',
    note: ''
  })
  const [showPassword, setShowPassword] = useState(false)
  const [editingId, setEditingId] = useState(null)

  const API_BASE = import.meta.env.VITE_API_BASE || '/api'

  // 从后端加载数据
  useEffect(() => {
    fetchAccounts()
  }, [])

  // 获取账号列表
  const fetchAccounts = async () => {
    try {
      const res = await fetch(`${API_BASE}/accounts`)
      const data = await res.json()
      if (data.code === 200) {
        setAccounts(data.data || [])
      }
    } catch (err) {
      console.error('获取账号列表失败:', err)
      // 降级使用localStorage
      const saved = localStorage.getItem('homeAccounts')
      if (saved) {
        setAccounts(JSON.parse(saved))
      }
    }
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    try {
      if (editingId) {
        // 编辑
        const res = await fetch(`${API_BASE}/accounts/${editingId}`, {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(form)
        })
        const data = await res.json()
        if (data.code === 200) {
          setAccounts(accounts.map(acc => 
            acc.id === editingId ? data.data : acc
          ))
          setEditingId(null)
        }
      } else {
        // 新增
        const res = await fetch(`${API_BASE}/accounts`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(form)
        })
        const data = await res.json()
        if (data.code === 200) {
          setAccounts([...accounts, data.data])
        }
      }
      setForm({
        name: '',
        type: '其他',
        username: '',
        password: '',
        note: ''
      })
    } catch (err) {
      console.error('保存失败:', err)
      // 降级使用localStorage
      if (editingId) {
        setAccounts(accounts.map(acc => 
          acc.id === editingId ? { ...form, id: editingId } : acc
        ))
        setEditingId(null)
      } else {
        setAccounts([...accounts, { ...form, id: Date.now() }])
      }
      setForm({
        name: '',
        type: '其他',
        username: '',
        password: '',
        note: ''
      })
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
        const res = await fetch(`${API_BASE}/accounts/${id}`, {
          method: 'DELETE'
        })
        const data = await res.json()
        if (data.code === 200) {
          setAccounts(accounts.filter(acc => acc.id !== id))
        }
      } catch (err) {
        console.error('删除失败:', err)
        // 降级使用localStorage
        setAccounts(accounts.filter(acc => acc.id !== id))
      }
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

  return (
    <div className="app">
      <header>
        <h1>🏠 家庭账号管理</h1>
        <p>安全管理家庭各类账号信息</p>
      </header>

      <div className="container">
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
                      note: ''
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
                <div key={account.id} className="account-card">
                  <div className="card-header">
                    <h3>{account.name}</h3>
                    <span className="tag">{account.type}</span>
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
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>
      </div>
    </div>
  )
}

export default App
