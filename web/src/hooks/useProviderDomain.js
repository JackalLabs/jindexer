import { useState, useEffect, useRef } from 'react'

// In-memory cache for provider domains
const domainCache = new Map()

export function useProviderDomain(address) {
  const [domain, setDomain] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const mountedRef = useRef(true)

  useEffect(() => {
    // Reset mounted flag when effect runs
    mountedRef.current = true
    
    if (!address) {
      setLoading(false)
      return
    }

    // Check cache first
    if (domainCache.has(address)) {
      const cachedDomain = domainCache.get(address)
      setDomain(cachedDomain)
      setLoading(false)
      return
    }

    // Fetch from API
    const fetchDomain = async () => {
      try {
        setLoading(true)
        setError(null)

        const response = await fetch(`/api/provider/${encodeURIComponent(address)}`)
        
        if (!response.ok) {
          if (response.status === 404) {
            // Provider not found, use address as fallback
            if (mountedRef.current) {
              setDomain(address)
              setLoading(false)
            }
            domainCache.set(address, address)
            return
          }
          throw new Error('Failed to fetch provider domain')
        }

        const data = await response.json()
        const providerDomain = data.ip || address
        
        // Update cache
        domainCache.set(address, providerDomain)
        
        if (mountedRef.current) {
          setDomain(providerDomain)
          setLoading(false)
        }
      } catch (err) {
        if (mountedRef.current) {
          setError(err.message)
          // Fallback to address if fetch fails
          setDomain(address)
          setLoading(false)
        }
        domainCache.set(address, address)
      }
    }

    fetchDomain()

    return () => {
      mountedRef.current = false
    }
  }, [address])

  return { domain, loading, error }
}

