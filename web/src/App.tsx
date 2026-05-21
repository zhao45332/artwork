import { AnimatePresence, motion } from 'framer-motion'
import { FormEvent, useEffect, useMemo, useState } from 'react'

type AuthUser = {
  id: number
  username: string
  nickname: string
  avatar_url?: string
  bio?: string
}

type Artwork = {
  id: number
  user_id: number
  username: string
  user_nickname: string
  user_avatar_url: string
  title: string
  description: string
  author: string
  image_url: string
  cover_image_url: string
  tags: string[]
  view_count: number
  like_count: number
  favorite_count: number
  created_at?: string
}

type Category = {
  id: number
  name: string
  description: string
}

const API_BASE = 'http://127.0.0.1:32800'
const defaultAvatar = 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&w=240&q=80'

const reveal = {
  hidden: { opacity: 0, y: 28 },
  visible: { opacity: 1, y: 0 },
}

async function request<T>(path: string, init?: RequestInit, token?: string): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(init?.headers ?? {}),
    },
  })
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || `Request failed: ${res.status}`)
  }
  return res.json() as Promise<T>
}

function readStorage(key: string) {
  return window.localStorage.getItem(key) ?? ''
}

export default function App() {
  const [mode, setMode] = useState<'discover' | 'studio'>('discover')
  const [token, setToken] = useState('')
  const [refreshToken, setRefreshToken] = useState('')
  const [me, setMe] = useState<AuthUser | null>(null)
  const [artworks, setArtworks] = useState<Artwork[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [selectedArtwork, setSelectedArtwork] = useState<Artwork | null>(null)
  const [keyword, setKeyword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [authForm, setAuthForm] = useState({ username: '', nickname: '', email: '', password: '', account: '' })
  const [publishForm, setPublishForm] = useState({
    title: '',
    description: '',
    author: '',
    imageUrl: '',
    imageUrls: '',
    tags: '',
    categoryId: 0,
    visibility: 1,
  })

  useEffect(() => {
    const savedToken = readStorage('artwork_token')
    const savedRefresh = readStorage('artwork_refresh')
    if (savedToken) setToken(savedToken)
    if (savedRefresh) setRefreshToken(savedRefresh)
  }, [])

  useEffect(() => {
    void Promise.all([loadDiscover(), loadCategories()])
  }, [])

  useEffect(() => {
    if (!token) return
    void loadMe(token)
  }, [token])

  const statsLabel = useMemo(() => {
    if (!me) return 'Guest Mode'
    return `${me.nickname || me.username} / studio online`
  }, [me])

  async function loadDiscover(q = '') {
    setLoading(true)
    setError('')
    try {
      const data = await request<{ artworks: Artwork[] }>('/v1/artworks?sort_by=created_at&sort_order=desc&page=1&page_size=12' + (q ? `&keyword=${encodeURIComponent(q)}` : ''))
      setArtworks(data.artworks ?? [])
      if (data.artworks?.length) {
        setSelectedArtwork(data.artworks[0])
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  async function loadCategories() {
    try {
      const data = await request<{ categories: Category[] }>('/v1/categories')
      setCategories(data.categories ?? [])
    } catch {
      setCategories([])
    }
  }

  async function loadMe(activeToken: string) {
    try {
      const data = await request<{ user: AuthUser }>('/v1/auth/me', undefined, activeToken)
      setMe(data.user)
    } catch {
      setMe(null)
    }
  }

  async function handleRegister(e: FormEvent) {
    e.preventDefault()
    setError('')
    try {
      const data = await request<{ access_token: string; refresh_token: string; user: AuthUser }>('/v1/auth/register', {
        method: 'POST',
        body: JSON.stringify({
          username: authForm.username,
          nickname: authForm.nickname,
          email: authForm.email,
          password: authForm.password,
        }),
      })
      commitAuth(data.access_token, data.refresh_token, data.user)
    } catch (err) {
      setError(err instanceof Error ? err.message : '注册失败')
    }
  }

  async function handleLogin(e: FormEvent) {
    e.preventDefault()
    setError('')
    try {
      const data = await request<{ access_token: string; refresh_token: string; user: AuthUser }>('/v1/auth/login', {
        method: 'POST',
        body: JSON.stringify({ account: authForm.account, password: authForm.password }),
      })
      commitAuth(data.access_token, data.refresh_token, data.user)
    } catch (err) {
      setError(err instanceof Error ? err.message : '登录失败')
    }
  }

  function commitAuth(nextToken: string, nextRefresh: string, user: AuthUser) {
    setToken(nextToken)
    setRefreshToken(nextRefresh)
    setMe(user)
    window.localStorage.setItem('artwork_token', nextToken)
    window.localStorage.setItem('artwork_refresh', nextRefresh)
    setMode('studio')
  }

  async function handlePublish(e: FormEvent) {
    e.preventDefault()
    if (!token) {
      setError('请先登录')
      return
    }
    setError('')
    try {
      await request('/v1/artworks', {
        method: 'POST',
        body: JSON.stringify({
          title: publishForm.title,
          description: publishForm.description,
          author: publishForm.author,
          image_url: publishForm.imageUrl,
          cover_image_url: publishForm.imageUrl,
          image_urls: publishForm.imageUrls.split('\n').map((item) => item.trim()).filter(Boolean),
          tags: publishForm.tags.split(',').map((item) => item.trim()).filter(Boolean),
          category_id: Number(publishForm.categoryId),
          visibility: Number(publishForm.visibility),
          status: 1,
        }),
      }, token)
      setPublishForm({ title: '', description: '', author: '', imageUrl: '', imageUrls: '', tags: '', categoryId: 0, visibility: 1 })
      setMode('discover')
      await loadDiscover()
    } catch (err) {
      setError(err instanceof Error ? err.message : '发布失败')
    }
  }

  return (
    <div className="shell">
      <div className="grid-noise" />
      <motion.header className="topbar" initial="hidden" animate="visible" variants={reveal} transition={{ duration: 0.7 }}>
        <div>
          <p className="eyebrow">ARTWORK COMMUNITY / DIGITAL MINIMALISM</p>
          <h1>画作社区 · 极简动态界面</h1>
        </div>
        <div className="topbar-side">
          <span>{statsLabel}</span>
          <div className="toggle">
            <button className={mode === 'discover' ? 'active' : ''} onClick={() => setMode('discover')}>Discover</button>
            <button className={mode === 'studio' ? 'active' : ''} onClick={() => setMode('studio')}>Studio</button>
          </div>
        </div>
      </motion.header>

      <main className="layout">
        <section className="left-panel">
          <motion.div className="panel intro" initial="hidden" animate="visible" variants={reveal} transition={{ delay: 0.1, duration: 0.7 }}>
            <div className="line" />
            <p className="mono">A living board for artworks, notes, and visual identity.</p>
            <form className="search-row" onSubmit={(e) => { e.preventDefault(); void loadDiscover(keyword) }}>
              <input value={keyword} onChange={(e) => setKeyword(e.target.value)} placeholder="搜索标题 / 作者 / 描述" />
              <button type="submit">Scan</button>
            </form>
          </motion.div>

          <motion.div className="panel feed" initial="hidden" animate="visible" variants={reveal} transition={{ delay: 0.2, duration: 0.7 }}>
            <div className="panel-title">
              <span>DISCOVERY FEED</span>
              <span>{loading ? 'loading…' : `${artworks.length} items`}</span>
            </div>
            <div className="artwork-list">
              {artworks.map((artwork, index) => (
                <motion.button
                  key={artwork.id}
                  className={`artwork-card ${selectedArtwork?.id === artwork.id ? 'selected' : ''}`}
                  onClick={() => setSelectedArtwork(artwork)}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.04 * index, duration: 0.45 }}
                >
                  <img src={artwork.cover_image_url || artwork.image_url || defaultAvatar} alt={artwork.title} />
                  <div>
                    <div className="card-meta">
                      <span>{artwork.user_nickname || artwork.username || artwork.author}</span>
                      <span>{String(artwork.view_count).padStart(3, '0')}</span>
                    </div>
                    <h3>{artwork.title}</h3>
                    <p>{artwork.description || 'No description yet.'}</p>
                  </div>
                </motion.button>
              ))}
            </div>
          </motion.div>
        </section>

        <section className="center-panel">
          <AnimatePresence mode="wait">
            {selectedArtwork ? (
              <motion.article key={selectedArtwork.id} className="panel spotlight" initial={{ opacity: 0, scale: 0.98 }} animate={{ opacity: 1, scale: 1 }} exit={{ opacity: 0, scale: 0.98 }} transition={{ duration: 0.45 }}>
                <img className="hero-image" src={selectedArtwork.cover_image_url || selectedArtwork.image_url || defaultAvatar} alt={selectedArtwork.title} />
                <div className="spotlight-body">
                  <div className="panel-title">
                    <span>FEATURE / #{selectedArtwork.id}</span>
                    <span>{selectedArtwork.user_nickname || selectedArtwork.author}</span>
                  </div>
                  <h2>{selectedArtwork.title}</h2>
                  <p>{selectedArtwork.description}</p>
                  <div className="chip-row">
                    {(selectedArtwork.tags || []).map((tag) => <span key={tag}>{tag}</span>)}
                  </div>
                  <div className="stat-grid">
                    <div><strong>{selectedArtwork.view_count}</strong><span>views</span></div>
                    <div><strong>{selectedArtwork.like_count}</strong><span>likes</span></div>
                    <div><strong>{selectedArtwork.favorite_count}</strong><span>favs</span></div>
                  </div>
                </div>
              </motion.article>
            ) : (
              <motion.article className="panel spotlight empty" initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
                <p>还没有作品数据，启动后端并创建第一条内容就会在这里出现。</p>
              </motion.article>
            )}
          </AnimatePresence>
        </section>

        <section className="right-panel">
          <motion.div className="panel auth-panel" initial="hidden" animate="visible" variants={reveal} transition={{ delay: 0.25, duration: 0.7 }}>
            <div className="panel-title"><span>AUTH NODE</span><span>{me ? 'online' : 'guest'}</span></div>
            {!me ? (
              <>
                <form className="stack-form" onSubmit={handleLogin}>
                  <input placeholder="账号 / 邮箱" value={authForm.account} onChange={(e) => setAuthForm({ ...authForm, account: e.target.value })} />
                  <input type="password" placeholder="密码" value={authForm.password} onChange={(e) => setAuthForm({ ...authForm, password: e.target.value })} />
                  <button type="submit">Login</button>
                </form>
                <div className="divider">OR REGISTER</div>
                <form className="stack-form" onSubmit={handleRegister}>
                  <input placeholder="用户名" value={authForm.username} onChange={(e) => setAuthForm({ ...authForm, username: e.target.value })} />
                  <input placeholder="昵称" value={authForm.nickname} onChange={(e) => setAuthForm({ ...authForm, nickname: e.target.value })} />
                  <input placeholder="邮箱" value={authForm.email} onChange={(e) => setAuthForm({ ...authForm, email: e.target.value })} />
                  <input type="password" placeholder="密码" value={authForm.password} onChange={(e) => setAuthForm({ ...authForm, password: e.target.value })} />
                  <button type="submit">Create Account</button>
                </form>
              </>
            ) : (
              <div className="profile-card">
                <img src={me.avatar_url || defaultAvatar} alt={me.nickname} />
                <div>
                  <strong>{me.nickname || me.username}</strong>
                  <p>@{me.username}</p>
                  <small>{me.bio || 'A new curator enters the grid.'}</small>
                </div>
              </div>
            )}
            {error ? <p className="error">{error}</p> : null}
          </motion.div>

          <motion.div className="panel publish-panel" initial="hidden" animate="visible" variants={reveal} transition={{ delay: 0.3, duration: 0.7 }}>
            <div className="panel-title"><span>STUDIO FORM</span><span>{token ? 'ready' : 'locked'}</span></div>
            <form className="stack-form" onSubmit={handlePublish}>
              <input placeholder="作品标题" value={publishForm.title} onChange={(e) => setPublishForm({ ...publishForm, title: e.target.value })} />
              <input placeholder="署名 / 作者" value={publishForm.author} onChange={(e) => setPublishForm({ ...publishForm, author: e.target.value })} />
              <input placeholder="封面图片 URL" value={publishForm.imageUrl} onChange={(e) => setPublishForm({ ...publishForm, imageUrl: e.target.value })} />
              <textarea placeholder="更多图片 URL，每行一张" value={publishForm.imageUrls} onChange={(e) => setPublishForm({ ...publishForm, imageUrls: e.target.value })} rows={4} />
              <textarea placeholder="一句描述" value={publishForm.description} onChange={(e) => setPublishForm({ ...publishForm, description: e.target.value })} rows={3} />
              <input placeholder="标签，用逗号分隔" value={publishForm.tags} onChange={(e) => setPublishForm({ ...publishForm, tags: e.target.value })} />
              <select value={publishForm.categoryId} onChange={(e) => setPublishForm({ ...publishForm, categoryId: Number(e.target.value) })}>
                <option value={0}>选择分类</option>
                {categories.map((category) => <option key={category.id} value={category.id}>{category.name}</option>)}
              </select>
              <select value={publishForm.visibility} onChange={(e) => setPublishForm({ ...publishForm, visibility: Number(e.target.value) })}>
                <option value={1}>公开</option>
                <option value={2}>仅自己可见</option>
              </select>
              <button type="submit">Publish Artwork</button>
            </form>
          </motion.div>
        </section>
      </main>
    </div>
  )
}
