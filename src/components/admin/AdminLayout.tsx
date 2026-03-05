import { useState } from 'react';
import { Outlet, Navigate } from 'react-router-dom';
import { AdminSidebar } from './AdminSidebar';
import { useAuth } from '../../context/AuthContext';
import { Loader2, Menu } from 'lucide-react';

export function AdminLayout() {
  const { user, isAuthenticated, loading } = useAuth();
  const [sidebarHidden, setSidebarHidden] = useState(() => {
    // Load from localStorage or default to hidden
    const saved = localStorage.getItem('admin_sidebar_hidden');
    if (saved !== null) return saved === 'true';
    return true; // Hidden by default
  });

  // Toggle sidebar visibility
  const toggleSidebar = () => {
    const newState = !sidebarHidden;
    setSidebarHidden(newState);
    localStorage.setItem('admin_sidebar_hidden', String(newState));
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <Loader2 className="w-10 h-10 text-[var(--color-primary)] animate-spin mx-auto mb-4" />
          <p className="text-gray-500">Loading...</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (user?.role !== 'admin') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center max-w-md p-8">
          <div className="w-20 h-20 mx-auto mb-6 bg-red-100 rounded-full flex items-center justify-center">
            <span className="text-4xl">🚫</span>
          </div>
          <h1 className="text-2xl font-bold text-gray-800 mb-2">Access Denied</h1>
          <p className="text-gray-600 mb-6">
            You don't have permission to access the admin panel. Please contact an administrator.
          </p>
          <a
            href="/"
            className="inline-block px-6 py-3 bg-[var(--color-primary)] text-white rounded-lg hover:bg-[var(--color-primary-hover)] transition-colors"
          >
            Back to Home
          </a>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 flex">
      {/* Main Content */}
      <div className="flex-1 flex flex-col min-h-screen">
        {/* Page Content */}
        <main className="flex-1 p-4 lg:p-6 overflow-y-auto">
          <Outlet />
        </main>
      </div>

      {/* Sidebar - positioned on right */}
      {!sidebarHidden && (
        <>
          {/* Overlay to close sidebar */}
          <div
            className="fixed inset-0 bg-black/30 z-40"
            onClick={toggleSidebar}
          />
          {/* Sidebar */}
          <div className="fixed inset-y-0 right-0 z-50">
            <AdminSidebar onToggle={toggleSidebar} />
          </div>
        </>
      )}

      {/* Floating button to show sidebar when hidden */}
      {sidebarHidden && (
        <button
          onClick={toggleSidebar}
          className="fixed top-4 right-4 z-50 p-3 bg-[var(--color-primary)] text-white rounded-lg shadow-lg hover:bg-[var(--color-primary-hover)] transition-all"
          title="Open menu"
        >
          <Menu className="w-5 h-5" />
        </button>
      )}
    </div>
  );
}
