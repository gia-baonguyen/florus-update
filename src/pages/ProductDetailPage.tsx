import { useState } from 'react';
import { Star, ShoppingCart, Heart, Share2, Users, Sparkles, Award } from 'lucide-react';
import { ProductCarousel } from '../components/ProductCarousel';
import { ProductImage } from '../components/ProductImage';
import { products, similarProducts, coViewedProducts } from '../data/products';
import type { Product } from '../data/products';

interface ProductDetailPageProps {
  productId: string;
  onProductClick: (productId: string) => void;
  onAddToCart: (productId: string) => void;
}

export function ProductDetailPage({ productId, onProductClick, onAddToCart }: ProductDetailPageProps) {
  const product = products.find(p => p.id === productId);
  const [quantity, setQuantity] = useState(1);
  const [selectedShade, setSelectedShade] = useState(0);

  if (!product) {
    return (
      <div className="min-h-screen bg-[var(--color-surface-warm)] flex items-center justify-center">
        <p>Không tìm thấy sản phẩm</p>
      </div>
    );
  }

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('vi-VN', {
      style: 'currency',
      currency: 'VND',
    }).format(price);
  };

  const discountPercent = product.originalPrice
    ? Math.round(((product.originalPrice - product.price) / product.originalPrice) * 100)
    : 0;

  // Content-based recommendations (similar products)
  const contentBasedProducts = (similarProducts[productId] || [])
    .map(id => products.find(p => p.id === id))
    .filter(Boolean) as Product[];

  // Collaborative filtering (co-viewed products from Neo4j Graph)
  const collaborativeProducts = (coViewedProducts[productId] || [])
    .map(id => products.find(p => p.id === id))
    .filter(Boolean) as Product[];

  const handleAddToCart = () => {
    for (let i = 0; i < quantity; i++) {
      onAddToCart(product.id);
    }
  };

  return (
    <div className="min-h-screen bg-[var(--color-surface-warm)]">
      <div className="container py-6 space-y-6">
        {/* Main Product Section */}
        <div className="bg-white rounded-2xl border border-[var(--color-border)] overflow-hidden shadow-[var(--shadow-md)]">
          <div className="grid md:grid-cols-2 gap-8 p-6 md:p-8">
            {/* Left - Image */}
            <div>
              <div className="aspect-square rounded-xl overflow-hidden bg-gray-50 mb-4">
                <ProductImage
                  imageUrl={product.imageUrl}
                  alt={product.name}
                  size="large"
                />
              </div>
              {/* Thumbnail Gallery */}
              <div className="grid grid-cols-4 gap-3">
                {[1, 2, 3, 4].map((i) => (
                  <div
                    key={i}
                    className="aspect-square rounded-lg overflow-hidden bg-gray-50 border-2 border-transparent hover:border-[var(--color-primary)] cursor-pointer transition-all"
                  >
                    <ProductImage
                      imageUrl={product.imageUrl}
                      alt={`${product.name} view ${i}`}
                      size="small"
                    />
                  </div>
                ))}
              </div>
            </div>

            {/* Right - Product Info */}
            <div className="flex flex-col">
              {/* Brand & Stock */}
              <div className="flex items-center justify-between mb-3">
                <p className="text-gray-500 uppercase tracking-wider text-sm">{product.brand}</p>
                <span className="px-3 py-1.5 bg-green-50 text-green-700 border border-green-200 rounded-full text-xs">
                  ✓ Còn hàng
                </span>
              </div>

              {/* Product Name */}
              <h1 className="mb-4 font-serif">{product.name}</h1>

              {/* Rating */}
              <div className="flex items-center gap-4 mb-5 pb-5 border-b border-gray-100">
                <div className="flex items-center gap-2">
                  <div className="flex">
                    {[...Array(5)].map((_, i) => (
                      <Star
                        key={i}
                        className={`w-5 h-5 ${
                          i < Math.floor(product.rating)
                            ? 'fill-amber-400 text-amber-400'
                            : 'text-gray-300'
                        }`}
                      />
                    ))}
                  </div>
                  <span className="text-gray-900">{product.rating}</span>
                </div>
                <div className="h-4 w-px bg-gray-200" />
                <p className="text-gray-600 text-sm">
                  {product.reviewCount?.toLocaleString() || 0} đánh giá
                </p>
              </div>

              {/* Price */}
              <div className="mb-6">
                <div className="flex items-baseline gap-3 mb-2">
                  <h2 className="text-[var(--color-primary)]">
                    {formatPrice(product.price)}
                  </h2>
                  {product.originalPrice && (
                    <>
                      <span className="text-[var(--color-text-muted)] line-through">
                        {formatPrice(product.originalPrice)}
                      </span>
                      <span className="px-3 py-1 bg-gradient-to-r from-red-500 to-pink-500 text-white rounded-full text-sm shadow-sm">
                        -{discountPercent}%
                      </span>
                    </>
                  )}
                </div>
              </div>

              {/* Why Choose This - AI Insight */}
              <div className="mb-6 p-4 bg-[var(--color-secondary-light)] rounded-xl border border-[var(--color-secondary)]/20">
                <div className="flex items-start gap-2">
                  <Award className="w-5 h-5 text-[var(--color-secondary)] mt-0.5 flex-shrink-0" />
                  <div>
                    <h4 className="text-[var(--color-secondary)] mb-1">Vì sao chọn sản phẩm này?</h4>
                    <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">
                      Sản phẩm này được <strong>{Math.floor(product.rating * 100 / 5)}%</strong> người dùng 
                      đánh giá 5 sao. Phù hợp với <strong>
                      {product.skinType ? product.skinType.join(', ') : 'mọi loại da'}</strong>.
                    </p>
                  </div>
                </div>
              </div>

              {/* Description */}
              <div className="mb-6">
                <h4 className="mb-3">Mô tả sản phẩm</h4>
                <p className="text-[var(--color-text-secondary)] leading-relaxed">
                  {product.description}
                </p>
              </div>

              {/* Shades Selection */}
              {product.shades && product.shades.length > 0 && (
                <div className="mb-6">
                  <h4 className="mb-3">Chọn màu sắc</h4>
                  <div className="flex flex-wrap gap-2">
                    {product.shades.map((shade, index) => (
                      <button
                        key={shade}
                        onClick={() => setSelectedShade(index)}
                        className={`px-4 py-2 rounded-full border-2 text-sm transition-all ${
                          selectedShade === index
                            ? 'border-[var(--color-primary)] bg-[var(--color-primary-light)] text-[var(--color-primary)]'
                            : 'border-[var(--color-border)] hover:border-[var(--color-primary)]'
                        }`}
                      >
                        {shade}
                      </button>
                    ))}
                  </div>
                </div>
              )}

              {/* Tags */}
              <div className="mb-6">
                <div className="flex flex-wrap gap-2">
                  {product.tags.map((tag) => (
                    <span
                      key={tag}
                      className="px-3 py-1.5 bg-[var(--color-surface)] text-[var(--color-text-secondary)] rounded-full text-xs border border-[var(--color-border)]"
                    >
                      #{tag}
                    </span>
                  ))}
                </div>
              </div>

              {/* Quantity Selector */}
              <div className="mb-6">
                <h4 className="mb-3">Số lượng</h4>
                <div className="flex items-center gap-3">
                  <button
                    onClick={() => setQuantity(Math.max(1, quantity - 1))}
                    className="w-10 h-10 rounded-full border-2 border-[var(--color-border)] hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-light)] transition-all"
                  >
                    -
                  </button>
                  <span className="w-12 text-center">{quantity}</span>
                  <button
                    onClick={() => setQuantity(quantity + 1)}
                    className="w-10 h-10 rounded-full border-2 border-[var(--color-border)] hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-light)] transition-all"
                  >
                    +
                  </button>
                </div>
              </div>

              {/* Action Buttons */}
              <div className="flex gap-3 mb-4">
                <button
                  onClick={handleAddToCart}
                  className="flex-1 flex items-center justify-center gap-2 px-6 py-3.5 bg-[var(--color-primary)] text-white rounded-lg hover:opacity-90 transition-all"
                >
                  <ShoppingCart className="w-5 h-5" />
                  Thêm vào giỏ
                </button>
                <button className="flex-1 px-6 py-3.5 bg-gray-900 text-white rounded-lg hover:opacity-90 transition-all">
                  Mua ngay
                </button>
              </div>

              {/* Secondary Actions */}
              <div className="flex gap-3">
                <button className="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 border-2 border-gray-200 rounded-lg hover:bg-gray-50 transition-all">
                  <Heart className="w-4 h-4" />
                  Yêu thích
                </button>
                <button className="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 border-2 border-gray-200 rounded-lg hover:bg-gray-50 transition-all">
                  <Share2 className="w-4 h-4" />
                  Chia sẻ
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Content-based Recommendations */}
        {contentBasedProducts.length > 0 && (
          <div>
            <div className="bg-gradient-to-r from-[var(--color-primary-light)] to-pink-50 rounded-2xl p-5 border border-[var(--color-primary)]/20 mb-4 shadow-sm">
              <div className="flex items-center gap-3 mb-2">
                <div className="p-2 bg-white rounded-xl shadow-sm">
                  <Sparkles className="w-5 h-5 text-[var(--color-primary)]" />
                </div>
                <h3 className="text-[var(--color-primary)] font-serif">Sản phẩm tương tự</h3>
              </div>
              <p className="text-[var(--color-text-secondary)] text-sm leading-relaxed">
                <strong>Content-based filtering</strong> - Dựa trên đặc điểm sản phẩm như thương hiệu, 
                loại hình và khoảng giá để tìm sản phẩm phù hợp với bạn.
              </p>
            </div>

            <ProductCarousel
              title="Bạn có thể thích"
              subtitle="Sản phẩm cùng loại và mục đích sử dụng"
              products={contentBasedProducts}
              onProductClick={onProductClick}
            />
          </div>
        )}

        {/* Collaborative Filtering Recommendations - Neo4j Graph */}
        {collaborativeProducts.length > 0 && (
          <div>
            <div className="bg-gradient-to-r from-purple-50 to-pink-50 rounded-2xl p-5 border border-purple-200 mb-4 shadow-sm">
              <div className="flex items-center gap-3 mb-2">
                <div className="p-2 bg-white rounded-xl shadow-sm">
                  <Users className="w-5 h-5 text-purple-600" />
                </div>
                <h3 className="text-purple-600 font-serif">Hoàn thiện bộ trang điểm</h3>
              </div>
              <p className="text-[var(--color-text-secondary)] text-sm leading-relaxed">
                <strong>Collaborative filtering</strong> - Dựa trên mối quan hệ <code className="px-2 py-0.5 bg-purple-100 rounded text-xs">CO_VIEWED</code> trong 
                Knowledge Graph (Neo4j) để tìm sản phẩm được khách hàng khác quan tâm cùng.
              </p>
            </div>

            <ProductCarousel
              title="Thường được xem cùng"
              subtitle="Sản phẩm được khách hàng khác quan tâm"
              products={collaborativeProducts}
              onProductClick={onProductClick}
              recommendationLabel="Người khác cũng xem"
            />
          </div>
        )}

        {/* Tech Explanation */}
        <div className="bg-white rounded-2xl border border-[var(--color-border)] p-6 shadow-[var(--shadow-sm)]">
          <h3 className="mb-4 font-serif">Công nghệ gợi ý (Knowledge Graph)</h3>
          <div className="grid md:grid-cols-3 gap-4">
            <div className="p-5 bg-[var(--color-surface-warm)] rounded-xl border border-[var(--color-border)]">
              <h4 className="text-[var(--color-text-primary)] mb-2">Content-based</h4>
              <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">
                Phân tích đặc điểm sản phẩm như category, brand, price range để tìm sản phẩm tương tự.
              </p>
            </div>
            <div className="p-5 bg-[var(--color-surface-warm)] rounded-xl border border-[var(--color-border)]">
              <h4 className="text-[var(--color-text-primary)] mb-2">Collaborative</h4>
              <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">
                Sử dụng mối quan hệ CO_VIEWED trong Neo4j Graph Database để tìm pattern xem sản phẩm.
              </p>
            </div>
            <div className="p-5 bg-[var(--color-surface-warm)] rounded-xl border border-[var(--color-border)]">
              <h4 className="text-[var(--color-text-primary)] mb-2">Real-time</h4>
              <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">
                Cập nhật gợi ý theo thời gian thực dựa trên hành vi người dùng mới nhất.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}