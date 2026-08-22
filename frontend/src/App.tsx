import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { Sidebar } from '@/components/Sidebar'
import { AgentPage } from '@/pages/AgentPage'
import { BeansPage } from '@/pages/BeansPage'
import { CRMPage } from '@/pages/CustomerAndPaymentPage'
import { HomePage } from '@/pages/HomePage'
import { InventoryPage } from '@/pages/InventoryPage'
import { NotificationsPage } from '@/pages/NotificationsPage'
import { ScoutPage } from '@/pages/ScoutPage'

export default function App() {
  return (
    <BrowserRouter>
      <Sidebar>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/agent" element={<AgentPage />} />
          <Route path="/scout" element={<ScoutPage />} />
          <Route path="/inventory" element={<InventoryPage />} />
          <Route path="/beans" element={<BeansPage />} />
          <Route path="/crm" element={<CRMPage />} />
          <Route path="/notifications" element={<NotificationsPage />} />
        </Routes>
      </Sidebar>
    </BrowserRouter>
  )
}
