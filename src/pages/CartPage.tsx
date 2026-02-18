import { Trash2, Plus, Minus, ShoppingBag, Zap, Gift } from 'lucide-react';
import { ProductImage } from '../components/ProductImage';
import { products, crossSellProducts } from '../data/products';

interface CartItem {
  productId: string;
  quantity: number;
}

interface CartPageProps {
  cartItems: CartItem[];
  onUpdateQuantity: (productId: string, quantity: number) => void;
  onRemoveItem: (productId: string) => void;
  onProductClick: (productId: string) => void;
  onAddToCart: (productId: string) => void;
}

export function CartPage({
  cartItems,
  onUpdateQuantity,
  onRemoveItem,
  onProductClick,
  onAddToCart,
}: CartPageProps) {
  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('vi-VN', {
      style: 'currency',
      currency: 'VND',
    }).format(price);
  };

  // Calculate totals
  const subtotal = cartItems.reduce((sum, item) => {
    const product = products.find(p => p.id === item.productId);
    return sum + (product?.price || 0) * item.quantity;
  }, 0);

  const shipping = subtotal > 0 ? (subtotal > 5000000 ? 0 : 50000) : 0;
  const total = subtotal + shipping;

  // Get cross-sell recommendations
  const crossSellIds = new Set<string>();
  cartItems.forEach(item => {
    const recommendations = crossSellProducts[item.productId] || [];
    recommendations.forEach(id => {
      // Only show products not already in cart
      if (!cartItems.find(ci => ci.productId === id)) {
        crossSellIds.add(id);
      }
    });
  });

  const crossSellProductList = Array.from(crossSellIds)
    .map(id => products.find(p => p.id === id))
    .filter(Boolean)
    .slice(0, 3);

  if (cartItems.length === 0) {
    return (
      <div className="min-h-screen bg-[var(--color-surface-warm)]">
        <div className="container py-12">
          <div className="max-w-md mx-auto text-center bg-white rounded-2xl p-12 shadow-[var(--shadow-md)] border border-[var(--color-border)]">
            <div className="w-20 h-20 mx-auto mb-4 bg-[var(--color-primary-light)] rounded-full flex items-center justify-center">
              <ShoppingBag className="w-10 h-10 text-[var(--color-primary)]" />
            </div>
            <h2 className="mb-2 font-serif">Giỏ hàng trống</h2>
            <p className="text-[var(--color-text-secondary)] mb-6 leading-relaxed">
              Hãy thêm những sản phẩm làm đẹp yêu thích vào giỏ hàng để tiếp tục mua sắm
            </p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[var(--color-surface-warm)]">
      <div className="container py-6">
        <h1 className="mb-6 font-serif">Giỏ hàng của bạn</h1>

        <div className="grid lg:grid-cols-3 gap-6">
          {/* Cart Items - Left Column */}
          <div className="lg:col-span-2 space-y-4">
            {cartItems.map((item) => {
              const product = products.find(p => p.id === item.productId);
              if (!product) return null;

              return (
                <div
                  key={item.productId}
                  className="bg-white rounded-xl border border-[var(--color-border)] p-4 hover:shadow-[var(--shadow-md)] transition-shadow"
                >
                  <div className="flex gap-4">
                    {/* Product Image */}
                    <div
                      onClick={() => onProductClick(product.id)}
                      className="flex-none w-24 h-24 rounded-xl overflow-hidden bg-[var(--color-surface-warm)] cursor-pointer hover:opacity-75 transition-opacity border border-[var(--color-border)]"
                    >
                      <ProductImage
                        gradient={product.imageGradient || ''}
                        emoji={product.imageEmoji || '📦'}
                        alt={product.name}
                        className="w-full h-full scale-75"
                      />
                    </div>

                    {/* Product Info */}
                    <div className="flex-1 min-w-0">
                      <p className="text-xs text-[var(--color-text-muted)] uppercase tracking-wide mb-1">
                        {product.brand}
                      </p>
                      <h4
                        onClick={() => onProductClick(product.id)}
                        className="cursor-pointer hover:text-[var(--color-primary)] transition-colors line-clamp-2 mb-2"
                      >
                        {product.name}
                      </h4>
                      <p className="text-[var(--color-primary)]">
                        {formatPrice(product.price)}
                      </p>
                    </div>

                    {/* Quantity Controls - Desktop */}
                    <div className="hidden md:flex flex-col items-end justify-between">
                      <button
                        onClick={() => onRemoveItem(product.id)}
                        className="p-2 hover:bg-red-50 rounded-lg transition-colors group"
                        aria-label="Remove item"
                      >
                        <Trash2 className="w-5 h-5 text-[var(--color-text-muted)] group-hover:text-red-500" />
                      </button>

                      <div className="flex items-center gap-2">
                        <button
                          onClick={() => onUpdateQuantity(product.id, Math.max(1, item.quantity - 1))}
                          className="w-8 h-8 rounded-full border-2 border-[var(--color-border)] hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-light)] transition-all flex items-center justify-center"
                        >
                          <Minus className="w-4 h-4" />
                        </button>
                        <span className="w-8 text-center">{item.quantity}</span>
                        <button
                          onClick={() => onUpdateQuantity(product.id, item.quantity + 1)}
                          className="w-8 h-8 rounded-full border-2 border-[var(--color-border)] hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-light)] transition-all flex items-center justify-center"
                        >
                          <Plus className="w-4 h-4" />
                        </button>
                      </div>
                    </div>
                  </div>

                  {/* Quantity Controls - Mobile */}
                  <div className="md:hidden flex items-center justify-between mt-4 pt-4 border-t border-[var(--color-border)]">
                    <button
                      onClick={() => onRemoveItem(product.id)}
                      className="flex items-center gap-2 text-red-500 hover:text-red-600"
                    >
                      <Trash2 className="w-4 h-4" />
                      Xóa
                    </button>

                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => onUpdateQuantity(product.id, Math.max(1, item.quantity - 1))}
                        className="w-8 h-8 rounded-full border-2 border-[var(--color-border)] hover:border-[var(--color-primary)] transition-all flex items-center justify-center"
                      >
                        <Minus className="w-4 h-4" />
                      </button>
                      <span className="w-8 text-center">{item.quantity}</span>
                      <button
                        onClick={() => onUpdateQuantity(product.id, item.quantity + 1)}
                        className="w-8 h-8 rounded-full border-2 border-[var(--color-border)] hover:border-[var(--color-primary)] transition-all flex items-center justify-center"
                      >
                        <Plus className="w-4 h-4" />
                      </button>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>

          {/* Right Column - Summary & Cross-sell */}
          <div className="space-y-4">
            {/* Order Summary */}
            <div className="bg-white rounded-xl border border-[var(--color-border)] p-6 sticky top-20 shadow-[var(--shadow-sm)]">
              <h3 className="mb-4 font-serif">Tóm tắt đơn hàng</h3>

              <div className="space-y-3 mb-4 pb-4 border-b border-[var(--color-border)]">
                <div className="flex items-center justify-between">
                  <span className="text-[var(--color-text-secondary)] text-sm">
                    Tạm tính ({cartItems.reduce((sum, item) => sum + item.quantity, 0)} sản phẩm)
                  </span>
                  <span>{formatPrice(subtotal)}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-[var(--color-text-secondary)] text-sm">Phí vận chuyển</span>
                  <span>{shipping === 0 ? <span className="text-[var(--color-success)]">Miễn phí</span> : formatPrice(shipping)}</span>
                </div>
              </div>

              <div className="flex items-center justify-between mb-6">
                <h4>Tổng cộng</h4>
                <h3 className="text-[var(--color-primary)] font-serif">{formatPrice(total)}</h3>
              </div>

              <button className="w-full py-3.5 bg-[var(--color-primary)] text-white rounded-full hover:bg-[var(--color-primary-hover)] transition-all shadow-[var(--shadow-md)] hover:shadow-[var(--shadow-glow)]">
                Thanh toán ngay
              </button>

              {subtotal < 5000000 && (
                <div className="mt-4 p-3 bg-[var(--color-primary-light)] rounded-xl border border-[var(--color-primary)]/20">
                  <div className="flex items-start gap-2">
                    <Gift className="w-4 h-4 text-[var(--color-primary)] mt-0.5 flex-shrink-0" />
                    <p className="text-xs text-[var(--color-primary)] leading-relaxed">
                      Mua thêm <strong>{formatPrice(5000000 - subtotal)}</strong> để được miễn phí vận chuyển
                    </p>
                  </div>
                </div>
              )}
            </div>

            {/* Cross-sell Recommendations */}
            {crossSellProductList.length > 0 && (
              <div className="bg-white rounded-xl border border-[var(--color-border)] overflow-hidden shadow-[var(--shadow-sm)]">
                <div className="bg-gradient-to-r from-amber-50 to-orange-50 p-5 border-b border-[var(--color-border)]">
                  <div className="flex items-center gap-2 mb-2">
                    <div className="p-2 bg-white rounded-xl shadow-sm">
                      <Zap className="w-5 h-5 text-[var(--color-secondary)]" />
                    </div>
                    <h4 className="text-[var(--color-secondary)] font-serif">Thường được mua cùng</h4>
                  </div>
                  <p className="text-xs text-[var(--color-text-secondary)] leading-relaxed">
                    <strong>Cross-sell AI</strong> - Tăng giá trị đơn hàng (AOV)
                  </p>
                </div>

                <div className="p-4 space-y-3">
                  {crossSellProductList.map((product) => (
                    <div
                      key={product.id}
                      className="flex gap-3 p-3 rounded-xl border-2 border-dashed border-[var(--color-border)] hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-light)] transition-all"
                    >
                      <div
                        onClick={() => onProductClick(product.id)}
                        className="flex-none w-16 h-16 rounded-lg overflow-hidden bg-[var(--color-surface-warm)] cursor-pointer border border-[var(--color-border)]"
                      >
                        <ProductImage
                          gradient={product.imageGradient || ''}
                          emoji={product.imageEmoji || '📦'}
                          alt={product.name}
                          className="w-full h-full scale-75"
                          size="small"
                        />
                      </div>

                      <div className="flex-1 min-w-0">
                        <p className="text-[10px] text-[var(--color-text-muted)] uppercase tracking-wide mb-1">
                          {product.brand}
                        </p>
                        <h4
                          onClick={() => onProductClick(product.id)}
                          className="text-xs cursor-pointer hover:text-[var(--color-primary)] transition-colors line-clamp-2 mb-2"
                        >
                          {product.name}
                        </h4>
                        <div className="flex items-center justify-between">
                          <p className="text-[var(--color-primary)] text-sm">
                            {formatPrice(product.price)}
                          </p>
                          <button
                            onClick={() => onAddToCart(product.id)}
                            className="flex items-center gap-1 px-3 py-1.5 bg-[var(--color-primary)] text-white rounded-full text-xs hover:bg-[var(--color-primary-hover)] transition-all shadow-sm"
                          >
                            <Plus className="w-3 h-3" />
                            Thêm
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Tech Info */}
        <div className="bg-white rounded-2xl border border-[var(--color-border)] p-6 mt-6 shadow-[var(--shadow-sm)]">
          <h3 className="mb-3 font-serif">Về tính năng Cross-sell</h3>
          <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">
            Hệ thống phân tích các sản phẩm trong giỏ hàng và đề xuất các sản phẩm bổ sung dựa trên:
            <strong> (1)</strong> Pattern mua hàng từ lịch sử đơn hàng trong Amazon All_Beauty dataset,
            <strong> (2)</strong> Mối quan hệ <code className="px-2 py-0.5 bg-gray-100 rounded text-xs">FREQUENTLY_BOUGHT_WITH</code> trong Knowledge Graph (Neo4j),
            <strong> (3)</strong> Phân tích sản phẩm bổ trợ (ví dụ: son môi → chì kẻ viền môi). 
            Mục tiêu tăng AOV (Average Order Value) thông qua gợi ý thông minh theo thời gian thực.
          </p>
        </div>
      </div>
    </div>
  );
}