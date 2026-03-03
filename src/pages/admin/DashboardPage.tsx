import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import {
  ShoppingCart,
  Users,
  DollarSign,
  Clock,
  TrendingUp,
  Package,
  ArrowRight,
  Loader2
} from 'lucide-react';
import { adminApi } from '../../api/admin';
import { DashboardStats } from '../../types';
import { ProductImage } from '../../components/ProductImage';

export function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadStats();
  }, []);

  const loadStats = async () => {
    try {
      setLoading(true);
      const data = await adminApi.getStats();
      setStats(data);
      setError(null);
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to load dashboard stats');
    } finally {
      setLoading(false);
    }
  };

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(price / 1000);
  };

  const getStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      pending: 'bg-yellow-100 text-yellow-800',
      confirmed: 'bg-blue-100 text-blue-800',
      processing: 'bg-purple-100 text-purple-800',
      shipping: 'bg-indigo-100 text-indigo-800',
      delivered: 'bg-green-100 text-green-800',
      cancelled: 'bg-red-100 text-red-800',
    };
    return colors[status] || 'bg-gray-100 text-gray-800';
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 text-[var(--color-primary)] animate-spin" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-xl p-6 text-center">
        <p className="text-red-600 mb-4">{error}</p>
        <button
          onClick={loadStats}
          className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
        >
          Retry
        </button>
      </div>
    );
  }

  if (!stats) return null;

  const statCards = [
    {
      label: 'Total Revenue',
      value: formatPrice(stats.total_revenue),
      icon: DollarSign,
      color: 'bg-green-500',
      bgColor: 'bg-green-50',
    },
    {
      label: 'Total Orders',
      value: stats.total_orders,
      icon: ShoppingCart,
      color: 'bg-blue-500',
      bgColor: 'bg-blue-50',
    },
    {
      label: 'Total Users',
      value: stats.total_users,
      icon: Users,
      color: 'bg-purple-500',
      bgColor: 'bg-purple-50',
    },
    {
      label: 'Pending Orders',
      value: stats.pending_orders,
      icon: Clock,
      color: 'bg-yellow-500',
      bgColor: 'bg-yellow-50',
    },
  ];

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-800">Dashboard</h1>
        <p className="text-gray-500">Overview of your store performance</p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {statCards.map((stat, index) => (
          <div
            key={index}
            className={`${stat.bgColor} rounded-xl p-6 border border-white/50`}
          >
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600 mb-1">{stat.label}</p>
                <p className="text-2xl font-bold text-gray-800">{stat.value}</p>
              </div>
              <div className={`p-3 rounded-xl ${stat.color}`}>
                <stat.icon className="w-6 h-6 text-white" />
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Additional Stats Row */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* New Users This Month */}
        <div className="bg-gradient-to-r from-pink-50 to-rose-50 rounded-xl p-6 border border-pink-100">
          <div className="flex items-center gap-4">
            <div className="p-3 bg-pink-500 rounded-xl">
              <TrendingUp className="w-6 h-6 text-white" />
            </div>
            <div>
              <p className="text-sm text-gray-600">New Users This Month</p>
              <p className="text-3xl font-bold text-gray-800">{stats.new_users_this_month}</p>
            </div>
          </div>
        </div>

        {/* Orders by Status */}
        <div className="bg-white rounded-xl p-6 border border-gray-200">
          <h3 className="text-sm font-medium text-gray-600 mb-4">Orders by Status</h3>
          <div className="flex flex-wrap gap-2">
            {stats.orders_by_status.map((item) => (
              <span
                key={item.status}
                className={`px-3 py-1.5 rounded-full text-xs font-medium ${getStatusColor(item.status)}`}
              >
                {item.status}: {item.count}
              </span>
            ))}
          </div>
        </div>
      </div>

      {/* Two Column Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Orders */}
        <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
          <div className="p-4 border-b border-gray-200 flex items-center justify-between">
            <h3 className="font-semibold text-gray-800">Recent Orders</h3>
            <Link
              to="/admin/orders"
              className="text-sm text-[var(--color-primary)] hover:underline flex items-center gap-1"
            >
              View All <ArrowRight className="w-4 h-4" />
            </Link>
          </div>
          <div className="divide-y divide-gray-100">
            {stats.recent_orders.length === 0 ? (
              <div className="p-6 text-center text-gray-500">No orders yet</div>
            ) : (
              stats.recent_orders.slice(0, 5).map((order) => (
                <div key={order.id} className="p-4 hover:bg-gray-50 transition-colors">
                  <div className="flex items-center justify-between mb-2">
                    <span className="font-medium text-gray-800">#{order.order_code}</span>
                    <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(order.status)}`}>
                      {order.status}
                    </span>
                  </div>
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-gray-500">{order.user?.name || 'Unknown'}</span>
                    <span className="font-medium text-[var(--color-primary)]">
                      {formatPrice(order.total)}
                    </span>
                  </div>
                  <p className="text-xs text-gray-400 mt-1">{order.created_at}</p>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Top Products */}
        <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
          <div className="p-4 border-b border-gray-200 flex items-center justify-between">
            <h3 className="font-semibold text-gray-800">Top Products</h3>
            <Link
              to="/admin/products"
              className="text-sm text-[var(--color-primary)] hover:underline flex items-center gap-1"
            >
              View All <ArrowRight className="w-4 h-4" />
            </Link>
          </div>
          <div className="divide-y divide-gray-100">
            {stats.top_products.length === 0 ? (
              <div className="p-6 text-center text-gray-500">No products sold yet</div>
            ) : (
              stats.top_products.map((product, index) => (
                <div key={product.id} className="p-4 hover:bg-gray-50 transition-colors">
                  <div className="flex items-center gap-4">
                    <div className="flex items-center justify-center w-8 h-8 rounded-full bg-gray-100 text-gray-600 font-bold text-sm">
                      {index + 1}
                    </div>
                    <div className="w-12 h-12 rounded-lg overflow-hidden bg-gray-100 flex-shrink-0">
                      <ProductImage
                        imageUrl={product.image_url}
                        alt={product.name}
                        size="small"
                      />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="font-medium text-gray-800 truncate">{product.name}</p>
                      <p className="text-sm text-gray-500">
                        <Package className="w-3 h-3 inline mr-1" />
                        {product.total_sold} sold
                      </p>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
