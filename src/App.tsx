import { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, useNavigate, useParams } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import { CartProvider, useCart } from './context/CartContext';
import { WishlistProvider } from './context/WishlistContext';
import { Header } from './components/Header';
import { MobileNavigation } from './components/MobileNavigation';
import { HomePage } from './pages/HomePage';
import { ProductDetailPage } from './pages/ProductDetailPage';
import { CartPage } from './pages/CartPage';
import { CheckoutPage } from './pages/CheckoutPage';
import { OrderHistoryPage } from './pages/OrderHistoryPage';
import { OrderDetailPage } from './pages/OrderDetailPage';
import { LoginPage } from './pages/LoginPage';
import { RegisterPage } from './pages/RegisterPage';
import { ProfilePage } from './pages/ProfilePage';
import { WishlistPage } from './pages/WishlistPage';
import { CategoriesListPage } from './pages/CategoriesListPage';
import { ProductsPage } from './pages/ProductsPage';
import { PaymentResultPage } from './pages/PaymentResultPage';

// Admin imports
import { AdminLayout } from './components/admin/AdminLayout';
import { DashboardPage } from './pages/admin/DashboardPage';
import { OrdersPage as AdminOrdersPage } from './pages/admin/OrdersPage';
import { UsersPage } from './pages/admin/UsersPage';
import { ProductsPage as AdminProductsPage } from './pages/admin/ProductsPage';
import { CategoriesPage as AdminCategoriesPage } from './pages/admin/CategoriesPage';
import { CouponsPage } from './pages/admin/CouponsPage';

type Page = 'home' | 'product' | 'cart' | 'categories' | 'account';

// Wrapper component for CartPage with proper navigation
function CartPageWrapper() {
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const { cart } = useCart();
  const totalCartItems = cart?.item_count || 0;

  return (
    <div className="min-h-screen bg-[var(--color-surface-warm)]">
      <Header
        cartCount={totalCartItems}
        onCartClick={() => {}} // Already on cart page
        onLogoClick={() => navigate('/')}
        user={user}
        onLogout={logout}
        onProductClick={(id) => navigate(`/product/${id}`)}
      />
      <CartPage onProductClick={(id) => navigate(`/product/${id}`)} />
      <MobileNavigation
        currentPage="cart"
        onNavigate={(page) => {
          if (page === 'home') navigate('/');
          else if (page === 'cart') {} // Already here
          else if (page === 'categories') navigate('/categories');
          else if (page === 'account') navigate('/profile');
        }}
        cartCount={totalCartItems}
      />
    </div>
  );
}

// Wrapper component for CategoriesListPage
function CategoriesListPageWrapper() {
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const { cart } = useCart();
  const totalCartItems = cart?.item_count || 0;

  return (
    <div className="min-h-screen bg-[var(--color-surface-warm)]">
      <Header
        cartCount={totalCartItems}
        onCartClick={() => navigate('/cart')}
        onLogoClick={() => navigate('/')}
        user={user}
        onLogout={logout}
        onProductClick={(id) => navigate(`/product/${id}`)}
      />
      <CategoriesListPage />
      <MobileNavigation
        currentPage="categories"
        onNavigate={(page) => {
          if (page === 'home') navigate('/');
          else if (page === 'cart') navigate('/cart');
          else if (page === 'categories') navigate('/categories');
          else if (page === 'account') navigate('/profile');
        }}
        cartCount={totalCartItems}
      />
    </div>
  );
}

// Wrapper component for ProductsPage with proper navigation
function ProductsPageWrapper() {
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const { cart } = useCart();
  const totalCartItems = cart?.item_count || 0;

  return (
    <div className="min-h-screen bg-[var(--color-surface-warm)]">
      <Header
        cartCount={totalCartItems}
        onCartClick={() => navigate('/cart')}
        onLogoClick={() => navigate('/')}
        user={user}
        onLogout={logout}
        onProductClick={(id) => navigate(`/product/${id}`)}
      />
      <ProductsPage onProductClick={(id) => navigate(`/product/${id}`)} />
      <MobileNavigation
        currentPage="categories"
        onNavigate={(page) => {
          if (page === 'home') navigate('/');
          else if (page === 'cart') navigate('/cart');
          else if (page === 'categories') navigate('/categories');
          else if (page === 'account') navigate('/profile');
        }}
        cartCount={totalCartItems}
      />
    </div>
  );
}

// Wrapper component for ProductDetailPage with URL params
function ProductDetailPageWrapper() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { addItem, cart } = useCart();
  const { isAuthenticated, user, logout } = useAuth();
  const totalCartItems = cart?.item_count || 0;

  const handleProductClick = (productId: string) => {
    navigate(`/product/${productId}`);
  };

  const handleAddToCart = async (productId: string) => {
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }
    try {
      await addItem(parseInt(productId), 1);
    } catch (err) {
      console.error('Error adding to cart:', err);
    }
  };

  if (!id) {
    navigate('/');
    return null;
  }

  return (
    <div className="min-h-screen bg-[var(--color-surface-warm)]">
      <Header
        cartCount={totalCartItems}
        onCartClick={() => navigate('/cart')}
        onLogoClick={() => navigate('/')}
        user={user}
        onLogout={logout}
        onProductClick={handleProductClick}
      />
      <ProductDetailPage
        productId={id}
        onProductClick={handleProductClick}
        onAddToCart={handleAddToCart}
      />
      <MobileNavigation
        currentPage="home"
        onNavigate={(page) => {
          if (page === 'home') navigate('/');
          else if (page === 'cart') navigate('/cart');
          else if (page === 'categories') navigate('/categories');
          else if (page === 'account') navigate('/profile');
        }}
        cartCount={totalCartItems}
      />
    </div>
  );
}

function AppContent() {
  const [currentPage, setCurrentPage] = useState<Page>('home');
  const [selectedProductId, setSelectedProductId] = useState<string | null>(null);
  const [hasUserActivity, setHasUserActivity] = useState(false);
  const [lastViewedProduct, setLastViewedProduct] = useState<string | null>(null);

  const { user, isAuthenticated, logout } = useAuth();
  const { cart, addItem } = useCart();
  const navigate = useNavigate();

  const useStaticBanner = true;

  // Load user activity from user status
  useEffect(() => {
    if (user && user.user_status === 'warm') {
      setHasUserActivity(true);
    }
  }, [user]);

  const handleProductClick = (productId: string) => {
    setSelectedProductId(productId);
    setCurrentPage('product');
    setHasUserActivity(true);
    setLastViewedProduct(productId);
  };

  const handleAddToCart = async (productId: string) => {
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }
    try {
      await addItem(parseInt(productId), 1);
    } catch (err) {
      console.error('Error adding to cart:', err);
    }
  };

  const handleNavigate = (page: Page) => {
    if (page === 'account' && !isAuthenticated) {
      navigate('/login');
      return;
    }
    setCurrentPage(page);
    if (page === 'home') {
      setSelectedProductId(null);
    }
  };

  const totalCartItems = cart?.item_count || 0;

  useEffect(() => {
    const isMobile = window.innerWidth < 768;
    if (isMobile) {
      document.body.style.paddingBottom = '64px';
    } else {
      document.body.style.paddingBottom = '0';
    }

    const handleResize = () => {
      const isMobile = window.innerWidth < 768;
      if (isMobile) {
        document.body.style.paddingBottom = '64px';
      } else {
        document.body.style.paddingBottom = '0';
      }
    };

    window.addEventListener('resize', handleResize);
    return () => {
      window.removeEventListener('resize', handleResize);
      document.body.style.paddingBottom = '0';
    };
  }, []);

  return (
    <div className="min-h-screen bg-[var(--color-surface-warm)]">
      {/* Header */}
      <Header
        cartCount={totalCartItems}
        onCartClick={() => setCurrentPage('cart')}
        onLogoClick={() => handleNavigate('home')}
        user={user}
        onLogout={logout}
        onProductClick={handleProductClick}
      />

      {/* Main Content */}
      <main>
        {currentPage === 'home' && (
          <HomePage
            onProductClick={handleProductClick}
            hasUserActivity={hasUserActivity}
            lastViewedProduct={lastViewedProduct || undefined}
            useStaticBanner={useStaticBanner}
          />
        )}

        {currentPage === 'product' && selectedProductId && (
          <ProductDetailPage
            productId={selectedProductId}
            onProductClick={handleProductClick}
            onAddToCart={handleAddToCart}
          />
        )}

        {currentPage === 'cart' && (
          <CartPage
            onProductClick={handleProductClick}
          />
        )}

        {currentPage === 'categories' && (
          <CategoryPage onProductClick={handleProductClick} />
        )}

        {currentPage === 'account' && isAuthenticated && (
          <div className="min-h-screen bg-[var(--color-surface-warm)]">
            <div className="container py-12">
              <div className="max-w-md mx-auto bg-white rounded-2xl border border-[var(--color-border)] p-6 shadow-[var(--shadow-md)]">
                <h1 className="mb-6 font-serif">Account</h1>
                <div className="space-y-4">
                  <div className="flex items-center gap-4 p-5 bg-gradient-to-br from-[var(--color-primary-light)] to-pink-50 rounded-xl border border-[var(--color-primary)]/20">
                    <div className="w-16 h-16 bg-gradient-to-br from-[var(--color-primary)] to-[var(--color-rose-gold)] rounded-full flex items-center justify-center text-white shadow-[var(--shadow-md)]">
                      <span className="text-xl">{user?.name?.charAt(0).toUpperCase() || 'U'}</span>
                    </div>
                    <div>
                      <h4>{user?.name || 'User'}</h4>
                      <p className="text-xs text-[var(--color-text-muted)]">{user?.email}</p>
                      <span className={`inline-block mt-1 px-2 py-0.5 text-xs rounded-full ${
                        user?.role === 'admin'
                          ? 'bg-purple-100 text-purple-700'
                          : 'bg-blue-100 text-blue-700'
                      }`}>
                        {user?.role === 'admin' ? 'Admin' : 'User'}
                      </span>
                    </div>
                  </div>

                  <div className="border-t border-[var(--color-border)] pt-4 space-y-2">
                    {user?.role === 'admin' && (
                      <button
                        onClick={() => navigate('/admin')}
                        className="w-full text-left px-4 py-3 bg-purple-50 hover:bg-purple-100 rounded-xl transition-colors text-purple-700 font-medium"
                      >
                        Admin Dashboard
                      </button>
                    )}
                    <button
                      onClick={() => navigate('/orders')}
                      className="w-full text-left px-4 py-3 hover:bg-[var(--color-surface)] rounded-xl transition-colors"
                    >
                      My Orders
                    </button>
                    <button
                      onClick={() => navigate('/wishlist')}
                      className="w-full text-left px-4 py-3 hover:bg-[var(--color-surface)] rounded-xl transition-colors"
                    >
                      Wishlist
                    </button>
                    <button
                      onClick={() => navigate('/profile')}
                      className="w-full text-left px-4 py-3 hover:bg-[var(--color-surface)] rounded-xl transition-colors"
                    >
                      Account Settings
                    </button>
                  </div>

                  <div className="border-t border-[var(--color-border)] pt-4">
                    <div className={`p-4 rounded-xl border-2 ${
                      user?.user_status === 'warm'
                        ? 'bg-green-50 border-green-200'
                        : 'bg-gray-50 border-gray-200'
                    }`}>
                      <h4 className={user?.user_status === 'warm' ? 'text-green-700' : 'text-gray-700'}>AI Status</h4>
                      <p className="text-xs text-[var(--color-text-secondary)] mt-2 leading-relaxed">
                        {user?.user_status === 'warm' ? (
                          <>
                            <span className="inline-block w-2 h-2 bg-green-500 rounded-full mr-2"></span>
                            <strong>Warm User</strong> - Receiving personalized recommendations from AI Model
                          </>
                        ) : (
                          <>
                            <span className="inline-block w-2 h-2 bg-gray-400 rounded-full mr-2"></span>
                            <strong>Cold User</strong> - Receiving popular recommendations from Popular Engine
                          </>
                        )}
                      </p>
                    </div>
                  </div>

                  <button
                    onClick={logout}
                    className="w-full py-3 text-red-600 hover:bg-red-50 rounded-xl transition-colors font-medium"
                  >
                    Sign Out
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}
      </main>

      {/* Mobile Bottom Navigation */}
      <MobileNavigation
        currentPage={currentPage === 'product' ? 'home' : currentPage}
        onNavigate={handleNavigate}
        cartCount={totalCartItems}
      />

      {/* Sticky Buy Button on Product Detail (Mobile) */}
      {currentPage === 'product' && (
        <div className="md:hidden fixed bottom-16 left-0 right-0 bg-white border-t border-[var(--color-border)] p-4 z-40 shadow-[var(--shadow-lg)]">
          <button
            onClick={() => selectedProductId && handleAddToCart(selectedProductId)}
            className="w-full py-3.5 bg-[var(--color-primary)] text-white rounded-full hover:bg-[var(--color-primary-hover)] transition-all shadow-[var(--shadow-md)]"
          >
            Add to Cart
          </button>
        </div>
      )}
    </div>
  );
}

export default function App() {
  return (
    <Router>
      <AuthProvider>
        <CartProvider>
          <WishlistProvider>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route path="/checkout" element={<CheckoutPage />} />
            <Route path="/payment/result" element={<PaymentResultPage />} />
            <Route path="/orders" element={<OrderHistoryPage />} />
            <Route path="/orders/:id" element={<OrderDetailPage />} />
            <Route path="/profile" element={<ProfilePage />} />
            <Route path="/wishlist" element={<WishlistPage />} />
            <Route path="/cart" element={<CartPageWrapper />} />
            <Route path="/product/:id" element={<ProductDetailPageWrapper />} />
            <Route path="/products" element={<ProductsPageWrapper />} />
            <Route path="/categories" element={<CategoriesListPageWrapper />} />

            {/* Admin Routes */}
            <Route path="/admin" element={<AdminLayout />}>
              <Route index element={<DashboardPage />} />
              <Route path="orders" element={<AdminOrdersPage />} />
              <Route path="users" element={<UsersPage />} />
              <Route path="products" element={<AdminProductsPage />} />
              <Route path="categories" element={<AdminCategoriesPage />} />
              <Route path="coupons" element={<CouponsPage />} />
            </Route>

            <Route path="/*" element={<AppContent />} />
          </Routes>
          </WishlistProvider>
        </CartProvider>
      </AuthProvider>
    </Router>
  );
}
