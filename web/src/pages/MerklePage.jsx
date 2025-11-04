import React, { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import ProofList from '../components/ProofList'

function MerklePage() {
  const { merkle } = useParams()
  const [proofs, setProofs] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    if (merkle) {
      fetchProofsByMerkle(merkle)
    }
  }, [merkle])

  const fetchProofsByMerkle = async (merkleHash) => {
    try {
      setLoading(true)
      setError(null)
      
      // Calculate date range (last 30 days)
      const endDate = new Date()
      const startDate = new Date()
      startDate.setDate(startDate.getDate() - 30)
      
      const startDateStr = startDate.toISOString()
      const endDateStr = endDate.toISOString()
      
      const response = await fetch(
        `/api/query?merkle=${encodeURIComponent(merkleHash)}&start_date=${startDateStr}&end_date=${endDateStr}`
      )
      
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

  return (
    <div className="container mx-auto px-4 py-8 max-w-6xl">
      <div className="mb-6">
        <Link
          to="/"
          className="text-blue-600 hover:text-blue-800 inline-flex items-center gap-2 mb-4"
        >
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M15 19l-7-7 7-7"
            />
          </svg>
          Back to Home
        </Link>
      </div>

      <div className="mb-8">
        <h1 className="text-4xl font-bold text-gray-900 mb-2">Proofs for Merkle</h1>
        <div className="bg-gray-100 rounded-lg p-4">
          <p className="font-mono text-sm break-all text-gray-800">{merkle}</p>
        </div>
      </div>

      <div>
        <h2 className="text-2xl font-semibold text-gray-900 mb-4">
          Found {proofs.length} proof{proofs.length !== 1 ? 's' : ''}
        </h2>
        <ProofList proofs={proofs} loading={loading} error={error} />
      </div>
    </div>
  )
}

export default MerklePage

