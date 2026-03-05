import { useState, useEffect, useMemo } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Package, Clock, Truck, CheckCircle, XCircle, Loader2, ArrowLeft, MapPin, CreditCard, FileText, RotateCcw } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { ordersApi } from '../api/products';
import { returnsApi } from '../api/returns';
import { Order } from '../types';
import { ProductImage } from '../components/ProductImage';

const statusConfig: Record<string, { label: string; icon: typeof Package; color: string; bg: string }> = {
  pending: { label: 'Pending', icon: Clock, color: 'text-amber-600', bg: 'bg-amber-100' },
  confirmed: { label: 'Confirmed', icon: CheckCircle, color: 'text-blue-600', bg: 'bg-blue-100' },
  processing: { label: 'Processing', icon: Package, color: 'text-purple-600', bg: 'bg-purple-100' },
  shipping: { label: 'Shipping', icon: Truck, color: 'text-cyan-600', bg: 'bg-cyan-100' },
  delivered: { label: 'Delivered', icon: CheckCircle, color: 'text-green-600', bg: 'bg-green-100' },
  cancelled: { label: 'Cancelled', icon: XCircle, color: 'text-red-600', bg: 'bg-red-100' },
};

export function OrderDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();
  const [order, setOrder] = useState<Order | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [cancelling, setCancelling] = useState(false);
  const [creatingReturn, setCreatingReturn] = useState(false);

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(price / 1000);
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  useEffect(() => {
    if (!isAuthenticated || !id) return;

    const fetchOrder = async () => {
      try {
        const orderData = await ordersApi.getById(parseInt(id));
        setOrder(orderData);
      } catch (err) {
        console.error('Error fetching order:', err);
        setError('Unable to load order details');
      } finally {
        setLoading(false);
      }
    };

    fetchOrder();
  }, [isAuthenticated, id]);

  const handleCancelOrder = async () => {
    if (!order || !confirm('Are you sure you want to cancel this order?')) return;

    setCancelling(true);
    try {
      await ordersApi.cancel(order.id);
      setOrder({ ...order, status: 'cancelled' });
    } catch (err) {
      console.error('Error cancelling order:', err);
      alert('Unable to cancel order');
    } finally {
      setCancelling(false);
    }
  };

  const handleRequestReturn = async () => {
    if (!order || !confirm('Request return for this entire order?')) return;

    setCreatingReturn(true);
    try {
      const itemsPayload = order.items.map((item) => ({
        order_item_id: item.id,
        quantity: item.quantity,
      }));
      await returnsApi.create({
        order_id: order.id,
        items: itemsPayload,
        reason: 'Customer requested return',
      });
      alert('Return request created. Our team will review it soon.');
    } catch (err) {
      console.error('Error creating return:', err);
      alert('Unable to create return request');
    } finally {
      setCreatingReturn(false);
    }
  };

  // Not logged in
  if (!isAuthenticated) {
    return (
      <div className="min-h-screen bg-[var(--color-surface-warm)]">
        <div className="container py-12">
          <div className="max-w-md mx-auto text-center bg-white rounded-2xl p-12 shadow-[var(--shadow-md)] border border-[var(--color-border)]">
            <div className="w-20 h-20 mx-auto mb-4 bg-[var(--color-primary-light)] rounded-full flex items-center justify-center">
              <Package className="w-10 h-10 text-[var(--color-primary)]" />
            </div>
            <h2 className="mb-2 font-serif">Please Sign In</h2>
            <p className="text-[var(--color-text-secondary)] mb-6">
              Sign in to view order details
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

  // Loading
  if (loading) {
    return (
      <div className="min-h-screen bg-[var(--color-surface-warm)] flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="w-10 h-10 text-[var(--color-primary)] animate-spin mx-auto mb-4" />
          <p className="text-gray-500">Loading order details...</p>
        </div>
      </div>
    );
  }

  // Error or not found
  if (error || !order) {
    return (
      <div className="min-h-screen bg-[var(--color-surface-warm)]">
        <div className="container py-12">
          <button
            onClick={() => navigate('/orders')}
            className="flex items-center gap-2 text-[var(--color-text-secondary)] hover:text-[var(--color-primary)] transition-colors mb-6"
          >
            <ArrowLeft className="w-5 h-5" />
            Back to Orders
          </button>
          <div className="max-w-md mx-auto text-center bg-white rounded-2xl p-12 shadow-[var(--shadow-md)] border border-[var(--color-border)]">
            <div className="w-20 h-20 mx-auto mb-4 bg-red-100 rounded-full flex items-center justify-center">
              <XCircle className="w-10 h-10 text-red-600" />
            </div>
            <h2 className="mb-2 font-serif">Order Not Found</h2>
            <p className="text-[var(--color-text-secondary)] mb-6">{error || 'This order does not exist or you do not have permission to view it.'}</p>
            <button
              onClick={() => navigate('/orders')}
              className="px-8 py-3 bg-[var(--color-primary)] text-white rounded-full"
            >
              View All Orders
            </button>
          </div>
        </div>
      </div>
    );
  }

  const status = statusConfig[order.status] || statusConfig.pending;
  const StatusIcon = status.icon;
  const canCancel = order.status === 'pending' || order.status === 'confirmed';
  const canRequestReturn = order.status === 'delivered';

  // Gom các OrderItem cùng product_id lại thành 1 dòng, cộng quantity & total_price
  const groupedItems = useMemo(() => {
    if (!order.items || order.items.length === 0) return [];

    const map = new Map<number, typeof order.items[number]>();

    for (const item of order.items) {
      const key = item.product_id;
      const existing = map.get(key);

      if (existing) {
        existing.quantity += item.quantity;
        existing.total_price += item.total_price;
      } else {
        // clone để tránh mutate trực tiếp item trong state
        map.set(key, { ...item });
      }
    }

    return Array.from(map.values());
  }, [order.items]);

  return (
    <div className="min-h-screen bg-[var(--color-surface-warm)]">
      <div className="container py-6">
        {/* Back button */}
        <button
          onClick={() => navigate('/orders')}
          className="flex items-center gap-2 text-[var(--color-text-secondary)] hover:text-[var(--color-primary)] transition-colors mb-6"
        >
          <ArrowLeft className="w-5 h-5" />
          Back to Orders
        </button>

        {/* Order Header */}
        <div className="bg-white rounded-xl border border-[var(--color-border)] overflow-hidden shadow-[var(--shadow-sm)] mb-6">
          <div className="p-6 border-b border-[var(--color-border)] bg-[var(--color-surface)]">
            <div className="flex flex-wrap items-center justify-between gap-4">
              <div className="flex items-center gap-4">
                <div className={`p-3 rounded-xl ${status.bg}`}>
                  <StatusIcon className={`w-6 h-6 ${status.color}`} />
                </div>
                <div>
                  <h1 className="text-xl font-serif">{order.order_code}</h1>
                  <p className="text-sm text-[var(--color-text-muted)]">{formatDate(order.created_at)}</p>
                </div>
              </div>
              <span className={`px-4 py-2 rounded-full text-sm font-medium ${status.bg} ${status.color}`}>
                {status.label}
              </span>
            </div>
          </div>

          {/* Order Summary */}
          <div className="p-6 grid md:grid-cols-3 gap-6">
            {/* Shipping Address */}
            <div className="flex gap-3">
              <div className="p-2 bg-blue-50 rounded-lg h-fit">
                <MapPin className="w-5 h-5 text-blue-600" />
              </div>
              <div>
                <p className="text-sm font-medium text-[var(--color-text-muted)] mb-1">Shipping Address</p>
                <p className="text-sm">{order.shipping_address || 'Not provided'}</p>
              </div>
            </div>

            {/* Payment Method */}
            <div className="flex gap-3">
              <div className="p-2 bg-green-50 rounded-lg h-fit">
                <CreditCard className="w-5 h-5 text-green-600" />
              </div>
              <div>
                <p className="text-sm font-medium text-[var(--color-text-muted)] mb-1">Payment Method</p>
                <p className="text-sm">
                  {order.payment_method === 'cod' && 'Cash on Delivery (COD)'}
                  {order.payment_method === 'vnpay' && 'VNPay'}
                  {order.payment_method === 'momo' && 'MoMo'}
                  {order.payment_method === 'zalopay' && 'ZaloPay'}
                </p>
              </div>
            </div>

            {/* Note */}
            {order.note && (
              <div className="flex gap-3">
                <div className="p-2 bg-purple-50 rounded-lg h-fit">
                  <FileText className="w-5 h-5 text-purple-600" />
                </div>
                <div>
                  <p className="text-sm font-medium text-[var(--color-text-muted)] mb-1">Note</p>
                  <p className="text-sm">{order.note}</p>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Order Items (đã gom theo product) */}
        <div className="bg-white rounded-xl border border-[var(--color-border)] overflow-hidden shadow-[var(--shadow-sm)] mb-6">
          <div className="p-4 border-b border-[var(--color-border)] bg-[var(--color-surface)]">
            <h2 className="font-serif">Order Items ({groupedItems.length})</h2>
          </div>
          <div className="divide-y divide-[var(--color-border)]">
            {groupedItems.map((item) => (
              <div key={item.id} className="p-4 flex gap-4 hover:bg-[var(--color-surface-warm)] transition-colors">
                <div className="w-20 h-20 rounded-lg overflow-hidden bg-[var(--color-surface-warm)] flex-shrink-0 border border-[var(--color-border)]">
                  <ProductImage imageUrl={item.product?.image_url} alt={item.product?.name || ''} size="small" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="font-medium mb-1">{item.product?.name}</p>
                  <p className="text-sm text-[var(--color-text-muted)]">
                    Unit price: {formatPrice(item.unit_price)}
                  </p>
                  <p className="text-sm text-[var(--color-text-muted)]">
                    Quantity: {item.quantity}
                  </p>
                </div>
                <div className="text-right">
                  <p className="text-lg font-serif text-[var(--color-primary)]">{formatPrice(item.total_price)}</p>
                </div>
              </div>
            ))}
          </div>
        </div>

            {/* Order Total */}
        <div className="bg-white rounded-xl border border-[var(--color-border)] overflow-hidden shadow-[var(--shadow-sm)]">
          <div className="p-6">
            <div className="space-y-3">
              <div className="flex justify-between text-sm">
                <span className="text-[var(--color-text-muted)]">Subtotal</span>
                <span>{formatPrice(order.subtotal)}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-[var(--color-text-muted)]">Shipping Fee</span>
                <span>{order.shipping_fee === 0 ? 'Free' : formatPrice(order.shipping_fee)}</span>
              </div>
              {order.discount > 0 && (
                <div className="flex justify-between text-sm text-green-600">
                  <span>Discount</span>
                  <span>-{formatPrice(order.discount)}</span>
                </div>
              )}
              <div className="border-t border-[var(--color-border)] pt-3 flex justify-between">
                <span className="font-medium">Total</span>
                <span className="text-xl font-serif text-[var(--color-primary)]">{formatPrice(order.total)}</span>
              </div>
            </div>

            {/* Actions */}
            <div className="mt-6 pt-6 border-t border-[var(--color-border)] space-y-3">
              {canCancel && (
                <button
                  onClick={handleCancelOrder}
                  disabled={cancelling}
                  className="w-full py-3 text-red-600 border-2 border-red-200 rounded-full hover:bg-red-50 transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
                >
                  {cancelling ? (
                    <>
                      <Loader2 className="w-5 h-5 animate-spin" />
                      Cancelling...
                    </>
                  ) : (
                    <>
                      <XCircle className="w-5 h-5" />
                      Cancel Order
                    </>
                  )}
                </button>
              )}

              {canRequestReturn && (
                <button
                  onClick={handleRequestReturn}
                  disabled={creatingReturn}
                  className="w-full py-3 text-[var(--color-primary)] border-2 border-[var(--color-primary-light)] rounded-full hover:bg-[var(--color-primary-light)] transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
                >
                  {creatingReturn ? (
                    <>
                      <Loader2 className="w-5 h-5 animate-spin" />
                      Creating return...
                    </>
                  ) : (
                    <>
                      <RotateCcw className="w-5 h-5" />
                      Request Return/Refund
                    </>
                  )}
                </button>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
