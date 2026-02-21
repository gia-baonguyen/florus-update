import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Package, Clock, Truck, CheckCircle, XCircle, Loader2, ChevronRight, ArrowLeft, ShoppingBag } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { ordersApi } from '../api/products';
import { Order } from '../types';
import { ProductImage } from '../components/ProductImage';

const statusConfig: Record<string, { label: string; icon: typeof Package; color: string; bg: string }> = {
  pending: { label: 'Chờ xác nhận', icon: Clock, color: 'text-amber-600', bg: 'bg-amber-100' },
  confirmed: { label: 'Đã xác nhận', icon: CheckCircle, color: 'text-blue-600', bg: 'bg-blue-100' },
  processing: { label: 'Đang xử lý', icon: Package, color: 'text-purple-600', bg: 'bg-purple-100' },
  shipping: { label: 'Đang giao', icon: Truck, color: 'text-cyan-600', bg: 'bg-cyan-100' },
  delivered: { label: 'Đã giao', icon: CheckCircle, color: 'text-green-600', bg: 'bg-green-100' },
  cancelled: { label: 'Đã hủy', icon: XCircle, color: 'text-red-600', bg: 'bg-red-100' },
};

export function OrderHistoryPage() {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [cancellingId, setCancellingId] = useState<number | null>(null);

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('vi-VN', {
      style: 'currency',
      currency: 'VND',
    }).format(price);
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('vi-VN', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  useEffect(() => {
    if (!isAuthenticated) return;

    const fetchOrders = async () => {
      try {
        const { orders: ordersData } = await ordersApi.getAll(1, 50);
        setOrders(ordersData || []);
      } catch (err) {
        console.error('Error fetching orders:', err);
        setError('Không thể tải danh sách đơn hàng');
      } finally {
        setLoading(false);
      }
    };

    fetchOrders();
  }, [isAuthenticated]);

  const handleCancelOrder = async (orderId: number) => {
    if (!confirm('Bạn có chắc chắn muốn hủy đơn hàng này?')) return;

    setCancellingId(orderId);
    try {
      await ordersApi.cancel(orderId);
      setOrders(orders.map((order) =>
        order.id === orderId ? { ...order, status: 'cancelled' } : order
      ));
    } catch (err) {
      console.error('Error cancelling order:', err);
      alert('Không thể hủy đơn hàng');
    } finally {
      setCancellingId(null);
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
            <h2 className="mb-2 font-serif">Vui lòng đăng nhập</h2>
            <p className="text-[var(--color-text-secondary)] mb-6">
              Đăng nhập để xem lịch sử đơn hàng
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

  // Loading
  if (loading) {
    return (
      <div className="min-h-screen bg-[var(--color-surface-warm)] flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="w-10 h-10 text-[var(--color-primary)] animate-spin mx-auto mb-4" />
          <p className="text-gray-500">Đang tải đơn hàng...</p>
        </div>
      </div>
    );
  }

  // Error
  if (error) {
    return (
      <div className="min-h-screen bg-[var(--color-surface-warm)]">
        <div className="container py-12">
          <div className="max-w-md mx-auto text-center bg-white rounded-2xl p-12 shadow-[var(--shadow-md)] border border-[var(--color-border)]">
            <div className="w-20 h-20 mx-auto mb-4 bg-red-100 rounded-full flex items-center justify-center">
              <XCircle className="w-10 h-10 text-red-600" />
            </div>
            <h2 className="mb-2 font-serif">Có lỗi xảy ra</h2>
            <p className="text-[var(--color-text-secondary)] mb-6">{error}</p>
            <button
              onClick={() => window.location.reload()}
              className="px-8 py-3 bg-[var(--color-primary)] text-white rounded-full"
            >
              Thử lại
            </button>
          </div>
        </div>
      </div>
    );
  }

  // Empty orders
  if (orders.length === 0) {
    return (
      <div className="min-h-screen bg-[var(--color-surface-warm)]">
        <div className="container py-12">
          <button
            onClick={() => navigate('/')}
            className="flex items-center gap-2 text-[var(--color-text-secondary)] hover:text-[var(--color-primary)] transition-colors mb-6"
          >
            <ArrowLeft className="w-5 h-5" />
            Quay lại
          </button>

          <div className="max-w-md mx-auto text-center bg-white rounded-2xl p-12 shadow-[var(--shadow-md)] border border-[var(--color-border)]">
            <div className="w-20 h-20 mx-auto mb-4 bg-[var(--color-primary-light)] rounded-full flex items-center justify-center">
              <Package className="w-10 h-10 text-[var(--color-primary)]" />
            </div>
            <h2 className="mb-2 font-serif">Chưa có đơn hàng</h2>
            <p className="text-[var(--color-text-secondary)] mb-6">
              Bạn chưa có đơn hàng nào. Hãy bắt đầu mua sắm!
            </p>
            <button
              onClick={() => navigate('/')}
              className="px-8 py-3 bg-[var(--color-primary)] text-white rounded-full hover:bg-[var(--color-primary-dark)] transition-colors"
            >
              Mua sắm ngay
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
          onClick={() => navigate('/')}
          className="flex items-center gap-2 text-[var(--color-text-secondary)] hover:text-[var(--color-primary)] transition-colors mb-6"
        >
          <ArrowLeft className="w-5 h-5" />
          Quay lại
        </button>

        <h1 className="mb-6 font-serif">Lịch sử đơn hàng</h1>

        <div className="space-y-4">
          {orders.map((order) => {
            const status = statusConfig[order.status] || statusConfig.pending;
            const StatusIcon = status.icon;
            const canCancel = order.status === 'pending' || order.status === 'confirmed';

            return (
              <div
                key={order.id}
                className="bg-white rounded-xl border border-[var(--color-border)] overflow-hidden shadow-[var(--shadow-sm)] hover:shadow-[var(--shadow-md)] transition-shadow"
              >
                {/* Order Header */}
                <div className="flex flex-wrap items-center justify-between gap-3 p-4 border-b border-[var(--color-border)] bg-[var(--color-surface)]">
                  <div className="flex items-center gap-3">
                    <div className={`p-2 rounded-lg ${status.bg}`}>
                      <StatusIcon className={`w-5 h-5 ${status.color}`} />
                    </div>
                    <div>
                      <p className="font-medium">{order.order_code}</p>
                      <p className="text-xs text-[var(--color-text-muted)]">{formatDate(order.created_at)}</p>
                    </div>
                  </div>
                  <span className={`px-3 py-1 rounded-full text-sm font-medium ${status.bg} ${status.color}`}>
                    {status.label}
                  </span>
                </div>

                {/* Order Items */}
                <div className="p-4">
                  <div className="space-y-3 mb-4">
                    {order.items && order.items.slice(0, 3).map((item) => (
                      <div key={item.id} className="flex gap-3">
                        <div className="w-14 h-14 rounded-lg overflow-hidden bg-[var(--color-surface-warm)] flex-shrink-0 border border-[var(--color-border)]">
                          <ProductImage imageUrl={item.product?.image_url} alt={item.product?.name || ''} size="small" />
                        </div>
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium truncate">{item.product?.name}</p>
                          <p className="text-xs text-[var(--color-text-muted)]">
                            {formatPrice(item.unit_price)} x {item.quantity}
                          </p>
                        </div>
                        <p className="text-sm text-[var(--color-primary)]">{formatPrice(item.total_price)}</p>
                      </div>
                    ))}
                    {order.items && order.items.length > 3 && (
                      <p className="text-sm text-[var(--color-text-muted)]">
                        +{order.items.length - 3} sản phẩm khác
                      </p>
                    )}
                  </div>

                  {/* Order Footer */}
                  <div className="flex flex-wrap items-center justify-between gap-3 pt-4 border-t border-[var(--color-border)]">
                    <div>
                      <p className="text-sm text-[var(--color-text-muted)]">Tổng cộng</p>
                      <p className="text-lg font-serif text-[var(--color-primary)]">{formatPrice(order.total)}</p>
                    </div>
                    <div className="flex items-center gap-2">
                      {canCancel && (
                        <button
                          onClick={() => handleCancelOrder(order.id)}
                          disabled={cancellingId === order.id}
                          className="px-4 py-2 text-red-600 border-2 border-red-200 rounded-full hover:bg-red-50 transition-colors text-sm disabled:opacity-50 flex items-center gap-2"
                        >
                          {cancellingId === order.id ? (
                            <>
                              <Loader2 className="w-4 h-4 animate-spin" />
                              Đang hủy...
                            </>
                          ) : (
                            'Hủy đơn'
                          )}
                        </button>
                      )}
                      <button
                        onClick={() => navigate(`/orders/${order.id}`)}
                        className="flex items-center gap-1 px-4 py-2 bg-[var(--color-primary)] text-white rounded-full hover:bg-[var(--color-primary-hover)] transition-colors text-sm"
                      >
                        Chi tiết
                        <ChevronRight className="w-4 h-4" />
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
