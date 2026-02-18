import { useState, useEffect } from 'react';
import { Header } from './components/Header';
import { MobileNavigation } from './components/MobileNavigation';
import { HomePage } from './pages/HomePage';
import { ProductDetailPage } from './pages/ProductDetailPage';
import { CartPage } from './pages/CartPage';

type Page = 'home' | 'product' | 'cart' | 'categories' | 'account';

interface CartItem {
  productId: string;
  quantity: number;
}

export default function App() {
  const [currentPage, setCurrentPage] = useState<Page>('home');
  const [selectedProductId, setSelectedProductId] = useState<string | null>(null);
  const [cartItems, setCartItems] = useState<CartItem[]>([]);
  const [hasUserActivity, setHasUserActivity] = useState(false);
  const [lastViewedProduct, setLastViewedProduct] = useState<string | null>(null);
  
  // Toggle for Figma-friendly static banner (easier to copy to Figma)
  // Set to true when you want to copy design to Figma
  const useStaticBanner = true; // Changed to TRUE by default for better Figma copy experience

  // Simulate user activity when they view a product
  const handleProductClick = (productId: string) => {
    setSelectedProductId(productId);
    setCurrentPage('product');
    setHasUserActivity(true); // User becomes "warm" after viewing a product
    setLastViewedProduct(productId); // Track last viewed for personalization
  };

  const handleAddToCart = (productId: string) => {
    setCartItems((prev) => {
      const existing = prev.find((item) => item.productId === productId);
      if (existing) {
        return prev.map((item) =>
          item.productId === productId
            ? { ...item, quantity: item.quantity + 1 }
            : item
        );
      }
      return [...prev, { productId, quantity: 1 }];
    });
  };

  const handleUpdateQuantity = (productId: string, quantity: number) => {
    setCartItems((prev) =>
      prev.map((item) =>
        item.productId === productId ? { ...item, quantity } : item
      )
    );
  };

  const handleRemoveItem = (productId: string) => {
    setCartItems((prev) => prev.filter((item) => item.productId !== productId));
  };

  const handleNavigate = (page: Page) => {
    setCurrentPage(page);
    if (page === 'home') {
      setSelectedProductId(null);
    }
  };

  const totalCartItems = cartItems.reduce((sum, item) => sum + item.quantity, 0);

  // Add mobile bottom navigation padding
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
            cartItems={cartItems}
            onUpdateQuantity={handleUpdateQuantity}
            onRemoveItem={handleRemoveItem}
            onProductClick={handleProductClick}
            onAddToCart={handleAddToCart}
          />
        )}

        {currentPage === 'categories' && (
          <div className="min-h-screen bg-[var(--color-surface-warm)]">
            <div className="container py-12">
              <h1 className="mb-6 font-serif">Danh mục sản phẩm</h1>
              <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                {[
                  { name: 'Makeup', emoji: '💄', desc: 'Son môi, phấn mắt, mascara' },
                  { name: 'Skincare', emoji: '✨', desc: 'Serum, kem dưỡng, mặt nạ' },
                  { name: 'Fragrance', emoji: '🌸', desc: 'Nước hoa cao cấp' },
                  { name: 'Nails', emoji: '💅', desc: 'Sơn móng tay gel' },
                  { name: 'Hair Care', emoji: '💇', desc: 'Dầu gội, dưỡng tóc' },
                  { name: 'Body Care', emoji: '🧴', desc: 'Kem body, tắm trắng' },
                  { name: 'Tools', emoji: '🔧', desc: 'Cọ trang điểm, dụng cụ' },
                  { name: 'Sets', emoji: '🎁', desc: 'Bộ sản phẩm combo' },
                ].map((category) => (
                  <div
                    key={category.name}
                    className="bg-white rounded-xl border-2 border-[var(--color-border)] p-6 hover:border-[var(--color-primary)] hover:shadow-[var(--shadow-lg)] transition-all cursor-pointer group"
                  >
                    <div className="text-4xl mb-3">{category.emoji}</div>
                    <h4 className="text-center mb-2 group-hover:text-[var(--color-primary)] transition-colors">{category.name}</h4>
                    <p className="text-xs text-center text-[var(--color-text-muted)]">{category.desc}</p>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {currentPage === 'account' && (
          <div className="min-h-screen bg-[var(--color-surface-warm)]">
            <div className="container py-12">
              <div className="max-w-md mx-auto bg-white rounded-2xl border border-[var(--color-border)] p-6 shadow-[var(--shadow-md)]">
                <h1 className="mb-6 font-serif">Tài khoản</h1>
                <div className="space-y-4">
                  <div className="flex items-center gap-4 p-5 bg-gradient-to-br from-[var(--color-primary-light)] to-pink-50 rounded-xl border border-[var(--color-primary)]/20">
                    <div className="w-16 h-16 bg-gradient-to-br from-[var(--color-primary)] to-[var(--color-rose-gold)] rounded-full flex items-center justify-center text-white shadow-[var(--shadow-md)]">
                      <span className="text-xl">U</span>
                    </div>
                    <div>
                      <h4>Beauty Lover</h4>
                      <p className="text-xs text-[var(--color-text-muted)]">beautylover@florus.vn</p>
                    </div>
                  </div>

                  <div className="border-t border-[var(--color-border)] pt-4 space-y-2">
                    <button className="w-full text-left px-4 py-3 hover:bg-[var(--color-surface)] rounded-xl transition-colors">
                      Đơn hàng của tôi
                    </button>
                    <button className="w-full text-left px-4 py-3 hover:bg-[var(--color-surface)] rounded-xl transition-colors">
                      Sản phẩm yêu thích
                    </button>
                    <button className="w-full text-left px-4 py-3 hover:bg-[var(--color-surface)] rounded-xl transition-colors">
                      Cài đặt
                    </button>
                  </div>

                  <div className="border-t border-[var(--color-border)] pt-4">
                    <div className={`p-4 rounded-xl border-2 ${
                      hasUserActivity 
                        ? 'bg-green-50 border-green-200' 
                        : 'bg-gray-50 border-gray-200'
                    }`}>
                      <h4 className={hasUserActivity ? 'text-green-700' : 'text-gray-700'}>Trạng thái người dùng</h4>
                      <p className="text-xs text-[var(--color-text-secondary)] mt-2 leading-relaxed">
                        {hasUserActivity ? (
                          <>
                            <span className="inline-block w-2 h-2 bg-green-500 rounded-full mr-2"></span>
                            <strong>Warm User</strong> - Nhận gợi ý cá nhân hóa từ AI Model
                          </>
                        ) : (
                          <>
                            <span className="inline-block w-2 h-2 bg-gray-400 rounded-full mr-2"></span>
                            <strong>Cold User</strong> - Nhận gợi ý phổ biến từ Popular Engine
                          </>
                        )}
                      </p>
                    </div>
                  </div>
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
            Thêm vào giỏ hàng
          </button>
        </div>
      )}
    </div>
  );
}