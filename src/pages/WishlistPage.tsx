import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Heart, ShoppingCart, Trash2, Loader2 } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { useWishlist } from '../context/WishlistContext';
import { useCart } from '../context/CartContext';

export function WishlistPage() {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();
  const { wishlist, isLoading, removeFromWishlist } = useWishlist();
  const { addItem } = useCart();

  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login');
    }
  }, [isAuthenticated, navigate]);

  const handleAddToCart = async (productId: number) => {
    try {
      await addItem(productId, 1);
    } catch (error) {
      console.error('Failed to add to cart:', error);
    }
  };

  const handleRemove = async (productId: number) => {
    try {
      await removeFromWishlist(productId);
    } catch (error) {
      console.error('Failed to remove from wishlist:', error);
    }
  };

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(price / 1000);
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-[var(--color-surface-warm)] flex items-center justify-center">
        <Loader2 className="w-8 h-8 text-[var(--color-primary)] animate-spin" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[var(--color-surface-warm)]">
      {/* Header */}
      <header className="sticky top-0 z-50 bg-white/95 backdrop-blur-md border-b border-[var(--color-border)]">
        <div className="container">
          <div className="flex items-center h-16 gap-4">
            <button
              onClick={() => navigate(-1)}
              className="p-2 hover:bg-gray-100 rounded-full transition-colors"
            >
              <ArrowLeft className="w-5 h-5" />
            </button>
            <h1 className="text-lg font-semibold flex items-center gap-2">
              <Heart className="w-5 h-5 text-[var(--color-primary)]" fill="currentColor" />
              Wishlist
            </h1>
            {wishlist && wishlist.total_items > 0 && (
              <span className="px-2 py-0.5 bg-[var(--color-primary-light)] text-[var(--color-primary)] rounded-full text-sm">
                {wishlist.total_items} items
              </span>
            )}
          </div>
        </div>
      </header>

      <div className="container py-6">
        {!wishlist || !wishlist.items || wishlist.items.length === 0 ? (
          <div className="text-center py-16">
            <Heart className="w-16 h-16 text-gray-300 mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-gray-600 mb-2">Wishlist is empty</h2>
            <p className="text-gray-500 mb-6">Save products you love for later</p>
            <button
              onClick={() => navigate('/')}
              className="px-6 py-3 bg-[var(--color-primary)] text-white rounded-full hover:bg-[var(--color-primary-hover)] transition-colors"
            >
              Browse Products
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {wishlist.items.map((item) => (
              <div
                key={item.id}
                className="bg-white rounded-2xl border border-[var(--color-border)] overflow-hidden hover:shadow-lg transition-shadow"
              >
                {/* Product Image */}
                <div className="relative aspect-square">
                  <img
                    src={item.product.image_url || '/placeholder.png'}
                    alt={item.product.name}
                    className="w-full h-full object-cover"
                  />
                  <button
                    onClick={() => handleRemove(item.product_id)}
                    className="absolute top-3 right-3 p-2 bg-white/90 rounded-full shadow-md hover:bg-red-50 transition-colors"
                    title="Remove from wishlist"
                  >
                    <Trash2 className="w-4 h-4 text-red-500" />
                  </button>
                  {item.product.original_price && item.product.original_price > item.product.price && (
                    <span className="absolute top-3 left-3 px-2 py-1 bg-red-500 text-white text-xs rounded-full">
                      -{Math.round((1 - item.product.price / item.product.original_price) * 100)}%
                    </span>
                  )}
                </div>

                {/* Product Info */}
                <div className="p-4">
                  <p className="text-xs text-gray-500 mb-1">{item.product.brand}</p>
                  <h3 className="font-medium text-gray-800 mb-2 line-clamp-2">
                    {item.product.name}
                  </h3>
                  <div className="flex items-center gap-2 mb-3">
                    <span className="text-lg font-bold text-[var(--color-primary)]">
                      {formatPrice(item.product.price)}
                    </span>
                    {item.product.original_price && item.product.original_price > item.product.price && (
                      <span className="text-sm text-gray-400 line-through">
                        {formatPrice(item.product.original_price)}
                      </span>
                    )}
                  </div>

                  {/* Rating */}
                  {item.product.rating > 0 && (
                    <div className="flex items-center gap-1 mb-3">
                      <span className="text-yellow-400">★</span>
                      <span className="text-sm text-gray-600">{item.product.rating.toFixed(1)}</span>
                      <span className="text-sm text-gray-400">({item.product.review_count})</span>
                    </div>
                  )}

                  {/* Add to Cart Button */}
                  <button
                    onClick={() => handleAddToCart(item.product_id)}
                    disabled={!item.product.in_stock}
                    className="w-full py-2.5 bg-[var(--color-primary)] text-white rounded-full hover:bg-[var(--color-primary-hover)] transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                  >
                    <ShoppingCart className="w-4 h-4" />
                    {item.product.in_stock ? 'Add to Cart' : 'Out of Stock'}
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
