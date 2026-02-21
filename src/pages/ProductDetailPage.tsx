import { useState, useEffect } from 'react';
import { Star, ShoppingCart, Heart, Share2, Users, Sparkles, Award, Loader2 } from 'lucide-react';
import { ProductCarousel } from '../components/ProductCarousel';
import { ProductImageGallery } from '../components/ProductImageGallery';
import { ReviewSection } from '../components/ReviewSection';
import { productsApi, recommendationsApi } from '../api/products';
import { useWishlist } from '../context/WishlistContext';
import { useAuth } from '../context/AuthContext';
import { useNavigate } from 'react-router-dom';
import type { Product } from '../types';

interface ProductDetailPageProps {
  productId: string;
  onProductClick: (productId: string) => void;
  onAddToCart: (productId: string) => void;
}

export function ProductDetailPage({ productId, onProductClick, onAddToCart }: ProductDetailPageProps) {
  const [product, setProduct] = useState<Product | null>(null);
  const [similarProducts, setSimilarProducts] = useState<Product[]>([]);
  const [coViewedProducts, setCoViewedProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [quantity, setQuantity] = useState(1);

  const { isInWishlist, toggleWishlist } = useWishlist();
  const { isAuthenticated } = useAuth();
  const navigate = useNavigate();

  const handleWishlistClick = async () => {
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }
    if (product) {
      try {
        await toggleWishlist(product.id);
      } catch (error) {
        console.error('Failed to toggle wishlist:', error);
      }
    }
  };

  const isWishlisted = product ? isInWishlist(product.id) : false;

  useEffect(() => {
    const fetchProduct = async () => {
      try {
        setLoading(true);
        setError(null);

        const productData = await productsApi.getById(parseInt(productId));
        setProduct(productData);

        // Fetch recommendations in parallel
        const [similar, coViewed] = await Promise.all([
          recommendationsApi.getSimilar(parseInt(productId)).catch(() => []),
          recommendationsApi.getCoViewed(parseInt(productId)).catch(() => []),
        ]);

        setSimilarProducts(similar);
        setCoViewedProducts(coViewed);
      } catch (err) {
        console.error('Error fetching product:', err);
        setError('Không thể tải thông tin sản phẩm.');
      } finally {
        setLoading(false);
      }
    };

    if (productId) {
      fetchProduct();
    }
  }, [productId]);

  if (loading) {
    return (
      <div className="min-h-screen bg-[var(--color-surface-warm)] flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="w-10 h-10 text-[var(--color-primary)] animate-spin mx-auto mb-4" />
          <p className="text-gray-500">Đang tải sản phẩm...</p>
        </div>
      </div>
    );
  }

  if (error || !product) {
    return (
      <div className="min-h-screen bg-[var(--color-surface-warm)] flex items-center justify-center">
        <div className="text-center p-8 bg-white rounded-2xl shadow-lg max-w-md">
          <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <span className="text-2xl">😢</span>
          </div>
          <h3 className="text-lg font-semibold text-gray-900 mb-2">Không tìm thấy sản phẩm</h3>
          <p className="text-gray-500 mb-4">{error || 'Sản phẩm này không tồn tại hoặc đã bị xóa.'}</p>
          <button
            onClick={() => window.history.back()}
            className="px-6 py-2 bg-[var(--color-primary)] text-white rounded-lg hover:bg-[var(--color-primary-dark)] transition-colors"
          >
            Quay lại
          </button>
        </div>
      </div>
    );
  }

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('vi-VN', {
      style: 'currency',
      currency: 'VND',
    }).format(price);
  };

  const discountPercent = product.original_price
    ? Math.round(((product.original_price - product.price) / product.original_price) * 100)
    : (product.discount || 0);

  const handleAddToCart = () => {
    for (let i = 0; i < quantity; i++) {
      onAddToCart(product.id.toString());
    }
  };

  return (
    <div className="min-h-screen bg-[var(--color-surface-warm)]">
      <div className="container py-6 space-y-6">
        {/* Main Product Section */}
        <div className="bg-white rounded-2xl border border-[var(--color-border)] overflow-hidden shadow-[var(--shadow-md)]">
          <div className="grid md:grid-cols-2 gap-8 p-6 md:p-8">
            {/* Left - Image Gallery */}
            <div>
              <ProductImageGallery
                images={product.images}
                mainImage={product.image_url}
                productName={product.name}
              />
            </div>

            {/* Right - Product Info */}
            <div className="flex flex-col">
              {/* Brand & Stock */}
              <div className="flex items-center justify-between mb-3">
                <p className="text-gray-500 uppercase tracking-wider text-sm">{product.brand}</p>
                <span className={`px-3 py-1.5 rounded-full text-xs ${
                  product.in_stock
                    ? 'bg-green-50 text-green-700 border border-green-200'
                    : 'bg-red-50 text-red-700 border border-red-200'
                }`}>
                  {product.in_stock ? '✓ Còn hàng' : '✕ Hết hàng'}
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
                  <span className="text-gray-900">{product.rating.toFixed(1)}</span>
                </div>
                <div className="h-4 w-px bg-gray-200" />
                <p className="text-gray-600 text-sm">
                  {product.review_count?.toLocaleString() || 0} đánh giá
                </p>
              </div>

              {/* Price */}
              <div className="mb-6">
                <div className="flex items-baseline gap-3 mb-2">
                  <h2 className="text-[var(--color-primary)]">
                    {formatPrice(product.price)}
                  </h2>
                  {product.original_price && product.original_price > product.price && (
                    <>
                      <span className="text-[var(--color-text-muted)] line-through">
                        {formatPrice(product.original_price)}
                      </span>
                      <span className="px-3 py-1 bg-gradient-to-r from-red-500 to-pink-500 text-white rounded-full text-sm shadow-sm">
                        -{discountPercent}%
                      </span>
                    </>
                  )}
                </div>
              </div>

              {/* Why Choose This - AI Insight */}
              {product.ai_score && (
                <div className="mb-6 p-4 bg-[var(--color-secondary-light)] rounded-xl border border-[var(--color-secondary)]/20">
                  <div className="flex items-start gap-2">
                    <Award className="w-5 h-5 text-[var(--color-secondary)] mt-0.5 flex-shrink-0" />
                    <div>
                      <h4 className="text-[var(--color-secondary)] mb-1">Vì sao chọn sản phẩm này?</h4>
                      <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">
                        Sản phẩm có điểm AI <strong>{product.ai_score}</strong>/100 - được đánh giá cao bởi hệ thống gợi ý thông minh.
                        {product.ai_recommend_type && (
                          <> Loại gợi ý: <strong>{product.ai_recommend_type}</strong>.</>
                        )}
                      </p>
                    </div>
                  </div>
                </div>
              )}

              {/* Description */}
              <div className="mb-6">
                <h4 className="mb-3">Mô tả sản phẩm</h4>
                <p className="text-[var(--color-text-secondary)] leading-relaxed">
                  {product.description}
                </p>
              </div>

              {/* Tags */}
              {product.tags && product.tags.length > 0 && (
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
              )}

              {/* Ingredients */}
              {product.ingredients && product.ingredients.length > 0 && (
                <div className="mb-6">
                  <h4 className="mb-3">Thành phần chính</h4>
                  <div className="flex flex-wrap gap-2">
                    {product.ingredients.slice(0, 5).map((ingredient) => (
                      <span
                        key={ingredient}
                        className="px-3 py-1.5 bg-green-50 text-green-700 rounded-full text-xs border border-green-200"
                      >
                        {ingredient}
                      </span>
                    ))}
                  </div>
                </div>
              )}

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
                    onClick={() => setQuantity(Math.min(product.stock, quantity + 1))}
                    className="w-10 h-10 rounded-full border-2 border-[var(--color-border)] hover:border-[var(--color-primary)] hover:bg-[var(--color-primary-light)] transition-all"
                  >
                    +
                  </button>
                  <span className="text-sm text-gray-500">({product.stock} sản phẩm có sẵn)</span>
                </div>
              </div>

              {/* Action Buttons */}
              <div className="flex gap-3 mb-4">
                <button
                  onClick={handleAddToCart}
                  disabled={!product.in_stock}
                  className="flex-1 flex items-center justify-center gap-2 px-6 py-3.5 bg-[var(--color-primary)] text-white rounded-lg hover:opacity-90 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <ShoppingCart className="w-5 h-5" />
                  Thêm vào giỏ
                </button>
                <button
                  disabled={!product.in_stock}
                  className="flex-1 px-6 py-3.5 bg-gray-900 text-white rounded-lg hover:opacity-90 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Mua ngay
                </button>
              </div>

              {/* Secondary Actions */}
              <div className="flex gap-3">
                <button
                  onClick={handleWishlistClick}
                  className={`flex-1 flex items-center justify-center gap-2 px-4 py-2.5 border-2 rounded-lg transition-all ${
                    isWishlisted
                      ? 'border-red-300 bg-red-50 text-red-600'
                      : 'border-gray-200 hover:bg-gray-50'
                  }`}
                >
                  <Heart className={`w-4 h-4 ${isWishlisted ? 'fill-current' : ''}`} />
                  {isWishlisted ? 'Đã yêu thích' : 'Yêu thích'}
                </button>
                <button className="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 border-2 border-gray-200 rounded-lg hover:bg-gray-50 transition-all">
                  <Share2 className="w-4 h-4" />
                  Chia sẻ
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Reviews Section */}
        <div className="bg-white rounded-2xl border border-[var(--color-border)] p-6 shadow-[var(--shadow-sm)]">
          <ReviewSection productId={parseInt(productId)} />
        </div>

        {/* Content-based Recommendations */}
        {similarProducts.length > 0 && (
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
              products={similarProducts}
              onProductClick={(id) => onProductClick(id.toString())}
            />
          </div>
        )}

        {/* Collaborative Filtering Recommendations - Neo4j Graph */}
        {coViewedProducts.length > 0 && (
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
              products={coViewedProducts}
              onProductClick={(id) => onProductClick(id.toString())}
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
