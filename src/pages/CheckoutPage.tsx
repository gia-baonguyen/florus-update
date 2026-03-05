import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { MapPin, FileText, Loader2, ShoppingBag, Check, ArrowLeft, Truck, CreditCard, Tag, X, Wallet, Banknote, Percent } from 'lucide-react';
import { useCart } from '../context/CartContext';
import { useAuth } from '../context/AuthContext';
import { ordersApi } from '../api/products';
import { couponsApi, ValidateCouponResponse, Coupon } from '../api/coupons';
import { paymentsApi, PaymentMethod } from '../api/payments';
import { addressesApi } from '../api/addresses';
import { shippingApi } from '../api/shipping';
import { loyaltyApi } from '../api/loyalty';
import type { UserAddress, ShippingMethod } from '../types';
import { ProductImage } from '../components/ProductImage';
import { useEventTracking } from '../hooks/useEventTracking';

const paymentMethods = [
  { id: 'cod' as PaymentMethod, name: 'Cash on Delivery', icon: Banknote, description: 'Pay when you receive' },
  { id: 'zalopay' as PaymentMethod, name: 'ZaloPay', icon: Wallet, description: 'Pay with ZaloPay e-wallet' },
  { id: 'momo' as PaymentMethod, name: 'MoMo', icon: Wallet, description: 'Pay with MoMo e-wallet' },
  { id: 'vnpay' as PaymentMethod, name: 'VNPay', icon: CreditCard, description: 'Pay with bank card via VNPay' },
];

export function CheckoutPage() {
  const navigate = useNavigate();
  const { cart, loading: cartLoading, refreshCart } = useCart();
  const { isAuthenticated } = useAuth();
  const [shippingAddress, setShippingAddress] = useState('');
  const [addresses, setAddresses] = useState<UserAddress[]>([]);
  const [selectedAddressId, setSelectedAddressId] = useState<number | null>(null);
  const [shippingMethods, setShippingMethods] = useState<ShippingMethod[]>([]);
  const [selectedShippingMethod, setSelectedShippingMethod] = useState<string>('');
  const [note, setNote] = useState('');
  const [selectedPaymentMethod, setSelectedPaymentMethod] = useState<PaymentMethod>('cod');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [orderSuccess, setOrderSuccess] = useState<{ orderCode: string; orderId: number } | null>(null);

  // Loyalty
  const [loyaltyPoints, setLoyaltyPoints] = useState<number>(0);
  const [loyaltyTier, setLoyaltyTier] = useState<string>('');
  const [pointsToUse, setPointsToUse] = useState<number>(0);

  // Coupon states
  const [couponCode, setCouponCode] = useState('');
  const [appliedCoupon, setAppliedCoupon] = useState<ValidateCouponResponse | null>(null);
  const [couponError, setCouponError] = useState<string | null>(null);
  const [isApplyingCoupon, setIsApplyingCoupon] = useState(false);
  const [availableCoupons, setAvailableCoupons] = useState<Coupon[]>([]);
  const [showCouponList, setShowCouponList] = useState(false);
  const { trackPurchase } = useEventTracking();

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(price / 1000);
  };

  const applyCouponCode = async (code: string) => {
    setIsApplyingCoupon(true);
    setCouponError(null);

    try {
      const result = await couponsApi.validateCoupon(code.toUpperCase(), cart?.subtotal || 0);
      setAppliedCoupon(result);
      setCouponError(null);
    } catch (err: any) {
      setCouponError(err.response?.data?.message || 'Invalid coupon code');
      setAppliedCoupon(null);
    } finally {
      setIsApplyingCoupon(false);
    }
  };

  const handleApplyCoupon = async () => {
    if (!couponCode.trim()) {
      setCouponError('Please enter a coupon code');
      return;
    }
    await applyCouponCode(couponCode.trim());
  };

  const handleRemoveCoupon = () => {
    setAppliedCoupon(null);
    setCouponCode('');
    setCouponError(null);
  };

  // Fetch available coupons for dropdown list, similar to Shopee
  useEffect(() => {
    const fetchCoupons = async () => {
      if (!cart || !cart.subtotal) return;
      try {
        const coupons = await couponsApi.getAvailable(cart.subtotal);
        setAvailableCoupons(coupons);
      } catch (err) {
        console.error('Failed to fetch available coupons', err);
      }
    };
    fetchCoupons();
  }, [cart?.subtotal]);

  // Fetch addresses, shipping methods, and loyalty info
  useEffect(() => {
    const fetchData = async () => {
      if (!isAuthenticated) return;
      try {
        // Addresses
        const addrList = await addressesApi.getMyAddresses();
        setAddresses(addrList);
        const defaultAddr = addrList.find((a) => a.is_default) || addrList[0];
        if (defaultAddr) {
          setSelectedAddressId(defaultAddr.id);
          setShippingAddress(
            `${defaultAddr.full_name}, ${defaultAddr.phone}, ${defaultAddr.street}, ${defaultAddr.city}${
              defaultAddr.state ? ', ' + defaultAddr.state : ''
            }${defaultAddr.postal_code ? ', ' + defaultAddr.postal_code : ''}`
          );
        }
      } catch (err) {
        console.warn('Failed to fetch addresses', err);
      }

      try {
        const methods = await shippingApi.getMethods();
        setShippingMethods(methods);
        if (methods.length > 0) {
          setSelectedShippingMethod(methods[0].code);
        }
      } catch (err) {
        console.warn('Failed to fetch shipping methods', err);
      }

      try {
        const loyalty = await loyaltyApi.getMyLoyalty();
        setLoyaltyPoints(loyalty.points);
        setLoyaltyTier(loyalty.tier);
      } catch (err) {
        console.warn('Failed to fetch loyalty info', err);
      }
    };
    fetchData();
  }, [isAuthenticated]);

  // Calculate final total with coupon + loyalty discount (client-side preview)
  const couponDiscount = appliedCoupon ? appliedCoupon.discount : 0;
  const loyaltyDiscount = Math.min(pointsToUse, loyaltyPoints) * 1000;
  const baseSubtotal = cart?.subtotal || 0;
  const baseShipping = cart?.shipping_fee || 0;
  const finalTotal = baseSubtotal + baseShipping - couponDiscount - loyaltyDiscount;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!shippingAddress.trim()) {
      setError('Please enter a shipping address');
      return;
    }

    setIsSubmitting(true);
    setError(null);

    try {
      const order = await ordersApi.create(
        shippingAddress,
        note,
        appliedCoupon?.coupon.code,
        selectedPaymentMethod,
        selectedAddressId ?? undefined,
        selectedShippingMethod || undefined,
        pointsToUse > 0 ? pointsToUse : undefined
      );

      // If online payment method, redirect to payment gateway
      if (selectedPaymentMethod !== 'cod') {
        try {
          const paymentResponse = await paymentsApi.createPayment({
            order_code: order.order_code,
            payment_method: selectedPaymentMethod,
          });
          // Redirect to payment gateway
          window.location.href = paymentResponse.payment_url;
          return;
        } catch (paymentErr: any) {
          console.error('Error creating payment:', paymentErr);
          // Order created but payment failed - still show success for COD fallback
          setError('Payment gateway unavailable. Your order has been placed as Cash on Delivery.');
        }
      }

      // Log purchase events for each order item (if available)
      if (order.items && order.items.length > 0) {
        order.items.forEach((item) => {
          trackPurchase(item.product_id, order.id, item.quantity, item.unit_price);
        });
      }

      setOrderSuccess({ orderCode: order.order_code, orderId: order.id });
      await refreshCart();
    } catch (err: any) {
      console.error('Error creating order:', err);
      setError(err.response?.data?.message || 'An error occurred while placing the order');
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
            <h2 className="mb-2 font-serif">Please Sign In</h2>
            <p className="text-[var(--color-text-secondary)] mb-6">
              Sign in to proceed with checkout
            </p>
            <button
              onClick={() => navigate('/login')}
              className="px-8 py-3 bg-[var(--color-primary)] text-white rounded-full hover:bg-[var(--color-primary-dark)] transition-colors"
            >
              Sign In
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
            <h2 className="mb-2 font-serif text-green-700">Order Placed Successfully!</h2>
            <p className="text-[var(--color-text-secondary)] mb-2">
              Thank you for shopping at Florus Beauty
            </p>
            <p className="text-lg font-medium mb-6">
              Order Code: <span className="text-[var(--color-primary)]">{orderSuccess.orderCode}</span>
            </p>
            <div className="flex flex-col gap-3">
              <button
                onClick={() => navigate('/orders')}
                className="w-full px-6 py-3 bg-[var(--color-primary)] text-white rounded-full hover:bg-[var(--color-primary-dark)] transition-colors"
              >
                View Orders
              </button>
              <button
                onClick={() => navigate('/')}
                className="w-full px-6 py-3 border-2 border-[var(--color-border)] rounded-full hover:border-[var(--color-primary)] transition-colors"
              >
                Continue Shopping
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
          <p className="text-gray-500">Loading...</p>
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
            <p className="text-[var(--color-text-secondary)] mb-6">
              Please add products to your cart before checkout
            </p>
            <button
              onClick={() => navigate('/')}
              className="px-8 py-3 bg-[var(--color-primary)] text-white rounded-full hover:bg-[var(--color-primary-dark)] transition-colors"
            >
              Continue Shopping
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
          Go Back
        </button>

        <h1 className="mb-6 font-serif">Checkout</h1>

        <div className="grid lg:grid-cols-3 gap-6">
          {/* Checkout Form - Left Column */}
          <div className="lg:col-span-2 space-y-6">
            {/* Shipping Address */}
            <div className="bg-white rounded-xl border border-[var(--color-border)] p-6 shadow-[var(--shadow-sm)]">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2 bg-[var(--color-primary-light)] rounded-xl">
                  <MapPin className="w-5 h-5 text-[var(--color-primary)]" />
                </div>
                <h3 className="font-serif">Shipping Address</h3>
              </div>

              {/* Saved addresses selector */}
              {addresses.length > 0 && (
                <div className="mb-4 flex flex-col gap-2">
                  <label className="text-xs text-[var(--color-text-secondary)]">
                    Choose from your saved addresses
                  </label>
                  <select
                    value={selectedAddressId ?? ''}
                    onChange={(e) => {
                      const id = Number(e.target.value);
                      setSelectedAddressId(id || null);
                      const addr = addresses.find((a) => a.id === id);
                      if (addr) {
                        setShippingAddress(
                          `${addr.full_name}, ${addr.phone}, ${addr.street}, ${addr.city}${
                            addr.state ? ', ' + addr.state : ''
                          }${addr.postal_code ? ', ' + addr.postal_code : ''}`
                        );
                      }
                    }}
                    className="w-full p-3 border-2 border-[var(--color-border)] rounded-xl focus:outline-none focus:border-[var(--color-primary)] text-sm"
                  >
                    <option value="">-- Select address --</option>
                    {addresses.map((addr) => (
                      <option key={addr.id} value={addr.id}>
                        {addr.full_name} - {addr.city} {addr.is_default ? '(Default)' : ''}
                      </option>
                    ))}
                  </select>
                </div>
              )}

              <textarea
                value={shippingAddress}
                onChange={(e) => setShippingAddress(e.target.value)}
                placeholder="Enter your full shipping address (Street, City, State, ZIP Code)"
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
                <h3 className="font-serif">Note</h3>
              </div>
              <textarea
                value={note}
                onChange={(e) => setNote(e.target.value)}
                placeholder="Order notes (optional)"
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
                <h3 className="font-serif">Coupon Code</h3>
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
                          ? `${appliedCoupon.coupon.discount_value}% off`
                          : `${formatPrice(appliedCoupon.coupon.discount_value)} off`}
                        {appliedCoupon.coupon.max_discount_amount > 0 &&
                          appliedCoupon.coupon.discount_type === 'percent' &&
                          ` (max ${formatPrice(appliedCoupon.coupon.max_discount_amount)})`}
                      </p>
                      <p className="text-sm font-medium text-green-700 mt-1">
                        You save: {formatPrice(appliedCoupon.discount)}
                      </p>
                    </div>
                    <button
                      onClick={handleRemoveCoupon}
                      className="p-2 text-gray-400 hover:text-red-500 transition-colors"
                      title="Remove coupon"
                    >
                      <X className="w-5 h-5" />
                    </button>
                  </div>
                </div>
              ) : (
                // Coupon input + dropdown list
                <div>
                  <div className="flex gap-3 mb-2">
                    <input
                      type="text"
                      value={couponCode}
                      onChange={(e) => {
                        setCouponCode(e.target.value.toUpperCase());
                        setCouponError(null);
                      }}
                      placeholder="Enter coupon code"
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
                        'Apply'
                      )}
                    </button>
                  </div>

                  {/* Shopee-style coupon list */}
                  {availableCoupons.length > 0 && (
                    <>
                      <button
                        type="button"
                        onClick={() => setShowCouponList((prev) => !prev)}
                        className="inline-flex items-center gap-1 text-[11px] text-[var(--color-primary)] hover:text-[var(--color-primary-hover)]"
                      >
                        <Percent className="w-3 h-3" />
                        <span>{showCouponList ? 'Hide available coupons' : 'Choose from available coupons'}</span>
                      </button>

                      {showCouponList && (
                        <div className="mt-2 space-y-2 max-h-40 overflow-y-auto rounded-lg border border-[var(--color-border)] bg-[var(--color-surface)] p-2">
                          {availableCoupons.map((c) => (
                            <button
                              key={c.id}
                              type="button"
                              onClick={() => {
                                setCouponCode(c.code.toUpperCase());
                                setShowCouponList(false);
                                applyCouponCode(c.code);
                              }}
                              className="w-full flex items-center justify-between px-3 py-2 text-xs text-left rounded-md hover:bg-white"
                            >
                              <div>
                                <span className="font-semibold mr-2">{c.code}</span>
                                <span className="text-[var(--color-text-secondary)]">
                                  {c.discount_type === 'percent'
                                    ? `${c.discount_value}% off`
                                    : `Save ${formatPrice(c.discount_value * 1000)}`}
                                </span>
                                {c.min_order_amount > 0 && (
                                  <p className="text-[10px] text-[var(--color-text-muted)] mt-0.5">
                                    Min order {formatPrice(c.min_order_amount)}
                                  </p>
                                )}
                              </div>
                              <span className="text-[var(--color-primary)] text-[11px] font-medium">Select</span>
                            </button>
                          ))}
                        </div>
                      )}
                    </>
                  )}

                  {couponError && (
                    <p className="mt-2 text-sm text-red-500">{couponError}</p>
                  )}
                </div>
              )}
            </div>

            {/* Payment Method Selection */}
            <div className="bg-white rounded-xl border border-[var(--color-border)] p-6 shadow-[var(--shadow-sm)]">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2 bg-purple-100 rounded-xl">
                  <CreditCard className="w-5 h-5 text-purple-600" />
                </div>
                <h3 className="font-serif">Payment Method</h3>
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                {paymentMethods.map((method) => {
                  const Icon = method.icon;
                  const isSelected = selectedPaymentMethod === method.id;
                  return (
                    <button
                      key={method.id}
                      type="button"
                      onClick={() => setSelectedPaymentMethod(method.id)}
                      className={`p-4 rounded-xl border-2 transition-all text-left ${
                        isSelected
                          ? 'border-[var(--color-primary)] bg-[var(--color-primary-light)]'
                          : 'border-[var(--color-border)] hover:border-[var(--color-primary-light)]'
                      }`}
                    >
                      <div className="flex items-center gap-3">
                        <div className={`p-2 rounded-lg ${
                          isSelected ? 'bg-[var(--color-primary)] text-white' : 'bg-gray-100 text-gray-600'
                        }`}>
                          <Icon className="w-5 h-5" />
                        </div>
                        <div>
                          <p className={`font-medium ${isSelected ? 'text-[var(--color-primary)]' : ''}`}>
                            {method.name}
                          </p>
                          <p className="text-xs text-[var(--color-text-muted)]">{method.description}</p>
                        </div>
                        {isSelected && (
                          <Check className="w-5 h-5 text-[var(--color-primary)] ml-auto" />
                        )}
                      </div>
                    </button>
                  );
                })}
              </div>
            </div>

            {/* Shipping Method Selection */}
            <div className="bg-white rounded-xl border border-[var(--color-border)] p-6 shadow-[var(--shadow-sm)]">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2 bg-sky-100 rounded-xl">
                  <Truck className="w-5 h-5 text-sky-600" />
                </div>
                <h3 className="font-serif">Shipping Method</h3>
              </div>
              {shippingMethods.length === 0 ? (
                <p className="text-sm text-[var(--color-text-muted)]">
                  Standard shipping will be applied.
                </p>
              ) : (
                <div className="flex flex-col gap-2">
                  {shippingMethods.map((method) => (
                    <label
                      key={method.code}
                      className={`flex items-center justify-between p-3 rounded-xl border-2 cursor-pointer ${
                        selectedShippingMethod === method.code
                          ? 'border-[var(--color-primary)] bg-[var(--color-primary-light)]'
                          : 'border-[var(--color-border)] hover:border-[var(--color-primary-light)]'
                      }`}
                    >
                      <div className="flex flex-col">
                        <span className="font-medium text-sm">{method.name}</span>
                        {method.description && (
                          <span className="text-xs text-[var(--color-text-muted)]">
                            {method.description}
                          </span>
                        )}
                      </div>
                      <input
                        type="radio"
                        className="ml-3"
                        checked={selectedShippingMethod === method.code}
                        onChange={() => setSelectedShippingMethod(method.code)}
                      />
                    </label>
                  ))}
                </div>
              )}
            </div>

            {/* Cart Items Summary */}
            <div className="bg-white rounded-xl border border-[var(--color-border)] p-6 shadow-[var(--shadow-sm)]">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2 bg-blue-100 rounded-xl">
                  <ShoppingBag className="w-5 h-5 text-blue-600" />
                </div>
                <h3 className="font-serif">Products ({cart.item_count})</h3>
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
                  <h4 className="text-blue-700 mb-1">Shipping Information</h4>
                  <p className="text-sm text-blue-600 leading-relaxed">
                    Orders will be processed within 1-2 business days.
                    {cart.shipping_fee === 0
                      ? ' Free shipping for orders over $500!'
                      : ' Nationwide delivery within 3-5 days.'}
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* Order Summary - Right Column */}
          <div className="space-y-4">
            <div className="bg-white rounded-xl border border-[var(--color-border)] p-6 sticky top-20 shadow-[var(--shadow-sm)]">
              <h3 className="mb-4 font-serif">Order Summary</h3>

              <div className="space-y-3 mb-4 pb-4 border-b border-[var(--color-border)]">
                <div className="flex items-center justify-between">
                  <span className="text-[var(--color-text-secondary)] text-sm">
                    Subtotal ({cart.item_count} items)
                  </span>
                  <span>{formatPrice(cart.subtotal)}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-[var(--color-text-secondary)] text-sm">Shipping</span>
                  <span>
                    {cart.shipping_fee === 0 ? (
                      <span className="text-[var(--color-success)]">Free</span>
                    ) : (
                      formatPrice(cart.shipping_fee)
                    )}
                  </span>
                </div>
                {appliedCoupon && (
                  <div className="flex items-center justify-between text-green-600">
                    <span className="text-sm flex items-center gap-1">
                      <Tag className="w-4 h-4" />
                      Discount ({appliedCoupon.coupon.code})
                    </span>
                    <span>-{formatPrice(appliedCoupon.discount)}</span>
                  </div>
                )}
              </div>

              {/* Loyalty summary */}
              <div className="space-y-2 mb-4 pb-4 border-b border-[var(--color-border)]">
                <div className="flex items-center justify-between">
                  <span className="text-[var(--color-text-secondary)] text-sm">Loyalty</span>
                  <span className="text-sm">
                    {loyaltyTier ? `${loyaltyTier} • ${loyaltyPoints} pts` : 'No loyalty points yet'}
                  </span>
                </div>
                {loyaltyPoints > 0 && (
                  <div className="flex items-center gap-2">
                    <input
                      type="number"
                      min={0}
                      max={loyaltyPoints}
                      value={pointsToUse}
                      onChange={(e) => {
                        const val = Number(e.target.value) || 0;
                        setPointsToUse(Math.max(0, Math.min(val, loyaltyPoints)));
                      }}
                      className="w-24 p-2 border-2 border-[var(--color-border)] rounded-lg text-sm focus:outline-none focus:border-[var(--color-primary)]"
                      placeholder="0"
                    />
                    <span className="text-xs text-[var(--color-text-muted)]">
                      ≈ {formatPrice(pointsToUse * 1000)} discount
                    </span>
                    <button
                      type="button"
                      className="ml-auto text-xs text-[var(--color-primary)]"
                      onClick={() => setPointsToUse(loyaltyPoints)}
                    >
                      Use max
                    </button>
                  </div>
                )}
              </div>

              <div className="flex items-center justify-between mb-6">
                <h4>Total</h4>
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
                    Processing...
                  </>
                ) : (
                  <>
                    <CreditCard className="w-5 h-5" />
                    Place Order
                  </>
                )}
              </button>

              <p className="text-xs text-center text-[var(--color-text-muted)] mt-4">
                By placing an order, you agree to our Terms of Service and Privacy Policy
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
