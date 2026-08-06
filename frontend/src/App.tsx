import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { Sidebar } from '@/components/Sidebar'
import { CRMPage } from '@/pages/CustomerAndPaymentPage'
import { HomePage } from '@/pages/HomePage'
import { InventoryPage } from '@/pages/InventoryPage'
import { ScoutPage } from '@/pages/ScoutPage'

export default function App() {
  return (
    <BrowserRouter>
      <Sidebar>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/scout" element={<ScoutPage />} />
          <Route path="/inventory" element={<InventoryPage />} />
          <Route path="/crm" element={<CRMPage />} />
        </Routes>
      </Sidebar>
    </BrowserRouter>
  )
}
