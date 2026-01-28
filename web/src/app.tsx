import { useEffect, useMemo, useState } from 'react'

type User = {
  id: number
  name: string
  email: string
  avatar_url: string
}

type Channel = {
  id: number
  name: string
  display_name: string
}

type Message = {
  id: number
  body: string
  created_at: string
  user_name: string
  avatar_url: string
}

export default function App() {
  const [user, setUser] = useState<User | null>(null)
  const [channels, setChannels] = useState<Channel[]>([])
  const [messages, setMessages] = useState<Message[]>([])
  const [channelId, setChannelId] = useState<number>(1)
  const [draft, setDraft] = useState('')
  const [error, setError] = useState<string | null>(null)

  const activeChannel = useMemo(
    () => channels.find((channel) => channel.id === channelId),
    [channels, channelId]
  )

  useEffect(() => {
    fetch('/api/me', { credentials: 'include' })
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => setUser(data))
      .catch(() => setUser(null))
  }, [])

  useEffect(() => {
    fetch('/api/channels', { credentials: 'include' })
      .then((res) => res.json())
      .then((data) => setChannels(data))
      .catch(() => setError('Failed to load channels'))
  }, [])

  useEffect(() => {
    if (!user) {
      return
    }
    fetch(`/api/messages?channel_id=${channelId}`, { credentials: 'include' })
      .then((res) => res.json())
      .then((data) => setMessages(data.reverse()))
      .catch(() => setError('Failed to load messages'))
  }, [user, channelId])

  const handleSend = async () => {
    if (!draft.trim()) {
      return
    }

    const response = await fetch('/api/messages', {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ channel_id: channelId, body: draft })
    })

    if (!response.ok) {
      setError('Failed to send message')
      return
    }

    const message = (await response.json()) as Message
    setMessages((prev) => [...prev, message])
    setDraft('')
  }

  return (
    <div className="app">
      <aside className="sidebar">
        <div className="brand">ChatApp</div>
        <div className="channel-list">
          <h3>Channels</h3>
          {channels.map((channel) => (
            <button
              key={channel.id}
              type="button"
              className={channel.id === channelId ? 'active' : ''}
              onClick={() => setChannelId(channel.id)}
            >
              {channel.display_name}
            </button>
          ))}
        </div>
        <div className="auth">
          {user ? (
            <div className="user">
              <img src={user.avatar_url} alt={user.name} />
              <div>
                <strong>{user.name}</strong>
                <span>{user.email}</span>
              </div>
              <button type="button" onClick={() => fetch('/auth/logout', { method: 'POST' })}>
                Logout
              </button>
            </div>
          ) : (
            <a className="login" href="/auth/google/login">
              Sign in with Google
            </a>
          )}
        </div>
      </aside>
      <main className="chat">
        <header>
          <h2>{activeChannel?.display_name ?? 'Channel'}</h2>
          {error && <span className="error">{error}</span>}
        </header>
        <section className="messages">
          {user ? (
            messages.map((message) => (
              <article key={message.id}>
                <img src={message.avatar_url} alt={message.user_name} />
                <div>
                  <div className="meta">
                    <strong>{message.user_name}</strong>
                    <span>{new Date(message.created_at).toLocaleString()}</span>
                  </div>
                  <p>{message.body}</p>
                </div>
              </article>
            ))
          ) : (
            <div className="empty">Sign in to start chatting.</div>
          )}
        </section>
        {user && (
          <footer>
            <input
              value={draft}
              placeholder="Send a message..."
              onChange={(event) => setDraft(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === 'Enter') {
                  handleSend()
                }
              }}
            />
            <button type="button" onClick={handleSend}>
              Send
            </button>
          </footer>
        )}
      </main>
    </div>
  )
}
