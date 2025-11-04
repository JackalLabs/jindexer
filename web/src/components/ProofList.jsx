import React from 'react'
import ProofCard from './ProofCard'

function ProofList({ proofs, loading, error }) {
  if (loading) {
    return (
      <div className="flex justify-center items-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4">
        <p className="text-red-800">{error}</p>
      </div>
    )
  }

  if (!proofs || proofs.length === 0) {
    return (
      <div className="bg-gray-50 border border-gray-200 rounded-lg p-8 text-center">
        <p className="text-gray-600 text-lg">No proofs found</p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {proofs.map((proof) => (
        <ProofCard key={proof.id} proof={proof} />
      ))}
    </div>
  )
}

export default ProofList

