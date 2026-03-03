import { Star, Sparkles, Heart } from 'lucide-react';
import { Product } from '../types';
import { ProductImage } from './ProductImage';
import { useWishlist } from '../context/WishlistContext';
import { useAuth } from '../context/AuthContext';
import { useNavigate } from 'react-router-dom';

interface ProductCardProps {
  product: Product;
  onClick?: () => void;
  compact?: boolean;
}

export function ProductCard({ product, onClick, compact = false }: ProductCardProps) {
  const { isInWishlist, toggleWishlist } = useWishlist();
  const { isAuthenticated } = useAuth();
  const navigate = useNavigate();

  const handleWishlistClick = async (e: React.MouseEvent) => {
    e.stopPropagation();
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }
    try {
      await toggleWishlist(product.id);
    } catch (error) {
      console.error('Failed to toggle wishlist:', error);
    }
  };

  const isWishlisted = isInWishlist(product.id);
  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(price / 1000);
  };

  const discountPercent = product.original_price
    ? Math.round(((product.original_price - product.price) / product.original_price) * 100)
    : (product.discount || 0);

  return (
    <div
      onClick={onClick}
      className={`bg-white rounded-xl overflow-hidden border border-gray-100 hover:border-gray-200 hover:shadow-md transition-all cursor-pointer group ${compact ? 'h-full' : ''}`}
    >
      {/* Product Image */}
      <div className="relative">
        <ProductImage imageUrl={product.image_url} alt={product.name} size={compact ? 'small' : 'medium'} />

        {/* AI Badge */}
        {product.ai_score && (
          <div className="absolute top-3 left-3 px-2.5 py-1 bg-[var(--color-secondary)] text-white rounded-md flex items-center gap-1.5 text-xs">
            <Sparkles className="w-3.5 h-3.5" />
            AI {product.ai_score}
          </div>
        )}

        {/* Wishlist Heart Button */}
        <button
          onClick={handleWishlistClick}
          className={`absolute top-3 right-3 p-2 rounded-full shadow-md transition-all ${
            isWishlisted
              ? 'bg-red-50 text-red-500'
              : 'bg-white/90 text-gray-400 hover:text-red-500 hover:bg-red-50'
          }`}
          title={isWishlisted ? 'Remove from wishlist' : 'Add to wishlist'}
        >
          <Heart className={`w-4 h-4 ${isWishlisted ? 'fill-current' : ''}`} />
        </button>

        {/* Discount Badge */}
        {discountPercent > 0 && (
          <div className="absolute bottom-3 right-3 px-2.5 py-1 bg-[var(--color-primary)] text-white rounded-md text-xs">
            -{discountPercent}%
          </div>
        )}
      </div>

      {/* Product Info */}
      <div className={compact ? 'p-3' : 'p-4'}>
        <div className="text-xs text-gray-500 mb-1">{product.brand}</div>
        <h3 className={`text-gray-900 mb-2 line-clamp-2 group-hover:text-[var(--color-primary)] transition-colors ${compact ? 'text-sm min-h-[2.5rem]' : 'min-h-[3rem]'}`}>
          {product.name}
        </h3>

        {/* Rating */}
        <div className="flex items-center gap-2 mb-3">
          <div className="flex items-center gap-1">
            <Star className="w-4 h-4 fill-yellow-400 text-yellow-400" />
            <span className="text-sm text-gray-700">{product.rating.toFixed(1)}</span>
          </div>
          <span className="text-xs text-gray-400">({product.review_count})</span>
        </div>

        {/* Price */}
        <div className="flex items-baseline gap-2 flex-wrap">
          <span className={`text-[var(--color-primary)] font-semibold ${compact ? 'text-sm' : ''}`}>
            {formatPrice(product.price)}
          </span>
          {product.original_price && product.original_price > product.price && (
            <span className="text-xs text-gray-400 line-through">
              {formatPrice(product.original_price)}
            </span>
          )}
        </div>

        {/* Tags */}
        {!compact && product.tags && product.tags.length > 0 && (
          <div className="flex flex-wrap gap-1.5 mt-3">
            {product.tags.slice(0, 2).map((tag, index) => (
              <span
                key={index}
                className="px-2 py-0.5 bg-gray-50 text-gray-600 rounded text-xs border border-gray-100"
              >
                {tag}
              </span>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}