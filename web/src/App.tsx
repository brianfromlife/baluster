import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { AuthProvider } from "@/providers/AuthProvider";
import { OrganizationProvider } from "@/providers/OrganizationProvider";
import { LandingPage } from "@/pages/Landing";
import { LoginPage } from "@/pages/Login";
import { CallbackPage } from "@/pages/Callback";
import { DashboardPage } from "@/pages/Dashboard";
import { OrganizationsPage } from "@/pages/Organizations";
import { ApplicationsPage } from "@/pages/Applications";
import { ApplicationDetailPage } from "@/pages/ApplicationDetail";
import { CreateServiceKeyPage } from "@/pages/CreateServiceKey";
import { CreateApiKeyPage } from "@/pages/CreateApiKey";
import { ProtectedRoute } from "@/components/ProtectedRoute";
import { DemoBanner } from "@/components/DemoBanner";
import { Layout } from "@/components/Layout";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <OrganizationProvider>
          <BrowserRouter>
            <DemoBanner />
            <Routes>
              <Route path="/" element={<LandingPage />} />
              <Route path="/login" element={<LoginPage />} />
              <Route path="/auth/callback" element={<CallbackPage />} />
              <Route
                path="/dashboard"
                element={
                  <ProtectedRoute>
                    <Layout>
                      <DashboardPage />
                    </Layout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/organizations"
                element={
                  <ProtectedRoute>
                    <Layout>
                      <OrganizationsPage />
                    </Layout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/applications"
                element={
                  <ProtectedRoute>
                    <Layout>
                      <ApplicationsPage />
                    </Layout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/applications/:id"
                element={
                  <ProtectedRoute>
                    <Layout>
                      <ApplicationDetailPage />
                    </Layout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/service-keys/create"
                element={
                  <ProtectedRoute>
                    <Layout>
                      <CreateServiceKeyPage />
                    </Layout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/api-keys/create"
                element={
                  <ProtectedRoute>
                    <Layout>
                      <CreateApiKeyPage />
                    </Layout>
                  </ProtectedRoute>
                }
              />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </BrowserRouter>
        </OrganizationProvider>
      </AuthProvider>
    </QueryClientProvider>
  );
}

export default App;
