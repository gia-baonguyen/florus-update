import { useState } from 'react';
import { Trash2, Plus, Minus, ShoppingBag, Zap, Gift, Loader2, Tag, X, Check } from 'lucide-react';
import { ProductImage } from '../components/ProductImage';
import { useCart } from '../context/CartContext';
import { useAuth } from '../context/AuthContext';
import { Link, useNavigate } from 'react-router-dom';
import { couponsApi, Coupon, ValidateCouponResponse } from '../api/coupons';

interface CartPageProps {
  onProductClick: (productId: string) => void;
}

export function CartPage({ onProductClick }: CartPageProps) {
  const navigate = useNavigate();
  const { cart, loading, error, crossSellProducts, addItem, updateItem, removeItem } = useCart();
  const { isAuthenticated } = useAuth();

  // Coupon state
  const [couponCode, setCouponCode] = useState('');
  const [couponLoading, setCouponLoading] = useState(false);
  const [couponError, setCouponError] = useState<string | null>(null);
  const [appliedCoupon, setAppliedCoupon] = useState<ValidateCouponResponse | null>(null);

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(price / 1000);
  };

  // Not logged in
  if (!isAuthenticated) {
    return (
      <div className="min-h-screen bg-[var(--color-surface-warm)]">
        <div className="container py-12">
          <div className="max-w-md mx-auto text-center bg-white rounded-2xl p-12 shadow-[var(--shadow-md)] border border-[var(--color-border)]">
            <div className="w-20 h-20 mx-auto mb-4 bg-[var(--color-primary-light)] rounded-full flex items-center justify-center">
              <ShoppingBag className="w-10 h-10 text-[var(--color-primary)]" />
            </div>
            <h2 className="mb-2 font-serif">Please Sign In</h2>
            <p className="text-[var(--color-text-secondary)] mb-6 leading-relaxed">
              Sign in to view and manage your cart
            </p>
            <Link
              to="/login"
              className="inline-block px-8 py-3 bg-[var(--color-primary)] text-white rounded-full hover:bg-[var(--color-primary-dark)] transition-colors"
            >
              Sign In
            </Link>
          </div>
        </div>
      </div>
    );
  }

  // Loading state
  if (loading && !cart) {
    return (
      <div className="min-h-screen bg-[var(--color-surface-warm)] flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="w-10 h-10 text-[var(--color-primary)] animate-spin mx-auto mb-4" />
          <p className="text-gray-500">Loading cart...</p>
        </div>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className="min-h-screen bg-[var(--color-surface-warm)]">
        <div className="container py-12">
          <div className="max-w-md mx-auto text-center bg-white rounded-2xl p-12 shadow-[var(--shadow-md)] border border-[var(--color-border)]">
            <div className="w-20 h-20 mx-auto mb-4 bg-red-100 rounded-full flex items-center justify-center">
              <span className="text-3xl">😢</span>
            </div>
            <h2 className="mb-2 font-serif">An Error Occurred</h2>
            <p className="text-[var(--color-text-secondary)] mb-6">{error}</p>
            <button
              onClick={() => window.location.reload()}
              className="px-8 py-3 bg-[var(--color-primary)] text-white rounded-full"
            >
              Try Again
            </button>
          </div>
        </div>
      </div>
    );
  }

  // Empty cart
  if (!cart || !cart.items || cart.items.length === 0) {
    return (
      <div className="min-h-screen bg-[var(--color-surface-warm)]">
        <div className="container py-12">
          <div className="max-w-md mx-auto text-center bg-white rounded-2xl p-12 shadow-[var(--shadow-md)] border border-[var(--color-border)]">
            <div className="w-20 h-20 mx-auto mb-4 bg-[var(--color-primary-light)] rounded-full flex items-center justify-center">
              <ShoppingBag className="w-10 h-10 text-[var(--color-primary)]" />
            </div>
            <h2 className="mb-2 font-serif">Your Cart is Empty</h2>
            <p className="text-[var(--color-text-secondary)] mb-6 leading-relaxed">
              Add your favorite beauty products to the cart to continue shopping
            </p>
          </div>
        </div>
      </div>
    );
  }

  const handleAddCrossSell = async (productId: number) => {
    try {
      await addItem(productId, 1);
    } catch (err) {
      console.error('Error adding item:', err);
    }
  };

  const handleApplyCoupon = async () => {
    if (!couponCode.trim() || !cart) return;

    setCouponLoading(true);
    setCouponError(null);

    try {
      const result = await couponsApi.validateCoupon(couponCode.trim(), cart.subtotal);
      setAppliedCoupon(result);
      setCouponCode('');
    } catch (err: any) {
      setCouponError(err.response?.data?.error || 'Invalid coupon code');
      setAppliedCoupon(null);
    } finally {
      setCouponLoading(false);
    }
  };

  const handleRemoveCoupon = () => {
    setAppliedCoupon(null);
    setCouponError(null);
  };

  const finalTotal = appliedCoupon
    ? cart!.total - appliedCoupon.discount
    : cart?.total || 0;

  return (
    <div className="min-h-screen bg-[var(--color-surface-warm)]">
      <div className="container py-6">
        <h1 className="mb-6 font-serif">Your Cart</h1>

        <div className="grid lg:grid-cols-3 gap-6">
          {/* Cart Items - Left Column */}
          <div className="lg:col-span-2 space-y-4">
            {cart.items.map((item) => {
              const product = item.product;
              if (!product) return null;

              return (
                <div
                  key={item.id}
                  className="bg-white rounded-xl border border-[var(--color-border)] p-4 hover:shadow-[var(--shadow-md)] transition-shadow"
                >
                  <div className="flex gap-4">
                    {/* Product Image */}
                    <div
                      onClick={() => onProductClick(product.id.toString())}
                      className="flex-none w-24 h-24 rounded-xl overflow-hidden bg-[var(--color-surface-warm)] cursor-pointer hover:opacity-75 transition-opacity border border-[var(--color-border)]"
                    >
                      <ProductImage
                        imageUrl={product.image_url}
                        alt={product.name}
                        size="small"
                      />
                    </div>

                    {/* Product Info */}
                    <div className="flex-1 min-w-0">
                      <p className="text-xs text-[var(--color-text-muted)] uppercase tracking-wide mb-1">
                        {product.brand}
                      </p>
                      <h4
                        onClick={() => onProductClick(product.id.toString())}
                        className="cursor-pointer hover:text-[var(--color-primary)] transition-colors line-clamp-2 mb-2"
                      >
                        {product.name}
                      </h4>
                      <p className="text-[var(--color-primary)]">
                        {formatPrice(item.price)}
                      </p>
                    </div>

                    {/* Quantity Controls - Desktop */}
                    <div className="hidden md:flex flex-col items-end justify-between">
                      <button
                        onClick={() => removeItem(product.id)}
                        disabled={loading}
                        className="p-2 hover:bg-red-50 rounded-lg transition-colors group disabled:opacity-50"
                        aria-label="Remove item"
                      >
                        <Trash2 className="w-5 h-5 text-[var(--color-text-muted)] group-hover:text-red-500" />
                      </button>

                      <div className="flex items-center gap-2">
                        <button
                          onClick={() => updateItem(product.id, Math.max(1, item.quantity - 1))}
                          disabled={loading || item.quantity <= 1}
                          className="w-8 h-8 rounded-full border-2 border-[var(--color-border)] hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-light)] transition-all flex items-center justify-center disabled:opacity-50"
                        >
                          <Minus className="w-4 h-4" />
                        </button>
                        <span className="w-8 text-center">{item.quantity}</span>
                        <button
                          onClick={() => updateItem(product.id, item.quantity + 1)}
                          disabled={loading}
                          className="w-8 h-8 rounded-full border-2 border-[var(--color-border)] hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-light)] transition-all flex items-center justify-center disabled:opacity-50"
                        >
                          <Plus className="w-4 h-4" />
                        </button>
                      </div>
                    </div>
                  </div>

                  {/* Quantity Controls - Mobile */}
                  <div className="md:hidden flex items-center justify-between mt-4 pt-4 border-t border-[var(--color-border)]">
                    <button
                      onClick={() => removeItem(product.id)}
                      disabled={loading}
                      className="flex items-center gap-2 text-red-500 hover:text-red-600 disabled:opacity-50"
                    >
                      <Trash2 className="w-4 h-4" />
                      Remove
                    </button>

                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => updateItem(product.id, Math.max(1, item.quantity - 1))}
                        disabled={loading || item.quantity <= 1}
                        className="w-8 h-8 rounded-full border-2 border-[var(--color-border)] hover:border-[var(--color-primary)] transition-all flex items-center justify-center disabled:opacity-50"
                      >
                        <Minus className="w-4 h-4" />
                      </button>
                      <span className="w-8 text-center">{item.quantity}</span>
                      <button
                        onClick={() => updateItem(product.id, item.quantity + 1)}
                        disabled={loading}
                        className="w-8 h-8 rounded-full border-2 border-[var(--color-border)] hover:border-[var(--color-primary)] transition-all flex items-center justify-center disabled:opacity-50"
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
              <h3 className="mb-4 font-serif">Order Summary</h3>

              {/* Coupon Input */}
              <div className="mb-4 pb-4 border-b border-[var(--color-border)]">
                <div className="flex items-center gap-2 mb-2">
                  <Tag className="w-4 h-4 text-[var(--color-primary)]" />
                  <span className="text-sm font-medium">Coupon Code</span>
                </div>

                {appliedCoupon ? (
                  <div className="flex items-center justify-between p-3 bg-green-50 rounded-lg border border-green-200">
                    <div className="flex items-center gap-2">
                      <Check className="w-4 h-4 text-green-600" />
                      <div>
                        <span className="text-sm font-medium text-green-700">{appliedCoupon.coupon.code}</span>
                        <p className="text-xs text-green-600">
                          Save {formatPrice(appliedCoupon.discount)}
                        </p>
                      </div>
                    </div>
                    <button
                      onClick={handleRemoveCoupon}
                      className="p-1 hover:bg-green-100 rounded"
                    >
                      <X className="w-4 h-4 text-green-600" />
                    </button>
                  </div>
                ) : (
                  <div className="flex gap-2">
                    <input
                      type="text"
                      value={couponCode}
                      onChange={(e) => setCouponCode(e.target.value.toUpperCase())}
                      placeholder="Enter coupon code"
                      className="flex-1 px-3 py-2 text-sm border border-[var(--color-border)] rounded-lg focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)] focus:border-transparent"
                    />
                    <button
                      onClick={handleApplyCoupon}
                      disabled={couponLoading || !couponCode.trim()}
                      className="px-4 py-2 bg-[var(--color-primary)] text-white text-sm rounded-lg hover:bg-[var(--color-primary-hover)] transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      {couponLoading ? (
                        <Loader2 className="w-4 h-4 animate-spin" />
                      ) : (
                        'Apply'
                      )}
                    </button>
                  </div>
                )}

                {couponError && (
                  <p className="mt-2 text-xs text-red-500">{couponError}</p>
                )}
              </div>

              <div className="space-y-3 mb-4 pb-4 border-b border-[var(--color-border)]">
                <div className="flex items-center justify-between">
                  <span className="text-[var(--color-text-secondary)] text-sm">
                    Subtotal ({cart.item_count} items)
                  </span>
                  <span>{formatPrice(cart.subtotal)}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-[var(--color-text-secondary)] text-sm">Shipping</span>
                  <span>{cart.shipping_fee === 0 ? <span className="text-[var(--color-success)]">Free</span> : formatPrice(cart.shipping_fee)}</span>
                </div>
                {appliedCoupon && (
                  <div className="flex items-center justify-between text-green-600">
                    <span className="text-sm">Discount ({appliedCoupon.coupon.code})</span>
                    <span>-{formatPrice(appliedCoupon.discount)}</span>
                  </div>
                )}
              </div>

              <div className="flex items-center justify-between mb-6">
                <h4>Total</h4>
                <h3 className="text-[var(--color-primary)] font-serif">{formatPrice(finalTotal)}</h3>
              </div>

              <button
                onClick={() => navigate('/checkout', { state: { appliedCoupon } })}
                className="w-full py-3.5 bg-[var(--color-primary)] text-white rounded-full hover:bg-[var(--color-primary-hover)] transition-all shadow-[var(--shadow-md)] hover:shadow-[var(--shadow-glow)]"
              >
                Checkout Now
              </button>

              {cart.subtotal < 500000 && (
                <div className="mt-4 p-3 bg-[var(--color-primary-light)] rounded-xl border border-[var(--color-primary)]/20">
                  <div className="flex items-start gap-2">
                    <Gift className="w-4 h-4 text-[var(--color-primary)] mt-0.5 flex-shrink-0" />
                    <p className="text-xs text-[var(--color-primary)] leading-relaxed">
                      Add <strong>{formatPrice(500000 - cart.subtotal)}</strong> more for free shipping
                    </p>
                  </div>
                </div>
              )}
            </div>

            {/* Cross-sell Recommendations */}
            {crossSellProducts.length > 0 && (
              <div className="bg-white rounded-xl border border-[var(--color-border)] overflow-hidden shadow-[var(--shadow-sm)]">
                <div className="bg-gradient-to-r from-amber-50 to-orange-50 p-5 border-b border-[var(--color-border)]">
                  <div className="flex items-center gap-2 mb-2">
                    <div className="p-2 bg-white rounded-xl shadow-sm">
                      <Zap className="w-5 h-5 text-[var(--color-secondary)]" />
                    </div>
                    <h4 className="text-[var(--color-secondary)] font-serif">Frequently Bought Together</h4>
                  </div>
                  <p className="text-xs text-[var(--color-text-secondary)] leading-relaxed">
                    <strong>Cross-sell AI</strong> - Increase Average Order Value (AOV)
                  </p>
                </div>

                <div className="p-4 space-y-3">
                  {crossSellProducts.map((product) => (
                    <div
                      key={product.id}
                      className="flex gap-3 p-3 rounded-xl border-2 border-dashed border-[var(--color-border)] hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-light)] transition-all"
                    >
                      <div
                        onClick={() => onProductClick(product.id.toString())}
                        className="flex-none w-16 h-16 rounded-lg overflow-hidden bg-[var(--color-surface-warm)] cursor-pointer border border-[var(--color-border)]"
                      >
                        <ProductImage
                          imageUrl={product.image_url}
                          alt={product.name}
                          size="small"
                        />
                      </div>

                      <div className="flex-1 min-w-0">
                        <p className="text-[10px] text-[var(--color-text-muted)] uppercase tracking-wide mb-1">
                          {product.brand}
                        </p>
                        <h4
                          onClick={() => onProductClick(product.id.toString())}
                          className="text-xs cursor-pointer hover:text-[var(--color-primary)] transition-colors line-clamp-2 mb-2"
                        >
                          {product.name}
                        </h4>
                        <div className="flex items-center justify-between">
                          <p className="text-[var(--color-primary)] text-sm">
                            {formatPrice(product.price)}
                          </p>
                          <button
                            onClick={() => handleAddCrossSell(product.id)}
                            disabled={loading}
                            className="flex items-center gap-1 px-3 py-1.5 bg-[var(--color-primary)] text-white rounded-full text-xs hover:bg-[var(--color-primary-hover)] transition-all shadow-sm disabled:opacity-50"
                          >
                            <Plus className="w-3 h-3" />
                            Add
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
          <h3 className="mb-3 font-serif">About Cross-sell Feature</h3>
          <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">
            The system analyzes products in your cart and suggests complementary products based on:
            <strong> (1)</strong> Purchase patterns from order history in Amazon All_Beauty dataset,
            <strong> (2)</strong> <code className="px-2 py-0.5 bg-gray-100 rounded text-xs">FREQUENTLY_BOUGHT_WITH</code> relationships in Knowledge Graph (Neo4j),
            <strong> (3)</strong> Complementary product analysis (e.g., lipstick → lip liner).
            Goal: increase AOV (Average Order Value) through real-time smart recommendations.
          </p>
        </div>
      </div>
    </div>
  );
}
