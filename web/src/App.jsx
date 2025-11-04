import { Routes, Route } from 'react-router-dom'
import Home from './pages/Home'
import MerklePage from './pages/MerklePage'

function App() {
  return (
    <div className="min-h-screen bg-gray-50">
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/:merkle" element={<MerklePage />} />
      </Routes>
    </div>
  )
}

export default App

