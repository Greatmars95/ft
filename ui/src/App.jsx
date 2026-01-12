import { useState, useEffect } from 'react'

function App() {
  const [quotes, setQuotes] = useState([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    const fetchQuotes = async () => {
      try {
        setLoading(true)
        const res = await fetch('/quotes')
        const data = await res.json()
        setQuotes(data)
      } catch (error) {
        console.error('Error:', error)
      } finally {
        setLoading(false)
      }
    }

    fetchQuotes()
    const interval = setInterval(fetchQuotes, 2000)
    return () => clearInterval(interval)
  }, [])

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-black text-white p-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-5xl font-black mb-12 text-center bg-gradient-to-r from-blue-400 to-purple-500 bg-clip-text text-transparent">
          Quotopia
        </h1>

        {loading && (
          <div className="text-center mb-8 animate-pulse">
            <div className="inline-block w-8 h-8 border-4 border-blue-500 border-t-transparent rounded-full"></div>
            <p className="mt-2 text-lg">Загружаю котировки...</p>
          </div>
        )}

        <div className="bg-gray-900/50 backdrop-blur-xl rounded-2xl p-8 shadow-2xl border border-gray-700">
          <table className="w-full">
            <thead>
              <tr className="border-b-2 border-gray-700">
                <th className="p-6 text-left font-bold text-xl text-gray-200">Symbol</th>
                <th className="p-6 text-left font-bold text-xl text-gray-200">Price</th>
                <th className="p-6 text-left font-bold text-xl text-gray-200">Time</th>
              </tr>
            </thead>
            <tbody>
              {quotes.map((quote) => (
                <tr key={quote.symbol} className="hover:bg-gray-800/50 transition-all duration-200 border-b border-gray-800">
                  <td className="p-6 font-mono text-2xl font-semibold">
                    {quote.symbol}
                  </td>
                  <td className="p-6">
                    <span className="text-3xl font-black text-green-400">
                      ${quote.price.toLocaleString()}
                    </span>
                  </td>
                  <td className="p-6 text-lg text-gray-400 font-mono">
                    {new Date(quote.timestamp).toLocaleTimeString()}
                  </td>
                </tr>
              ))}
              {quotes.length === 0 && !loading && (
                <tr>
                  <td colSpan="3" className="p-12 text-center text-gray-500">
                    🚀 Запусти Core сервис для котировок
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        <div className="mt-12 text-center text-sm text-gray-500 space-y-2">
          <p>DevOps Sandbox • gRPC + REST + React + Docker</p>
          <p>Обновление каждые 2 сек • TailwindCSS • Vite</p>
        </div>
      </div>
    </div>
  )
}

export default App
