import { Star, Sparkles } from 'lucide-react';
import { Product } from '../data/products';
import { ProductImage } from './ProductImage';

interface ProductCardProps {
  product: Product;
  onClick?: () => void;
}

export function ProductCard({ product, onClick }: ProductCardProps) {
  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('vi-VN', {
      style: 'currency',
      currency: 'VND',
    }).format(price);
  };

  const discountPercent = product.originalPrice
    ? Math.round(((product.originalPrice - product.price) / product.originalPrice) * 100)
    : 0;

  return (
    <div
      onClick={onClick}
      className="bg-white rounded-xl overflow-hidden border border-gray-100 hover:border-gray-200 hover:shadow-md transition-all cursor-pointer group"
    >
      {/* Product Image */}
      <div className="relative">
        <ProductImage imageUrl={product.imageUrl} alt={product.name} size="medium" />
        
        {/* AI Badge */}
        <div className="absolute top-3 left-3 px-2.5 py-1 bg-[var(--color-secondary)] text-white rounded-md flex items-center gap-1.5 text-xs">
          <Sparkles className="w-3.5 h-3.5" />
          AI {product.aiScore}
        </div>

        {/* Discount Badge */}
        {discountPercent > 0 && (
          <div className="absolute top-3 right-3 px-2.5 py-1 bg-[var(--color-primary)] text-white rounded-md text-xs">
            -{discountPercent}%
          </div>
        )}
      </div>

      {/* Product Info */}
      <div className="p-4">
        <div className="text-xs text-gray-500 mb-1">{product.brand}</div>
        <h3 className="text-gray-900 mb-2 line-clamp-2 min-h-[3rem] group-hover:text-[var(--color-primary)] transition-colors">
          {product.name}
        </h3>

        {/* Rating */}
        <div className="flex items-center gap-2 mb-3">
          <div className="flex items-center gap-1">
            <Star className="w-4 h-4 fill-yellow-400 text-yellow-400" />
            <span className="text-sm text-gray-700">{product.rating}</span>
          </div>
          <span className="text-xs text-gray-400">({product.reviewCount})</span>
        </div>

        {/* Price */}
        <div className="flex items-baseline gap-2">
          <span className="text-[var(--color-primary)] font-semibold">
            {formatPrice(product.price)}
          </span>
          {product.originalPrice && (
            <span className="text-sm text-gray-400 line-through">
              {formatPrice(product.originalPrice)}
            </span>
          )}
        </div>

        {/* Tags */}
        {product.tags && product.tags.length > 0 && (
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