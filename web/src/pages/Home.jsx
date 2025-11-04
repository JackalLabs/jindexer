import React, { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import ProofList from '../components/ProofList'

function Home() {
  const [proofs, setProofs] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [searchTerm, setSearchTerm] = useState('')
  const navigate = useNavigate()

  useEffect(() => {
    fetchRecentProofs()
  }, [])

  const fetchRecentProofs = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await fetch('/api/proofs?limit=100')
      if (!response.ok) {
        throw new Error('Failed to fetch proofs')
      }
      const data = await response.json()
      setProofs(data.proofs || [])
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const handleSearch = (e) => {
    e.preventDefault()
    if (searchTerm.trim()) {
      navigate(`/${encodeURIComponent(searchTerm.trim())}`)
    }
  }

  return (
    <div className="container mx-auto px-4 py-8 max-w-6xl">
      <div className="mb-8">
        <h1 className="text-4xl font-bold text-gray-900 mb-2">Jackal Proof Explorer</h1>
        <p className="text-gray-600">Search and explore proof data</p>
      </div>

      <div className="mb-8">
        <form onSubmit={handleSearch} className="max-w-2xl">
          <div className="flex gap-2">
            <input
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Enter merkle hash to search..."
              className="flex-1 px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-base"
            />
            <button
              type="submit"
              className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors font-medium"
            >
              Search
            </button>
          </div>
        </form>
      </div>

      <div>
        <h2 className="text-2xl font-semibold text-gray-900 mb-4">Recent Proofs</h2>
        <ProofList proofs={proofs} loading={loading} error={error} />
      </div>
    </div>
  )
}

export default Home

