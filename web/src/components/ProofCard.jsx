import React, { useState } from 'react'
import { useProviderDomain } from '../hooks/useProviderDomain'

function ProofCard({ proof }) {
  const { domain, loading } = useProviderDomain(proof.prover)
  const [isHovered, setIsHovered] = useState(false)

  const formatDate = (dateString) => {
    const date = new Date(dateString)
    return date.toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  // Extract domain from URL if it's a full URL
  const getDisplayDomain = (urlOrDomain) => {
    if (!urlOrDomain) return proof.prover
    
    try {
      // If it's a URL, extract the hostname
      if (urlOrDomain.startsWith('http://') || urlOrDomain.startsWith('https://')) {
        const url = new URL(urlOrDomain)
        return url.hostname
      }
      return urlOrDomain
    } catch {
      return urlOrDomain
    }
  }

  const displayDomain = domain ? getDisplayDomain(domain) : (loading ? 'Loading...' : proof.prover)
  const showAddress = domain && domain !== proof.prover && isHovered

  return (
    <div className="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition-shadow duration-200 border border-gray-200">
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
        <div className="flex-1">
          <div className="mb-2">
            <span className="text-sm font-medium text-gray-500">Merkle</span>
            <p className="text-lg font-mono text-gray-900 break-all mt-1">
              {proof.merkle}
            </p>
          </div>
          
          <div className="mb-2">
            <span className="text-sm font-medium text-gray-500">Prover</span>
            <p
              className="text-base font-mono text-gray-700 break-all mt-1 cursor-default transition-colors duration-200"
              title={proof.prover}
              onMouseEnter={() => setIsHovered(true)}
              onMouseLeave={() => setIsHovered(false)}
            >
              {showAddress ? proof.prover : displayDomain}
            </p>
          </div>

          {proof.block && (
            <div className="flex flex-wrap gap-4 mt-3">
              <div>
                <span className="text-sm font-medium text-gray-500">Height</span>
                <p className="text-base text-gray-700 font-semibold">
                  {proof.block.height.toLocaleString()}
                </p>
              </div>
              <div>
                <span className="text-sm font-medium text-gray-500">Time</span>
                <p className="text-base text-gray-700">
                  {formatDate(proof.block.time)}
                </p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default ProofCard

