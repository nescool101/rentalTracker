import { Routes, Route } from 'react-router-dom';
import MainLayout from './layouts/MainLayout';
import PublicLayout from './layouts/PublicLayout';
import ProtectedRoute from './components/ProtectedRoute';
import { Suspense, lazy } from 'react';
import { LoadingOverlay } from '@mantine/core';
import { AuthProvider } from './contexts/AuthContext';

// Debug component
const DebugAuth = lazy(() => import('./components/DebugAuth'));
const TestModal = lazy(() => import('./pages/TestModal'));

// Lazy load pages for better performance
const Dashboard = lazy(() => import('./pages/Dashboard'));
const Properties = lazy(() => import('./pages/Properties'));
const Persons = lazy(() => import('./pages/Persons'));
const Profile = lazy(() => import('./pages/Profile'));
const MaintenanceRequests = lazy(() => import('./pages/MaintenanceRequests'));
const Payments = lazy(() => import('./pages/Payments'));
const RentalHistory = lazy(() => import('./pages/RentalHistory'));
const Contracts = lazy(() => import('./pages/Contracts'));
const Users = lazy(() => import('./pages/Users'));
const NotFound = lazy(() => import('./pages/NotFound'));
const Landing = lazy(() => import('./pages/Landing'));
const About = lazy(() => import('./pages/About'));
const Contact = lazy(() => import('./pages/Contact'));
const Login = lazy(() => import('./pages/Login'));
const AccessDenied = lazy(() => import('./pages/AccessDenied'));
const RentalPricingPage = lazy(() => import('./pages/Admin/RentalPricingPage'));
const ManualEmailSenderPage = lazy(() => import('./pages/Admin/ManualEmailSenderPage'));
const ContractGenerationPage = lazy(() => import('./pages/Admin/ContractGenerationPage'));
const ContractSigningPage = lazy(() => import('./pages/ContractSigningPage'));
const ManagerRegistration = lazy(() => import('./pages/ManagerRegistration'));
const ManagerInvitationPage = lazy(() => import('./pages/Admin/ManagerInvitationPage'));
const BankAccountManagement = lazy(() => import('./pages/Admin/BankAccountManagement'));
const FileUploadManagement = lazy(() => import('./pages/Admin/FileUploadManagement'));
const FileManagement = lazy(() => import('./pages/Admin/FileManagement'));
const FileUploadPage = lazy(() => import('./pages/FileUploadPage'));
const OnboardingStepper = lazy(() => import('./components/OnboardingStepper'));


function App() {
  return (
    <AuthProvider>
      <Suspense fallback={<LoadingOverlay visible overlayProps={{ blur: 2 }} />}>
        <Routes>
          {/* Public Routes */}
          <Route path="/" element={<PublicLayout />}>
            <Route index element={<Landing />} />
            <Route path="about" element={<About />} />
            <Route path="contact" element={<Contact />} />
            <Route path="login" element={<Login />} />
            <Route path="access-denied" element={<AccessDenied />} />
            <Route path="debug" element={<DebugAuth />} />
            <Route path="test-modal" element={<TestModal />} />
            <Route path="sign/:signingId" element={<ContractSigningPage />} />
            <Route path="file-upload" element={<FileUploadPage />} />
          </Route>
          
          {/* Protected Routes (require authentication) */}
          <Route path="/" element={<ProtectedRoute />}>
            {/* Onboarding stepper is protected and only accessible to logged-in users */}
            <Route path="onboarding" element={<OnboardingStepper />} />
            
            {/* Main Dashboard Layout */}
            <Route path="dashboard" element={<MainLayout />}>
              <Route index element={<Dashboard />} />
            </Route>

            {/* Top-level routes for direct access */}
            <Route path="properties" element={<MainLayout />}>
              <Route index element={<Properties />} />
            </Route>
            <Route path="persons" element={<MainLayout />}>
              <Route index element={<Persons />} />
            </Route>
            <Route path="profile" element={<MainLayout />}>
              <Route index element={<Profile />} />
            </Route>
            <Route path="maintenance" element={<MainLayout />}>
              <Route index element={<MaintenanceRequests />} />
            </Route>
            <Route path="payments" element={<MainLayout />}>
              <Route index element={<Payments />} />
            </Route>
            <Route path="rental-history" element={<MainLayout />}>
              <Route index element={<RentalHistory />} />
            </Route>
            <Route path="contracts" element={<MainLayout />}>
              <Route index element={<Contracts />} />
            </Route>
            <Route path="users" element={<MainLayout />}>
              <Route index element={<Users />} />
            </Route>
            <Route path="register/manager" element={<MainLayout />}>
              <Route index element={<ManagerRegistration />} />
            </Route>
          </Route>
          
          {/* Admin Routes (require admin privileges) */}
          <Route path="/admin" element={<ProtectedRoute requireAdmin />}>
            <Route path="" element={<MainLayout />}>
              <Route index element={<Dashboard />} />
            </Route>
            <Route path="properties" element={<MainLayout />}>
              <Route index element={<Properties />} />
            </Route>
            <Route path="persons" element={<MainLayout />}>
              <Route index element={<Persons />} />
            </Route>
            <Route path="maintenance" element={<MainLayout />}>
              <Route index element={<MaintenanceRequests />} />
            </Route>
            <Route path="payments" element={<MainLayout />}>
              <Route index element={<Payments />} />
            </Route>
            <Route path="history" element={<MainLayout />}>
              <Route index element={<RentalHistory />} />
            </Route>
            <Route path="contracts" element={<MainLayout />}>
              <Route index element={<Contracts />} />
            </Route>
            <Route path="users" element={<MainLayout />}>
              <Route index element={<Users />} />
            </Route>
            <Route path="rental-pricing" element={<MainLayout />}>
              <Route index element={<RentalPricingPage />} />
            </Route>
            <Route path="manual-email-sender" element={<MainLayout />}>
              <Route index element={<ManualEmailSenderPage />} />
            </Route>
            <Route path="contract-generation" element={<MainLayout />}>
              <Route index element={<ContractGenerationPage />} />
            </Route>
            <Route path="manager-invitations" element={<MainLayout />}>
              <Route index element={<ManagerInvitationPage />} />
            </Route>
            <Route path="bank-accounts" element={<MainLayout />}>
              <Route index element={<BankAccountManagement />} />
            </Route>
            <Route path="file-upload" element={<MainLayout />}>
              <Route index element={<FileUploadManagement />} />
            </Route>
            <Route path="file-management" element={<MainLayout />}>
              <Route index element={<FileManagement />} />
            </Route>
            
          </Route>
          
          {/* Fallback route */}
          <Route path="*" element={<NotFound />} />
        </Routes>
      </Suspense>
    </AuthProvider>
  );
}

export default App;
