import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { MapPin, FileText, Loader2, ShoppingBag, Check, ArrowLeft, Truck, CreditCard, Tag, X } from 'lucide-react';
import { useCart } from '../context/CartContext';
import { useAuth } from '../context/AuthContext';
import { ordersApi } from '../api/products';
import { couponsApi, ValidateCouponResponse } from '../api/coupons';
import { ProductImage } from '../components/ProductImage';

export function CheckoutPage() {
  const navigate = useNavigate();
  const { cart, loading: cartLoading, refreshCart } = useCart();
  const { isAuthenticated } = useAuth();
  const [shippingAddress, setShippingAddress] = useState('');
  const [note, setNote] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [orderSuccess, setOrderSuccess] = useState<{ orderCode: string; orderId: number } | null>(null);

  // Coupon states
  const [couponCode, setCouponCode] = useState('');
  const [appliedCoupon, setAppliedCoupon] = useState<ValidateCouponResponse | null>(null);
  const [couponError, setCouponError] = useState<string | null>(null);
  const [isApplyingCoupon, setIsApplyingCoupon] = useState(false);

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('vi-VN', {
      style: 'currency',
      currency: 'VND',
    }).format(price);
  };

  const handleApplyCoupon = async () => {
    if (!couponCode.trim()) {
      setCouponError('Vui lòng nhập mã giảm giá');
      return;
    }

    setIsApplyingCoupon(true);
    setCouponError(null);

    try {
      const result = await couponsApi.validateCoupon(couponCode.trim().toUpperCase(), cart?.subtotal || 0);
      setAppliedCoupon(result);
      setCouponError(null);
    } catch (err: any) {
      setCouponError(err.response?.data?.message || 'Mã giảm giá không hợp lệ');
      setAppliedCoupon(null);
    } finally {
      setIsApplyingCoupon(false);
    }
  };

  const handleRemoveCoupon = () => {
    setAppliedCoupon(null);
    setCouponCode('');
    setCouponError(null);
  };

  // Calculate final total with discount
  const finalTotal = appliedCoupon
    ? (cart?.subtotal || 0) + (cart?.shipping_fee || 0) - appliedCoupon.discount
    : cart?.total || 0;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!shippingAddress.trim()) {
      setError('Vui lòng nhập địa chỉ giao hàng');
      return;
    }

    setIsSubmitting(true);
    setError(null);

    try {
      const order = await ordersApi.create(
        shippingAddress,
        note,
        appliedCoupon?.coupon.code
      );
      setOrderSuccess({ orderCode: order.order_code, orderId: order.id });
      await refreshCart();
    } catch (err: any) {
      console.error('Error creating order:', err);
      setError(err.response?.data?.message || 'Có lỗi xảy ra khi đặt hàng');
    } finally {
      setIsSubmitting(false);
    }
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
            <h2 className="mb-2 font-serif">Vui lòng đăng nhập</h2>
            <p className="text-[var(--color-text-secondary)] mb-6">
              Đăng nhập để tiến hành thanh toán
            </p>
            <button
              onClick={() => navigate('/login')}
              className="px-8 py-3 bg-[var(--color-primary)] text-white rounded-full hover:bg-[var(--color-primary-dark)] transition-colors"
            >
              Đăng nhập
            </button>
          </div>
        </div>
      </div>
    );
  }

  // Order success
  if (orderSuccess) {
    return (
      <div className="min-h-screen bg-[var(--color-surface-warm)]">
        <div className="container py-12">
          <div className="max-w-md mx-auto text-center bg-white rounded-2xl p-12 shadow-[var(--shadow-md)] border border-[var(--color-border)]">
            <div className="w-20 h-20 mx-auto mb-4 bg-green-100 rounded-full flex items-center justify-center">
              <Check className="w-10 h-10 text-green-600" />
            </div>
            <h2 className="mb-2 font-serif text-green-700">Đặt hàng thành công!</h2>
            <p className="text-[var(--color-text-secondary)] mb-2">
              Cảm ơn bạn đã mua hàng tại Florus Beauty
            </p>
            <p className="text-lg font-medium mb-6">
              Mã đơn hàng: <span className="text-[var(--color-primary)]">{orderSuccess.orderCode}</span>
            </p>
            <div className="flex flex-col gap-3">
              <button
                onClick={() => navigate('/orders')}
                className="w-full px-6 py-3 bg-[var(--color-primary)] text-white rounded-full hover:bg-[var(--color-primary-dark)] transition-colors"
              >
                Xem đơn hàng
              </button>
              <button
                onClick={() => navigate('/')}
                className="w-full px-6 py-3 border-2 border-[var(--color-border)] rounded-full hover:border-[var(--color-primary)] transition-colors"
              >
                Tiếp tục mua sắm
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Loading cart
  if (cartLoading && !cart) {
    return (
      <div className="min-h-screen bg-[var(--color-surface-warm)] flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="w-10 h-10 text-[var(--color-primary)] animate-spin mx-auto mb-4" />
          <p className="text-gray-500">Đang tải...</p>
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
            <h2 className="mb-2 font-serif">Giỏ hàng trống</h2>
            <p className="text-[var(--color-text-secondary)] mb-6">
              Vui lòng thêm sản phẩm vào giỏ hàng trước khi thanh toán
            </p>
            <button
              onClick={() => navigate('/')}
              className="px-8 py-3 bg-[var(--color-primary)] text-white rounded-full hover:bg-[var(--color-primary-dark)] transition-colors"
            >
              Tiếp tục mua sắm
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[var(--color-surface-warm)]">
      <div className="container py-6">
        {/* Back button */}
        <button
          onClick={() => navigate(-1)}
          className="flex items-center gap-2 text-[var(--color-text-secondary)] hover:text-[var(--color-primary)] transition-colors mb-6"
        >
          <ArrowLeft className="w-5 h-5" />
          Quay lại
        </button>

        <h1 className="mb-6 font-serif">Thanh toán</h1>

        <div className="grid lg:grid-cols-3 gap-6">
          {/* Checkout Form - Left Column */}
          <div className="lg:col-span-2 space-y-6">
            {/* Shipping Address */}
            <div className="bg-white rounded-xl border border-[var(--color-border)] p-6 shadow-[var(--shadow-sm)]">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2 bg-[var(--color-primary-light)] rounded-xl">
                  <MapPin className="w-5 h-5 text-[var(--color-primary)]" />
                </div>
                <h3 className="font-serif">Địa chỉ giao hàng</h3>
              </div>
              <textarea
                value={shippingAddress}
                onChange={(e) => setShippingAddress(e.target.value)}
                placeholder="Nhập địa chỉ giao hàng đầy đủ (Số nhà, đường, phường/xã, quận/huyện, tỉnh/thành phố)"
                autoComplete="street-address"
                className="w-full p-4 border-2 border-[var(--color-border)] rounded-xl focus:outline-none focus:border-[var(--color-primary)] resize-none transition-colors"
                rows={3}
                required
              />
            </div>

            {/* Note */}
            <div className="bg-white rounded-xl border border-[var(--color-border)] p-6 shadow-[var(--shadow-sm)]">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2 bg-amber-100 rounded-xl">
                  <FileText className="w-5 h-5 text-amber-600" />
                </div>
                <h3 className="font-serif">Ghi chú</h3>
              </div>
              <textarea
                value={note}
                onChange={(e) => setNote(e.target.value)}
                placeholder="Ghi chú cho đơn hàng (không bắt buộc)"
                className="w-full p-4 border-2 border-[var(--color-border)] rounded-xl focus:outline-none focus:border-[var(--color-primary)] resize-none transition-colors"
                rows={2}
              />
            </div>

            {/* Coupon Code */}
            <div className="bg-white rounded-xl border border-[var(--color-border)] p-6 shadow-[var(--shadow-sm)]">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2 bg-green-100 rounded-xl">
                  <Tag className="w-5 h-5 text-green-600" />
                </div>
                <h3 className="font-serif">Mã giảm giá</h3>
              </div>

              {appliedCoupon ? (
                // Applied coupon display
                <div className="p-4 bg-green-50 border border-green-200 rounded-xl">
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="flex items-center gap-2 mb-1">
                        <span className="font-medium text-green-700">{appliedCoupon.coupon.code}</span>
                        <Check className="w-4 h-4 text-green-600" />
                      </div>
                      <p className="text-sm text-green-600">
                        {appliedCoupon.coupon.discount_type === 'percent'
                          ? `Giảm ${appliedCoupon.coupon.discount_value}%`
                          : `Giảm ${formatPrice(appliedCoupon.coupon.discount_value)}`}
                        {appliedCoupon.coupon.max_discount_amount > 0 &&
                          appliedCoupon.coupon.discount_type === 'percent' &&
                          ` (tối đa ${formatPrice(appliedCoupon.coupon.max_discount_amount)})`}
                      </p>
                      <p className="text-sm font-medium text-green-700 mt-1">
                        Bạn tiết kiệm được: {formatPrice(appliedCoupon.discount)}
                      </p>
                    </div>
                    <button
                      onClick={handleRemoveCoupon}
                      className="p-2 text-gray-400 hover:text-red-500 transition-colors"
                      title="Xóa mã giảm giá"
                    >
                      <X className="w-5 h-5" />
                    </button>
                  </div>
                </div>
              ) : (
                // Coupon input
                <div>
                  <div className="flex gap-3">
                    <input
                      type="text"
                      value={couponCode}
                      onChange={(e) => {
                        setCouponCode(e.target.value.toUpperCase());
                        setCouponError(null);
                      }}
                      placeholder="Nhập mã giảm giá"
                      className="flex-1 p-3 border-2 border-[var(--color-border)] rounded-xl focus:outline-none focus:border-[var(--color-primary)] transition-colors uppercase"
                    />
                    <button
                      onClick={handleApplyCoupon}
                      disabled={isApplyingCoupon || !couponCode.trim()}
                      className="px-6 py-3 bg-[var(--color-primary)] text-white rounded-xl hover:bg-[var(--color-primary-dark)] transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                    >
                      {isApplyingCoupon ? (
                        <Loader2 className="w-4 h-4 animate-spin" />
                      ) : (
                        'Áp dụng'
                      )}
                    </button>
                  </div>
                  {couponError && (
                    <p className="mt-2 text-sm text-red-500">{couponError}</p>
                  )}
                </div>
              )}
            </div>

            {/* Cart Items Summary */}
            <div className="bg-white rounded-xl border border-[var(--color-border)] p-6 shadow-[var(--shadow-sm)]">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2 bg-blue-100 rounded-xl">
                  <ShoppingBag className="w-5 h-5 text-blue-600" />
                </div>
                <h3 className="font-serif">Sản phẩm ({cart.item_count})</h3>
              </div>
              <div className="space-y-3">
                {cart.items.map((item) => (
                  <div key={item.id} className="flex gap-3 p-3 bg-[var(--color-surface-warm)] rounded-xl">
                    <div className="w-16 h-16 rounded-lg overflow-hidden bg-white border border-[var(--color-border)] flex-shrink-0">
                      <ProductImage imageUrl={item.product.image_url} alt={item.product.name} size="small" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-xs text-[var(--color-text-muted)] uppercase">{item.product.brand}</p>
                      <p className="text-sm font-medium truncate">{item.product.name}</p>
                      <div className="flex items-center justify-between mt-1">
                        <span className="text-xs text-[var(--color-text-secondary)]">x{item.quantity}</span>
                        <span className="text-sm text-[var(--color-primary)]">{formatPrice(item.total)}</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Shipping Info */}
            <div className="bg-gradient-to-r from-blue-50 to-cyan-50 rounded-xl border border-blue-200 p-5">
              <div className="flex items-start gap-3">
                <Truck className="w-5 h-5 text-blue-600 mt-0.5 flex-shrink-0" />
                <div>
                  <h4 className="text-blue-700 mb-1">Thông tin vận chuyển</h4>
                  <p className="text-sm text-blue-600 leading-relaxed">
                    Đơn hàng sẽ được xử lý trong 1-2 ngày làm việc.
                    {cart.shipping_fee === 0
                      ? ' Miễn phí vận chuyển cho đơn hàng từ 500.000đ!'
                      : ' Giao hàng toàn quốc trong 3-5 ngày.'}
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* Order Summary - Right Column */}
          <div className="space-y-4">
            <div className="bg-white rounded-xl border border-[var(--color-border)] p-6 sticky top-20 shadow-[var(--shadow-sm)]">
              <h3 className="mb-4 font-serif">Tóm tắt đơn hàng</h3>

              <div className="space-y-3 mb-4 pb-4 border-b border-[var(--color-border)]">
                <div className="flex items-center justify-between">
                  <span className="text-[var(--color-text-secondary)] text-sm">
                    Tạm tính ({cart.item_count} sản phẩm)
                  </span>
                  <span>{formatPrice(cart.subtotal)}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-[var(--color-text-secondary)] text-sm">Phí vận chuyển</span>
                  <span>
                    {cart.shipping_fee === 0 ? (
                      <span className="text-[var(--color-success)]">Miễn phí</span>
                    ) : (
                      formatPrice(cart.shipping_fee)
                    )}
                  </span>
                </div>
                {appliedCoupon && (
                  <div className="flex items-center justify-between text-green-600">
                    <span className="text-sm flex items-center gap-1">
                      <Tag className="w-4 h-4" />
                      Giảm giá ({appliedCoupon.coupon.code})
                    </span>
                    <span>-{formatPrice(appliedCoupon.discount)}</span>
                  </div>
                )}
              </div>

              <div className="flex items-center justify-between mb-6">
                <h4>Tổng cộng</h4>
                <h3 className="text-[var(--color-primary)] font-serif">{formatPrice(finalTotal)}</h3>
              </div>

              {error && (
                <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-xl text-red-600 text-sm">
                  {error}
                </div>
              )}

              <button
                onClick={handleSubmit}
                disabled={isSubmitting || !shippingAddress.trim()}
                className="w-full py-3.5 bg-[var(--color-primary)] text-white rounded-full hover:bg-[var(--color-primary-hover)] transition-all shadow-[var(--shadow-md)] hover:shadow-[var(--shadow-glow)] disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
              >
                {isSubmitting ? (
                  <>
                    <Loader2 className="w-5 h-5 animate-spin" />
                    Đang xử lý...
                  </>
                ) : (
                  <>
                    <CreditCard className="w-5 h-5" />
                    Đặt hàng
                  </>
                )}
              </button>

              <p className="text-xs text-center text-[var(--color-text-muted)] mt-4">
                Bằng cách đặt hàng, bạn đồng ý với Điều khoản dịch vụ và Chính sách bảo mật của chúng tôi
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
