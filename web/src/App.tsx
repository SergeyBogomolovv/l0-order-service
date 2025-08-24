import { useState } from 'react'
import type { Order } from './types'
import './App.css'

function App() {
  const [orderId, setOrderId] = useState('')
  const [order, setOrder] = useState<Order | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchOrder = async () => {
    if (!orderId) return
    setLoading(true)
    setError(null)
    try {
      const res = await fetch(`http://localhost:9000/order/${orderId}`)
      if (res.status === 404) {
        throw new Error('Заказ не найден')
      }
      if (!res.ok) {
        throw new Error(`Ошибка: ${res.status}`)
      }
      const data: Order = await res.json()
      setOrder(data)
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message)
      }
      setOrder(null)
    } finally {
      setLoading(false)
    }
  }

  return (
    <main>
      <h1>Поиск заказа</h1>

      <div className='search-box'>
        <input
          className='search-input'
          type='text'
          value={orderId}
          onChange={(e) => setOrderId(e.target.value)}
          placeholder='Введите order_id'
        />
        <button onClick={fetchOrder}>Найти</button>
      </div>

      {loading && <p>Загрузка...</p>}
      {error && <p className='error'>{error}</p>}

      {order && (
        <div className='order-card'>
          <h2>Информация о заказе</h2>
          <p>
            <b>UID:</b> {order.order_uid}
          </p>
          <p>
            <b>Покупатель:</b> {order.delivery.name} ({order.delivery.phone})
          </p>
          <p>
            <b>Адрес:</b> {order.delivery.city}, {order.delivery.address}
          </p>
          <p>
            <b>Сумма:</b> {order.payment.amount} {order.payment.currency}
          </p>
          <h3>Товары:</h3>
          <ul className='items-list'>
            {order.items.map((item) => (
              <li key={item.chrt_id}>
                {item.name} — {item.price} ({item.brand})
              </li>
            ))}
          </ul>
        </div>
      )}
    </main>
  )
}

export default App
